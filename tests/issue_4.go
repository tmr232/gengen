//go:build gengen

package tests

import (
	"github.com/tmr232/gengen"
)

//go:generate go run github.com/tmr232/gengen/cmd/gengen

func Issue4(stop int) gengen.Generator[int] {
	for i := 0; i < stop; i++ {
		gengen.Yield(i)
	}

	for i := stop; i >= 0; i-- {
		gengen.Yield(i)
	}
	return nil
}
