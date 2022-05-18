//go:build !gengen

package main

func yield(value any) {}

type Generator[T any] interface {
	Next()
	Value()
	Error()
}
