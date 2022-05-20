//go:build !gengen

package gengen

func yield(value any) {}

type Generator[T any] interface {
	Next() bool
	Value() T
	Error() error
}

type GeneratorFunction[T any] struct {
	Advance func() (hasValue bool, value T, err error)
	value   T
	err     error
}

func (g *GeneratorFunction[T]) Next() bool {
	hasValue, value, err := g.Advance()
	g.value = value
	g.err = err
	return hasValue
}

func (g *GeneratorFunction[T]) Value() T {
	return g.value
}

func (g *GeneratorFunction[T]) Error() error {
	return g.err
}
