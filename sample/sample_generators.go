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
	gengen.Yield("")
}
func EmptyWithError() gengen.Generator[int] {
	return SomeGenError{}
	gengen.Yield(0)
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

func IterMapValues(dict map[int]string) gengen.Generator[string] {
	for _, value := range dict {
		gengen.Yield(value)
	}
	return nil
}

func IterMapKeys(dict map[int]string) gengen.Generator[int] {
	for key, _ := range dict {
		gengen.Yield(key)
	}
	return nil
}

func TakeIntFromSlice(slice []int, n int) gengen.Generator[int] {
	for i := 0; i < n && i < len(slice); i++ {
		gengen.Yield(slice[i])
	}
	return nil
}

func TakeIntFromGenerator(source gengen.Generator[int], n int) gengen.Generator[int] {
	for i := 0; i < n; i++ {
		if !source.Next() {
			break
		}
		gengen.Yield(source.Value())
	}
	return nil
}

func SomeIntScan(n int) gengen.Generator[[]int] {
	data := []int{}
	for i := 0; i < n; i++ {
		data = append(data, i)
		gengen.Yield(data)
	}
	return nil
}
