package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// nil chan - блокировка при попытке записать или читать
// nil chan - panic при закрытии
// закрытый канал - закрытие и запись panic
// закрытый канал - чтение - ziro value

func main() {
	//dethlock2()
	//empty2()
	//fix3()
	//task2()
	//work3()
	predictableTimeWork(3 * time.Second)
}

func dethlock() {
	var wg sync.WaitGroup
	c := make(chan string)

	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func(c chan<- string, i int) {
			defer wg.Done()

			c <- fmt.Sprintf("Goroutine %s", strconv.Itoa(i))
		}(c, i)
	}

	for { // fatal error: all goroutines are asleep - deadlock!
		select {
		case v := <-c:
			fmt.Println(v)
		}
	}

	wg.Wait()
	close(c)
}

func dethlock2() {
	go func() { // зависание.
		for {
		}
	}()

	var wg sync.WaitGroup
	c := make(chan string, 3)

	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func(c chan<- string, i int) { // до go 1.14 при GOMAXPROCS=1 выведется только одно значение
			defer wg.Done()

			c <- fmt.Sprintf("Goroutine %s", strconv.Itoa(i))
		}(c, i)
	}

	for {
		select {
		case v := <-c:
			fmt.Println(v)
		}
	}

	wg.Wait()
	close(c)
}

func dethlock3() {
	var ch chan int

	for i := 0; i < 3; i++ {
		go func(idx int) {
			ch <- (idx + 1) * 2 // блокировка при записи в nil канал gopark(...
		}(i)
	}

	fmt.Println("result:", <-ch)
	time.Sleep(2 * time.Second)
}

func dethlock4() {
	ch := make(chan int)
	ch <- 1 // блокировка на записи
	go func() {
		fmt.Println(<-ch)
	}()
}

func empty() {
	ch := make(chan string)

	go func() {
		for m := range ch {
			fmt.Println("processed:", m) // может не вывестись второе значение
		}
	}()

	ch <- "cmd.1"
	ch <- "cmd.2"
}

func empty2() {
	for i := 0; i < 10; i++ { // ничего не выведется
		go func() {
			fmt.Println(i)
		}()
	}
}

func fix() {
	ch := make(chan int) // добавить буффер или wait + close в горутине
	wg := &sync.WaitGroup{}

	wg.Add(3)

	for i := 0; i < 3; i++ {
		go func(v int) {
			defer wg.Done()
			ch <- v * v
		}(i)
	}

	go func() {
		wg.Wait()
		close(ch) // добавить close
	}()

	var sum int
	for v := range ch {
		sum += v
	}

	fmt.Printf("result: %d\n", sum)
}

func fix2() {
	a := 5
	wg := sync.WaitGroup{}

	for i := 0; i < a; i++ {
		wg.Go(func() {
			fmt.Println(i)
		})
	}

	wg.Wait()
}

func fix3() {
	ch := make(chan bool, 1) // нужна емкость канала 1
	ch <- true
	go func() {
		<-ch
	}()
	ch <- true
}

func task() {
	runtime.GOMAXPROCS(1)

	now := time.Now()

	ch := 0

	go func() {
		ch = 1
	}()

	for ch == 0 { // 10ms время на прерывание горутины
	}

	fmt.Println("finish", time.Since(now))
}

func task2() {
	runtime.GOMAXPROCS(1)

	for i := 0; i < 5; i++ { // вывод 4 0 1 2 3 лифо буфер локальной очереди в 1 элемент
		go func() {
			time.Sleep(1 * time.Microsecond)
			fmt.Println(i)
		}()
	}

	time.Sleep(time.Second)
}

func work() {
	results := make(chan string, 3)
	stop := make(chan struct{})
	done := make(chan struct{})
	var wg sync.WaitGroup

	for range 3 {
		wg.Go(func() {
			Worker(results, stop)
		})
	}

	go func() {
		for result := range results {
			fmt.Println(result)
		}

		fmt.Println("!! results channel closed, done")

		close(done)
	}()

	<-time.After(3100 * time.Millisecond)

	close(stop)
	wg.Wait()
	close(results)
	<-done

}

func Worker(results chan<- string, stop chan struct{}) {
	for range time.Tick(time.Second) {
		select {
		case <-stop:
			fmt.Println("worker stopped")
			return
		default:
		}

		results <- time.Now().Format(time.TimeOnly)
	}
}

func work2() {
	results := make(chan string, 3)
	stop := make(chan struct{})

	var wgWriter sync.WaitGroup
	for range 3 {
		wgWriter.Go(func() {
			writer(results, stop)
		})
	}

	var wgReader sync.WaitGroup
	for range 3 {
		wgReader.Go(func() {
			reader(results)
		})
	}

	<-time.After(3100 * time.Millisecond)

	close(stop)
	wgWriter.Wait()
	close(results)
	wgReader.Wait()
}

func writer(results chan<- string, stop chan struct{}) {
	for range time.Tick(time.Second) {
		select {
		case <-stop:
			fmt.Println("writer stopped")
			return
		default:
		}

		results <- time.Now().Format(time.TimeOnly)
	}
}

func reader(results <-chan string) {
	for result := range results {
		fmt.Println(result)
	}
	fmt.Println("reader done")
}

func timers() {
	time.NewTimer(time.Second)
	time.After(time.Second)
	time.NewTicker(time.Second)
	time.Tick(time.Second)
}

func work3() {
	reader2(double(writer2()))
}

func writer2() <-chan int {
	ch := make(chan int)
	go func() {
		for i := range 10 {
			ch <- i + 1
		}
		close(ch)
	}()
	return ch
}

func double(c <-chan int) <-chan int {
	ch := make(chan int)
	go func() {
		for v := range c {
			ch <- v * 2
			time.Sleep(500 * time.Millisecond)
		}
		close(ch)
	}()
	return ch
}

func reader2(c <-chan int) {
	for v := range c {
		fmt.Println(v)
	}
}

func randomTimeWork() {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
}

func predictableTimeWork(predict time.Duration) {
	ch := make(chan struct{})
	go func() {
		randomTimeWork()
		close(ch)
	}()
	select {
	case <-time.After(predict):
		fmt.Println("Timeout!")
	case <-ch:
		fmt.Println("work done...")
	}
}
