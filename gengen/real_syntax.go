//go:build !gengen

package gengen

func yield(value any) {}

type Generator[T any] interface {
	Next() bool
	Value() T
	Error() error
}
