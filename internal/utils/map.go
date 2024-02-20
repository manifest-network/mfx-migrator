package utils

// GetKeys returns the keys of a map
// TODO: Remove this function when https://github.com/golang/go/issues/61900 is resolved
func GetKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
