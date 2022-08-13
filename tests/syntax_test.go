package tests

import (
	"reflect"
	"testing"
)

func TestUsesGoRoutine(t *testing.T) {
	want := []int{1}
	got := ToSlice(UsesGoRoutine())
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UsesGoRoutine() = %v, want %v", got, want)
	}
}
