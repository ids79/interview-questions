package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type MyContext struct {
	context.Context
	done chan struct{}
}

func (c MyContext) Done() <-chan struct{} {
	return c.done
}

var processData = func(v int) int {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	return v * 2
}

func main() {
	in := make(chan int)
	out := make(chan int)

	c, myCancel := context.WithCancel(context.Background())
	defer myCancel()
	cc := MyContext{Context: c, done: make(chan struct{})}

	ctx, cancel := context.WithTimeout(cc, 5*time.Second)
	defer cancel()

	go func() {
		defer close(in)
		for i := range 10 {
			select {
			case <-ctx.Done():
				return
			case in <- i:
			}
		}
	}()

	start := time.Now()
	processParallel(ctx, in, out, 5)

	for v := range out {
		fmt.Println("v =", v)
	}
	fmt.Println("main duration:", time.Since(start))
}

func processParallel(ctx context.Context, in <-chan int, out chan<- int, numWorkers int) {
	wg := sync.WaitGroup{}
	for range numWorkers {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				case val, ok := <-in:
					if !ok {
						return
					}
					res, err := wrapFunc(ctx, processData, val)
					if err != nil {
						return
					}
					select {
					case <-ctx.Done():
						return
					case out <- res:
					}
				}
			}
		})
	}
	go func() {
		wg.Wait()
		close(out)
	}()
}

func wrapFunc(ctx context.Context, f func(v int) int, val int) (int, error) {
	result := make(chan int, 1)
	go func() {
		result <- f(val)
	}()
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case res := <-result:
		return res, nil
	}
}
