//go:build !gengen

package gengen

// Yield is used in generator-definitions to yield values.
// In normal Go code it does nothing.
func Yield(value any) {}

// Iterator defines an interface for iteration.
// Usage is as follows:
//
//		iter := GetIterator()
//		for iter.Next() {
//			fmt.Println(iter.Value())
//		}
//		if iter.Error() != nil {
//			panic(iter.Error())
//		}
type Iterator[T any] interface {
	// Next advances the iteration state and returns true if there's another value, false on exhaustion.
	Next() bool
	// Value returns the current value of the iterator
	Value() T
	// Error returns the termination error of the iterator. Will return nil if the iterator was exhausted
	// without errors.
	Error() error
}

// Iterator2 defines an iterator that returns 2 values instead of one.
type Iterator2[A, B any] interface {
	Next() bool
	Value() (A, B)
	Error() error
}

// Generator is the type returned from generator functions.
// Generator implements the Iterator interface.
// It is used by code-generation and not intended for manual creation.
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

// MakeGenerator creates a generator with the given advance function.
// Used by code-generation, and should not generally be used manually.
func MakeGenerator[T any](advance func(withValue func(value T) bool, withError func(err error) bool, exhausted func() bool) bool) Generator[T] {
	return Generator[T]{advance: advance}
}
