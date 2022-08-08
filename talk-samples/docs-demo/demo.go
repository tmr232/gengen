//go:build gengen

// file: demo.go

package main

import (
	"fmt"
	"github.com/tmr232/gengen"
)

//go:generate go run github.com/tmr232/gengen/cmd/gengen

func Range(stop int) gengen.Generator[int] {
	for i := 0; i < stop; i++ {
		gengen.Yield(i)
	}
	return nil
}

func main() {
	numberRange := Range(10)
	for numberRange.Next() {
		fmt.Println(numberRange.Value())
	}
	if numberRange.Error() != nil {
		panic(numberRange.Error())
	}
}
