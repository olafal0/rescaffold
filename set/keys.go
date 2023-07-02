package set

func Keys[T comparable, T2 any](m map[T]T2) []T {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
