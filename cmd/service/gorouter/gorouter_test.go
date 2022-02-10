package gorouter

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const defaultValueNotDefined = ""

func TestMatchRoutes(t *testing.T) {
	for _, f := range []struct {
		name     string
		path     string
		prefix   string
		vars     []string
		expected map[string]string
	}{
		{
			name:     "single value",
			path:     "/foo/",
			prefix:   "/",
			vars:     []string{"id"},
			expected: map[string]string{"id": "foo"},
		},
		{
			name:     "two values",
			path:     "/foo/bar",
			prefix:   "/",
			vars:     []string{"id", "id2"},
			expected: map[string]string{"id": "foo", "id2": "bar"},
		},
	} {
		f := f // pin
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()

			path := withoutPrefix(withoutTrailing(f.path), f.prefix)

			r := match(&http.Request{}, path, f.vars...)
			if assert.NotNil(t, r) {
				for k, v := range f.expected {
					assert.Equal(t, v, paramFunc(r, k))
				}
			}
		})
	}
}

func TestNoMatchRoutes(t *testing.T) {
	for _, f := range []struct {
		name   string
		path   string
		prefix string
		vars   []string
	}{
		{
			name:   "no vars",
			path:   "/foo/",
			prefix: "/",
			vars:   []string{},
		},
		{
			name:   "more values than in the path",
			path:   "/foo/",
			prefix: "/",
			vars:   []string{"id", "id2"},
		},
		{
			name:   "more in the path than values",
			path:   "/foo/bar/",
			prefix: "/",
			vars:   []string{"id"},
		},
	} {
		f := f // pin
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()

			path := withoutPrefix(withoutTrailing(f.path), f.prefix)

			r := match(&http.Request{}, path, f.vars...)
			assert.Nil(t, r)
		})
	}
}

func TestNoMatch(t *testing.T) {
	assert.Equal(t, defaultValueNotDefined, paramFunc(&http.Request{}, "foo"))
}
