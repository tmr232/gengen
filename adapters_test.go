package gengen

import (
	"reflect"
	"testing"
)

func First[First, Second any](first First, _ Second) First {
	return first
}
func Second[First, Second any](_ First, second Second) Second {
	return second
}

func ToSlice[Index, Value any](gen Iterator2[Index, Value]) (slice []Value) {
	for gen.Next() {
		slice = append(slice, Second(gen.Value()))
	}
	// TODO: Handle errors.
	return
}

func ToMap[K comparable, V any](gen Iterator2[K, V]) map[K]V {
	result := make(map[K]V)
	for gen.Next() {
		key, value := gen.Value()
		result[key] = value
	}
	return result
}

func TestNewSliceAdapter(t *testing.T) {
	tests := []struct {
		name string
		want []int
	}{
		{"nil", nil},
		//{"empty", []int{}}, // Will always fail because ToSlice defaults no a nil slice, not an empty slice.
		{"single", []int{1}},
		{"multiple", []int{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSlice[int, int](NewSliceAdapter(tt.want)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMapAdaptor(t *testing.T) {
	tests := []struct {
		name string
		want map[int]string
	}{
		{"empty", map[int]string{}},
		{"single", map[int]string{1: "a"}},
		{"multiple", map[int]string{1: "a", 2: "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToMap[int, string](NewMapAdapter(tt.want)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
