package main

import (
	"fmt"
	"runtime/metrics"
)

//nil slice - можно добавлять, нельзя читать

func main() {
	append1()
	append2()
	append3()
	append4()
	append5()
	append6()

	sample := []metrics.Sample{
		{Name: "/sched/goroutines:goroutines"},
		{Name: "/sched/latencies:seconds"},
		//	{Name: "/sched/gomaxprocs:threads"},
	}
	metrics.Read(sample)
	for i := range sample {
		fmt.Println(sample[i].Value.Uint64())
	}

}

func append1() {
	a := [...]int{0, 1, 2, 3}
	x := a[:1]          // [0] cap 4
	y := a[2:]          // [2, 3]  cap 2
	x = append(x, y...) // a=[0, 2, 3, 3] x=[0, 2, 3] y=[3, 3]
	x = append(x, y...) // a=[0, 2, 3, 3] x=[0, 2, 3, 3, 3]
	fmt.Println(a, x)
}

func append2() {
	numbers := []*int{}

	for i := 0; i < 5; i++ {
		numbers = append(numbers, &i)
	}

	for _, number := range numbers { // до go 1.22 4 4 4 4 4
		fmt.Printf("%d ", *number)
	}
	fmt.Printf("\n")
}

func append3() {
	var foo []int
	var bar []int

	foo = append(foo, 1)
	foo = append(foo, 2)
	foo = append(foo, 3)
	bar = append(foo, 4) // foo=[1, 2, 3] bar=[1, 2, 3, 4]
	foo = append(foo, 5) // foo=[1, 2, 3, 5] bar=[1, 2, 3, 5]
	foo = append(foo, 6) // foo=[1, 2, 3, 5, 6] bar=[1, 2, 3, 5]

	fmt.Println(foo, bar)
}

func append4() {
	x := []int{}
	x = append(x, 0)
	x = append(x, 1)
	x = append(x, 2)  // [0, 1, 2]
	y := append(x, 3) // [0, 1, 2, 3]
	z := append(x, 4) // [0, 1, 2, 4]

	fmt.Println(x, y, z)
}

func append5() {
	var i []int

	i = append(i, 1)
	i = append(i, 2)
	i = append(i, 3)

	fmt.Println(i, len(i), cap(i))

	i = append(i, 4)
	x := append(i, 5)

	fmt.Println(i, len(i), cap(i))
	fmt.Println(x, len(x), cap(x))
}

func append6() {
	s := []int{1, 2, 3, 4}
	s = append(s[:2], s[3:]...) // [1, 2] + [4]

	fmt.Println(len(s), cap(s)) // 3 4
}
