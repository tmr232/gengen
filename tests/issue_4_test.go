package tests

import (
	"github.com/tmr232/gengen"
	"reflect"
	"testing"
)

func ToSlice[T any](gen gengen.Generator[T]) (slice []T) {
	for gen.Next() {
		slice = append(slice, gen.Value())
	}
	// TODO: Handle errors.
	return
}

func TestIssue4(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{"2", args{2}, []int{0, 1, 2, 1, 0}},
		{"10", args{10}, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice(Issue4(tt.args.n)); !((len(got) == 0 && len(tt.want) == 0) || reflect.DeepEqual(got, tt.want)) {
				t.Errorf("TakeN() = %v, want %v", got, tt.want)
			}
		})
	}
}
