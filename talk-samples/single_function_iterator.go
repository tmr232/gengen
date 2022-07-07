package main

import "fmt"

type FibonacciIterator struct {
	a     int
	b     int
	value int
}

func NewFibonacci() *FibonacciIterator {
	return &FibonacciIterator{a: 1, b: 1}
}

func (it *FibonacciIterator) Next() bool {
	it.value = it.a
	it.a, it.b = it.b, it.a+it.b
	return true
}

func (it *FibonacciIterator) Value() int {
	return it.value
}

func (it *FibonacciIterator) Err() error {
	return nil
}

type SingleFunctionIterator[T any] struct {
	Advance func() (hasValue bool, value T, err error)
	value   T
	err     error
}

func (it *SingleFunctionIterator[T]) Next() bool {
	hasValue, value, err := it.Advance()
	it.value = value
	it.err = err
	return hasValue
}

func (it *SingleFunctionIterator[T]) Value() T {
	return it.value
}

func (it *SingleFunctionIterator[T]) Err() error {
	return it.err
}

func Fibonacci() SingleFunctionIterator[int] {
	a := 1
	b := 1
	return SingleFunctionIterator[int]{
		Advance: func() (hasValue bool, value int, err error) {
			value = a
			a, b = b, a+b
			return true, value, nil
		},
	}
}

func main() {
	fib1 := NewFibonacci()
	for i := 0; fib1.Next() && i < 10; i++ {
		fmt.Println(fib1.Value())
	}

	fib2 := Fibonacci()
	for i := 0; fib2.Next() && i < 10; i++ {
		fmt.Println(fib2.Value())
	}
}
