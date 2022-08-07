//go:build !gengen

package gengen

func Yield(value any) {}

type Generator[T any] interface {
	Next() bool
	Value() T
	Error() error
}

type Generator2[A, B any] interface {
	Next() bool
	Value() (A, B)
	Error() error
}

type ClosureIterator[T any] struct {
	Advance func(withValue func(value T) bool, withError func(err error) bool, exhausted func() bool) bool
	value   T
	err     error
}

func (it *ClosureIterator[T]) Value() T {
	return it.value
}

func (it *ClosureIterator[T]) Error() error {
	return it.err
}

func (it *ClosureIterator[T]) Next() bool {
	withValue := func(value T) bool {
		it.value = value
		return true
	}
	withError := func(err error) bool {
		it.err = err
		return false
	}
	exhausted := func() bool { return false }
	return it.Advance(withValue, withError, exhausted)
}
