//go:build gengen

package sample

import (
	"github.com/tmr232/gengen/gengen"
)

func Fibonacci() gengen.Generator[int] {
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

func If(flag bool) gengen.Generator[string] {
	if flag {
		gengen.Yield("true")
	} else {
		gengen.Yield("false")
	}
	return nil
}

func IterIntSlice(slice []int) gengen.Generator[int] {
	for _, val := range slice {
		gengen.Yield(val)
	}
	return nil
}
