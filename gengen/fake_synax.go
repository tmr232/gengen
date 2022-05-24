//go:build gengen

package gengen

func Yield(value any) {}

type Generator[T any] error

func (g Generator[T]) Next() bool   { return true }
func (g Generator[T]) Value() T     { return *new(T) }
func (g Generator[T]) Error() error { return nil }
