//go:build gengen

package tests

import "github.com/tmr232/gengen"

func simple(out chan int) {
	out <- 1
}

func UsesGoRoutine() gengen.Generator[int] {
	c := make(chan int)
	go simple(c)
	gengen.Yield(<-c)
	close(c)
	return nil
}
