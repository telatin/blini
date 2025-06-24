package sketching

// Replaces a map with slice values.
type svmap[K comparable, V any] struct {
	singles map[K]V
	slices  map[K][]V
}

// Creates a new empty svmap.
func newSVMap[K comparable, V any]() svmap[K, V] {
	return svmap[K, V]{
		singles: map[K]V{},
		slices:  map[K][]V{},
	}
}

// Appends v to the values of k.
func (s svmap[K, V]) put(k K, v V) {
	if _, ok := s.singles[k]; ok {
		s.slices[k] = append(s.slices[k], v)
	} else {
		s.singles[k] = v
	}
}

// Returns the slice associated in k.
func (s svmap[K, V]) get(k K) []V {
	if v, ok := s.singles[k]; ok {
		return append([]V{v}, s.slices[k]...)
	}
	return nil
}

// Removes keys with one value.
func (s svmap[K, V]) clearSingles() {
	for k := range s.singles {
		if s.slices[k] == nil {
			delete(s.singles, k)
		}
	}
}
