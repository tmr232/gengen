package examples

import (
	"github.com/tmr232/gengen"
	"reflect"
	"sort"
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

func TestIterIntSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"empty", []int{}, nil},
		{"single", []int{1}, []int{1}},
		{"multiple", []int{5, 3}, []int{5, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice[int](IterIntSlice(tt.input)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestIterMapValues(t *testing.T) {
	tests := []struct {
		name  string
		input map[int]string
		want  []string
	}{
		{"empty", map[int]string{}, nil},
		{"single", map[int]string{1: "a"}, []string{"a"}},
		{"multiple", map[int]string{1: "a", 2: "b"}, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToSlice[string](IterMapValues(tt.input))
			sort.Strings(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIterMapKeys(t *testing.T) {
	tests := []struct {
		name  string
		input map[int]string
		want  []int
	}{
		{"empty", map[int]string{}, nil},
		{"single", map[int]string{1: "a"}, []int{1}},
		{"multiple", map[int]string{1: "a", 2: "b"}, []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToSlice[int](IterMapKeys(tt.input))
			sort.Ints(got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTakeIntFromSlice(t *testing.T) {
	type Args struct {
		slice []int
		n     int
	}
	tests := []struct {
		name string
		args Args
		want []int
	}{
		{"empty", Args{[]int{}, 0}, nil},
		{"1 item", Args{[]int{1, 2, 3, 4}, 1}, []int{1}},
		{"2 items", Args{[]int{1, 2, 3, 4}, 2}, []int{1, 2}},
		{"all items", Args{[]int{1, 2, 3, 4}, 4}, []int{1, 2, 3, 4}},
		{"more than len", Args{[]int{1, 2, 3, 4}, 5}, []int{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice[int](TakeIntFromSlice(tt.args.slice, tt.args.n)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTakeIntFromGenerator(t *testing.T) {
	type Args struct {
		stop int
		n    int
	}
	tests := []struct {
		name string
		args Args
		want []int
	}{
		{"empty", Args{0, 0}, nil},
		{"1 item", Args{4, 1}, []int{0}},
		{"2 items", Args{4, 2}, []int{0, 1}},
		{"all items", Args{4, 4}, []int{0, 1, 2, 3}},
		{"more than len", Args{4, 5}, []int{0, 1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice[int](TakeIntFromGenerator(Range(tt.args.stop), tt.args.n)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSomeIntScan(t *testing.T) {
	type Args struct {
		n int
	}
	tests := []struct {
		name string
		args Args
		want [][]int
	}{
		{"empty", Args{0}, nil},
		{"1 item", Args{1}, [][]int{{0}}},
		{"2 items", Args{2}, [][]int{{0}, {0, 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice[[]int](SomeIntScan(tt.args.n)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTakeN(t *testing.T) {
	type args struct {
		n      int
		source gengen.Generator[int]
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"take 0", args{0, Range(10)}, []int{}},
		{"take 1", args{1, Range(10)}, []int{0}},
		{"take 2", args{2, Range(10)}, []int{0, 1}},
		{"take all", args{10, Range(10)}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"take more", args{20, Range(10)}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice(TakeN(tt.args.n, tt.args.source)); !((len(got) == 0 && len(tt.want) == 0) || reflect.DeepEqual(got, tt.want)) {
				t.Errorf("TakeN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDropN(t *testing.T) {
	type args struct {
		n      int
		source gengen.Generator[int]
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"drop 0", args{0, Range(10)}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"drop 1", args{1, Range(10)}, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"drop 2", args{2, Range(10)}, []int{2, 3, 4, 5, 6, 7, 8, 9}},
		{"drop all", args{10, Range(10)}, []int{}},
		{"drop more", args{20, Range(10)}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice(DropN(tt.args.n, tt.args.source)); !((len(got) == 0 && len(tt.want) == 0) || reflect.DeepEqual(got, tt.want)) {
				t.Errorf("DropN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTakeWhile(t *testing.T) {
	type args struct {
		predicate func(int) bool
		source    gengen.Generator[int]
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"Take none", args{func(int) bool { return false }, Range(10)}, []int{}},
		{"Take all", args{func(int) bool { return true }, Range(10)}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"Take <5", args{func(n int) bool { return n < 5 }, Range(10)}, []int{0, 1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice(TakeWhile(tt.args.predicate, tt.args.source)); !((len(got) == 0 && len(tt.want) == 0) || reflect.DeepEqual(got, tt.want)) {
				t.Errorf("TakeWhile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDropWhile(t *testing.T) {
	type args struct {
		predicate func(int) bool
		source    gengen.Generator[int]
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"Drop none", args{func(int) bool { return false }, Range(10)}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"Drop all", args{func(int) bool { return true }, Range(10)}, []int{}},
		{"Drop <5", args{func(n int) bool { return n < 5 }, Range(10)}, []int{5, 6, 7, 8, 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice(DropWhile(tt.args.predicate, tt.args.source)); !((len(got) == 0 && len(tt.want) == 0) || reflect.DeepEqual(got, tt.want)) {
				t.Errorf("DropWhile() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestFilterIn(t *testing.T) {
	type args struct {
		predicate func(int) bool
		source    gengen.Generator[int]
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"Filter none", args{func(int) bool { return false }, Range(10)}, []int{}},
		{"Filter all", args{func(int) bool { return true }, Range(10)}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"Filter even", args{func(n int) bool { return n%2 == 0 }, Range(10)}, []int{0, 2, 4, 6, 8}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice(FilterIn(tt.args.predicate, tt.args.source)); !((len(got) == 0 && len(tt.want) == 0) || reflect.DeepEqual(got, tt.want)) {
				t.Errorf("FilterIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterOut(t *testing.T) {
	type args struct {
		predicate func(int) bool
		source    gengen.Generator[int]
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"Filter none", args{func(int) bool { return false }, Range(10)}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"Filter all", args{func(int) bool { return true }, Range(10)}, []int{}},
		{"Filter even", args{func(n int) bool { return n%2 == 0 }, Range(10)}, []int{1, 3, 5, 7, 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice(FilterOut(tt.args.predicate, tt.args.source)); !((len(got) == 0 && len(tt.want) == 0) || reflect.DeepEqual(got, tt.want)) {
				t.Errorf("FilterOut() = %v, want %v", got, tt.want)
			}
		})
	}
}
