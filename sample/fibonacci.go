package sample

import "github.com/tmr232/gengen"

func ManualFibGen() gengen.Generator[int] {
	a := 1
	b := 1
	return &gengen.GeneratorFunction[int]{
		Advance: func() (__hasValue bool, __value int, __err error) {
			__value = a
			__hasValue = true
			a, b = b, a+b
			return
		},
	}
}

func ManualFibGenNoInterface() gengen.GeneratorFunction[int] {
	a := 1
	b := 1
	return gengen.GeneratorFunction[int]{
		Advance: func() (__hasValue bool, __value int, __err error) {
			__value = a
			__hasValue = true
			a, b = b, a+b
			return
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
