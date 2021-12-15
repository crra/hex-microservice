package value

import "fmt"

// GetFromMap returns an accociated value by a given key from a map. In case of key errors (e.g. not found)
// a default value is applied. Because the key could be nil (e.g. from marshalling), a transform function
// is applied after the map access and before the value lookup.
func GetFromMap[V any, K comparable](dictionary map[K]V, keyRef *K, fn func(K) K, defaultV V) V {
	if keyRef == nil {
		return defaultV
	}

	key := *keyRef
	if fn != nil {
		key = fn(key)
	}

	v, ok := dictionary[key]

	if !ok {
		return defaultV
	}

	return v
}

func FirstKeyByValue[K, V comparable](m map[K]V, val V) (K, bool) {
	for k, v := range m {
		if v == val {
			return k, true
		}
	}

	// Default
	var d K
	return d, false
}

// FirstByValue similar to sort.Search but no need to use a closure or
// use the index in a closure or return value.
func FirstByValue[V comparable](m []V, val V) (V, bool) {
	for _, v := range m {
		if v == val {
			return v, true
		}
	}

	// Default
	var d V
	return d, false
}

// FirstByString similar to sort.Search, but enforces the type V to be of type
// fmt.Stringer
func FirstByString[V fmt.Stringer](m []V, fn func(string) string, s string) (V, bool) {
	needle := fn(s)

	for _, v := range m {
		val := fn(v.String())

		if val == needle {
			return v, true
		}
	}

	var d V
	return d, false
}

// OrDefault returns the value if not nil or a default value.
func OrDefault[V any](value *V, defaultV V) V {
	if value == nil {
		return defaultV
	}

	return *value
}

func MustStringOrDefault(value *string, defaultV string) string {
	if value == nil || *value == "" {
		return defaultV
	}

	return *value
}

func PointerOf[T any](t T) *T { return &t }
