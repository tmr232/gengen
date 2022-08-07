package sample

import "github.com/tmr232/gengen"

func ManualFibGen() gengen.Generator[int] {
	a := 1
	b := 1
	return gengen.Generator[int]{
		Advance: func(withValue func(value int) bool, withError func(err error) bool, exhausted func() bool) bool {
			a, b = b, a+b
			return withValue(a)
		},
	}
}

func ChannelFibonacci(n int) chan int {
	c := make(chan int, 1000)
	go func() {
		x, y := 1, 1
		for i := 0; i < n; i++ {
			c <- x
			x, y = y, x+y
		}
		close(c)
	}()
	return c
}
