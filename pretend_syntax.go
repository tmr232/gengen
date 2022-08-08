//go:build gengen

package gengen

// Yield yields a value from a generator.
func Yield(value any) {}

// Generator is a fake generator type, to satisfy Go's type checking in generator-definitions.
// It implements the Iterator interface, but doesn't really work if executed.
// It is only a placeholder.
type Generator[T any] error

func (g Generator[T]) Next() bool   { return true }
func (g Generator[T]) Value() T     { return *new(T) }
func (g Generator[T]) Error() error { return nil }
