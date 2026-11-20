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
	Stop func() bool
}

func (c *MyContext) Done() <-chan struct{} {
	return c.done
}

func (c *MyContext) Err() error {
	select {
	case <-c.done:
		return context.Canceled
	default:
	}
	return nil
}

func (c *MyContext) AfterFunc(f func()) func() bool {
	c.Stop = func() bool {
		select {
		case <-c.done:
			return false
		default:
			fmt.Println("The child of MyContext is canceled")
			close(c.done)
		}
		f()
		return true
	}
	return c.Stop
}

func processData(v int) int {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	return v * 2
}

func main() {
	in := make(chan int)
	out := make(chan int)

	c, myCancel := context.WithCancel(context.Background())
	defer myCancel()

	cc := MyContext{Context: c, done: make(chan struct{})}

	ctx, cancel := context.WithTimeout(&cc, 5*time.Second)
	defer cancel()

	// cancel()
	// context.Cause(ctx)

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

	cc.Stop()

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
			worcker(ctx, in, out)
		})
	}
	go func() { // паттерн закрытия канала в отдельной горутине после ожидания
		wg.Wait()
		close(out)
	}()
}

func worcker(ctx context.Context, in <-chan int, out chan<- int) {
	for {
		select {
		case <-ctx.Done():
			return
		case val, ok := <-in:
			if !ok {
				return
			}
			res, err := wrapFunc(ctx, processData, val) // паттерн wrap
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
}

func wrapFunc(ctx context.Context, f func(v int) int, val int) (int, error) {
	result := make(chan int, 1) // паттерн прерывания длительной операции
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
