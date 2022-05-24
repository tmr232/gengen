package sample

import (
	"github.com/tmr232/gengen/gengen"
	"reflect"
	"testing"
)

func TestEmpty(t *testing.T) {
	empty := Empty()
	for empty.Next() {
		t.Error("The empty generator should have no values.")
	}
	if empty.Error() != nil {
		t.Error("The empty generator should have no errors.")
	}
}

func TestEmptyWithError(t *testing.T) {
	emptyWithError := EmptyWithError()
	for emptyWithError.Next() {
		t.Error("The empty generator should have no values.")
	}
	if emptyWithError.Error() == nil {
		t.Error("The empty generator should have an error.")
	}
}

func TestYield(t *testing.T) {
	yield := Yield()
	first := true
	for yield.Next() {
		if first {
			if yield.Value() != 1 {
				t.Error("The yielded value should be 1")
			}
			first = false
		} else {
			t.Error("There should only be one value")
		}
	}
	if yield.Error() != nil {
		t.Error("Unexpected error.")
	}
}

func TestRange(t *testing.T) {
	intRange := Range(10)
	var got []int

	for intRange.Next() {
		got = append(got, intRange.Value())
	}
	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Invalid range. Want %v but got %v.", want, got)
	}
}

func TestFibonacci(t *testing.T) {
	fib := Fibonacci()
	var got []int

	want := []int{1, 1, 2, 3, 5, 8, 13}

	for i := 0; i < len(want); i++ {
		if !fib.Next() {
			break
		}
		got = append(got, fib.Value())
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Invalid sequence. Want %v but got %v.", want, got)
	}

}

func ToSlice[T any](gen gengen.Generator[T]) (slice []T) {
	for gen.Next() {
		slice = append(slice, gen.Value())
	}
	// TODO: Handle errors.
	return
}

func TestIf(t *testing.T) {
	t.Run("if true", func(t *testing.T) {
		want := []string{"true"}
		got := ToSlice(If(true))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Want %v but got %v.", want, got)
		}
	})
	t.Run("if false", func(t *testing.T) {
		want := []string{"false"}
		got := ToSlice(If(false))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Want %v but got %v.", want, got)
		}
	})

}
