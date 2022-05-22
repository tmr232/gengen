//go:build gengen

package sample

import (
	xyz "fmt"
	"github.com/tmr232/gengen/gengen"
)

func fib() gengen.Generator[int] {
	a := 1
	b := 1
	for {
		gengen.Yield(a)
		a, b = b, a+b
	}
}

func Range(stop int) gengen.Generator[int] {
	for i := 0; i < stop; i++ {
		gengen.Yield(i)
	}
	return nil
}

type SomeGenError struct{}

func (s SomeGenError) Error() string {
	xyz.Println("abc")
	return "Ooh! Error!"
}

func Empty() gengen.Generator[string] {
	return nil
}
func EmptyWithError() gengen.Generator[int] {
	return SomeGenError{}
}

func Yield() gengen.Generator[int] {
	gengen.Yield(1)
	return nil
}
