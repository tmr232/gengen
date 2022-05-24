package gengen

type SliceAdapter[T any] struct {
	slice []T
	index int
}

func NewSliceAdapter[T any](slice []T) *SliceAdapter[T] {
	return &SliceAdapter[T]{slice: slice, index: -1}
}

func (s *SliceAdapter[T]) Next() bool {
	s.index++
	return s.index < len(s.slice)
}

func (s *SliceAdapter[T]) Value() T {
	return s.slice[s.index]
}

func (s *SliceAdapter[T]) Error() error {
	return nil
}

type Pair[First, Second any] struct {
	first  First
	second Second
}

func NewPair[First, Second any](first First, second Second) *Pair[First, Second] {
	return &Pair[First, Second]{first: first, second: second}
}

type MapAdapter[K comparable, V any] struct {
	items []Pair[K, V]
	index int
}

func (m *MapAdapter[K, V]) Next() bool {
	m.index++
	return m.index < len(m.items)
}

func (m *MapAdapter[K, V]) Value() (K, V) {
	item := m.items[m.index]
	k, v := item.first, item.second
	return k, v
}

func (m *MapAdapter[K, V]) Error() error {
	return nil
}

func NewMapAdaptor[K comparable, V any](map_ map[K]V) *MapAdapter[K, V] {
	items := make([]Pair[K, V], len(map_))
	for key, value := range map_ {
		items = append(items, *NewPair(key, value))
	}
	return &MapAdapter[K, V]{items: items, index: -1}
}
