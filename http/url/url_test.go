package url

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	t.Parallel()

	for _, f := range []struct {
		name     string
		elements []string
		expected string
	}{
		{
			name:     "empty",
			elements: []string{},
			expected: "",
		},
		{
			name:     "single",
			elements: []string{"https://example.com"},
			expected: "https://example.com",
		},
		{
			name:     "one",
			elements: []string{"https://example.com", "example"},
			expected: "https://example.com/example",
		},
		{
			name:     "two",
			elements: []string{"https://example.com", "foo", "bar"},
			expected: "https://example.com/foo/bar",
		},
		{
			name:     "two with empty",
			elements: []string{"https://example.com", "", "example"},
			expected: "https://example.com/example",
		},
	} {
		f := f // pin
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, f.expected, Join(f.elements...))
		})
	}
}

func TestAbsPath(t *testing.T) {
	t.Parallel()

	for _, f := range []struct {
		name     string
		elements []string
		expected string
	}{
		{
			name:     "empty",
			elements: []string{},
			expected: "/",
		},
		{
			name:     "single",
			elements: []string{""},
			expected: "/",
		},
		{
			name:     "one",
			elements: []string{"foo"},
			expected: "/foo",
		},
		{
			name:     "two",
			elements: []string{"foo", "bar"},
			expected: "/foo/bar",
		},
		{
			name:     "two with empty",
			elements: []string{"", "foo"},
			expected: "/foo",
		},
	} {
		f := f // pin
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, f.expected, AbsPath(f.elements...))
		})
	}
}
