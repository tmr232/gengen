//go:build gengen

package main

func fib() Generator[int] {
	a := 1
	b := 1
	for {
		yield(a)
		a, b = b, a+b
	}
}

func Range(stop int) Generator[int] {
	for i := 0; i < stop; i++ {
		yield(i)
	}
	return nil
}

type SomeGenError struct{}

func (s SomeGenError) Error() string {
	return "Ooh! Error!"
}

func WithError() Generator[int] {
	return SomeGenError{}
}
