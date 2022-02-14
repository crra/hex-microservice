package value

import (
	"reflect"
	"strings"

	"github.com/fatih/structtag"
)

func Ident[K comparable](key K) K {
	return key
}

func Join(sep string, elements ...string) string {
	return strings.Join(elements, sep)
}

// GetFromMap returns an associated value by a given key from a map. In case of key errors (e.g. not found)
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

func GetFromSlice[V any](s []V, index int, defaultV V) V {
	if index < 0 || len(s) < index {
		return defaultV
	}

	return s[index]
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

func FirstValueFromSlice[E any](elements []E, fn func(E) bool) *E {
	for _, e := range elements {
		if fn(e) {
			return &e
		}
	}

	return nil
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
// XXX: fmt.Stringer does not work
type Stringer interface {
	String() string
}

func FirstByString[V Stringer](m []V, fn func(string) string, s string) (V, bool) {
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

func Must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}

	return v
}

type nameExtractor func(f reflect.StructField) (string, error)

func extractNameFromTag(t string) nameExtractor {
	return func(f reflect.StructField) (string, error) {
		tags, err := structtag.Parse(string(f.Tag))
		if err != nil {
			return "", err
		}

		namedTag, ok := tags.Get(t)
		if ok != nil {
			return "", err
		}

		return namedTag.Name, nil
	}
}

func extractNameFromName(f reflect.StructField) (string, error) {
	return f.Name, nil
}

func Mapping(s any, tag string) (map[string]string, error) {
	t := reflect.TypeOf(s)

	var extract nameExtractor
	if tag == "" {
		extract = extractNameFromName
	} else {
		extract = extractNameFromTag(tag)
	}

	kv := make(map[string]string)

	for _, field := range reflect.VisibleFields(t) {
		renamed, err := extract(field)
		if err != nil {
			return nil, err
		}

		if renamed == "" {
			continue
		}

		kv[renamed] = field.Name
	}

	return kv, nil
}

func PickFlat(o any, keys []string) (map[string]any, error) {
	return nil, nil
}

// Values returns all keys of a map.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))

	index := 0
	for k := range m {
		keys[index] = k
		index++
	}

	return keys
}

// Values returns all values of a map.
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, len(m))

	index := 0
	for k := range m {
		values[index] = m[k]
		index++
	}

	return values
}
