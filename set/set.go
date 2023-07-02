package set

type Set[T comparable] map[T]bool

func New[T comparable](elements ...T) Set[T] {
	set := make(Set[T], len(elements))
	set.Add(elements...)
	return set
}

func NewWithCap[T comparable](cap int) Set[T] {
	return make(Set[T], cap)
}

func (s Set[T]) Add(elements ...T) Set[T] {
	for _, element := range elements {
		s[element] = true
	}
	return s
}

func (s Set[T]) Remove(elements ...T) Set[T] {
	for _, element := range elements {
		delete(s, element)
	}
	return s
}

func (s Set[T]) Contains(element T) bool {
	return s[element]
}

// Difference returns a new set containing all elements in s that are not in the
// other set
func (s Set[T]) Difference(other Set[T]) Set[T] {
	result := New[T]()
	for element := range s {
		if !other.Contains(element) {
			result.Add(element)
		}
	}
	return result
}
