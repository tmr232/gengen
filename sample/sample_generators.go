//go:build gengen

package sample

import (
	"errors"
	"github.com/tmr232/gengen"
)

//go:generate go run github.com/tmr232/gengen/cmd/gengen

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
	return errors.New("Generator Error!")
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

func TakeN[T any](n int, source gengen.Generator[T]) gengen.Generator[T] {
	for i := 0; i < n && source.Next(); i++ {
		gengen.Yield(source.Value())
	}
	return source.Error()
}

func DropN[T any](n int, source gengen.Generator[T]) gengen.Generator[T] {
	for i := 0; i < n; i++ {
		if !source.Next() {
			return source.Error()
		}
	}
	for source.Next() {
		gengen.Yield(source.Value())
	}
	return source.Error()
}

func TakeWhile[T any](predicate func(T) bool, source gengen.Generator[T]) gengen.Generator[T] {
	for source.Next() {
		value := source.Value()
		if !predicate(value) {
			return nil
		}
		gengen.Yield(value)
	}
	return source.Error()
}

func DropWhile[T any](predicate func(T) bool, source gengen.Generator[T]) gengen.Generator[T] {
	found := false
	for source.Next() {
		if !predicate(source.Value()) {
			found = true
			break
		}
	}
	if found {
		gengen.Yield(source.Value())
		for source.Next() {
			gengen.Yield(source.Value())
		}
	}
	return source.Error()

}

func FilterIn[T any](predicate func(T) bool, source gengen.Generator[T]) gengen.Generator[T] {
	for source.Next() {
		value := source.Value()
		if !predicate(value) {
			continue
		}
		gengen.Yield(value)
	}
	return source.Error()
}

func FilterOut[T any](predicate func(T) bool, source gengen.Generator[T]) gengen.Generator[T] {
	for source.Next() {
		value := source.Value()
		if predicate(value) {
			continue
		}
		gengen.Yield(value)
	}
	return source.Error()
}
