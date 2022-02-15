package url

import "strings"

func join(elems []string, offset int) string {
	switch lenE := len(elems); lenE {
	case 0:
		return ""
	case 1:
		return elems[0]
	default:
		s := make([]string, lenE+offset)

		index := offset
		for _, e := range elems {
			// empty elements from the input will be omitted to avoid double slashes
			if e != "" {
				s[index] = e
				index++
			}
		}

		return strings.Join(s[:index], "/")
	}
}

// Join is like strings.Join, but with variadic arguments.
func Join(elems ...string) string {
	return join(elems, 0)
}

// AbsPath is like string.Join with the leading "/", but with variadic arguments.
func AbsPath(elems ...string) string {
	path := join(elems, 1)
	if path == "" || path[0] != '/' {
		return "/" + path
	}

	return path
}
