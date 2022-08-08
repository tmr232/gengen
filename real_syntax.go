//go:build !gengen

package gengen

func Yield(value any) {}

type Iterator[T any] interface {
	Next() bool
	Value() T
	Error() error
}

type Generator2[A, B any] interface {
	Next() bool
	Value() (A, B)
	Error() error
}

type Generator[T any] struct {
	advance func(withValue func(value T) bool, withError func(err error) bool, exhausted func() bool) bool
	value   T
	err     error
}

func (it *Generator[T]) Value() T {
	return it.value
}

func (it *Generator[T]) Error() error {
	return it.err
}

func (it *Generator[T]) Next() bool {
	withValue := func(value T) bool {
		it.value = value
		return true
	}
	withError := func(err error) bool {
		it.err = err
		return false
	}
	exhausted := func() bool { return false }
	return it.advance(withValue, withError, exhausted)
}

func MakeGenerator[T any](advance func(withValue func(value T) bool, withError func(err error) bool, exhausted func() bool) bool) Generator[T] {
	return Generator[T]{advance: advance}
}
