package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var errTimeout = errors.New("timed out")

var processData = func(v int) int {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	return v * 2
}

func main() {
	in := make(chan int)
	out := make(chan int)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		defer close(in)
		for i := range 10 {
			select {
			case <-ctx.Done():
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
			select {
			case <-ctx.Done():
			case val, ok := <-in:
				if !ok {
					return
				}
				res, err := wrapFunc(ctx, processData, val)
				if errors.Is(err, errTimeout) {
					return
				}
				select {
				case <-ctx.Done():
				case out <- res:
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
	result := make(chan struct{})
	var res int
	go func() {
		res = f(val)
		close(result)
	}()
	select {
	case <-ctx.Done():
		return 0, errTimeout
	case <-result:
		return res, nil
	}
}
