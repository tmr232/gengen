//go:build gengen

package demo

import (
	"fmt"
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

func main() {
	fib := Fibonacci()
	for i := 0; i < 10 && fib.Next(); i++ {
		fmt.Println(fib.Value())
	}
}
