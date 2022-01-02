package adder

// TODO: use go generate to generate transformers that copy from the storage->application and application->storage

import (
	"time"
)

// RedirectStorage is the storage view of a redirect for the adder service.
type RedirectStorage struct {
	Code       string
	URL        string
	Token      string
	ClientInfo string
	CreatedAt  time.Time
}

// RedirectCommand is the request for the adder service.
type RedirectCommand struct {
	URL        string
	CustomCode string
	ClientInfo string
}

// RedirectResult is the result for the adder service.
type RedirectResult struct {
	Code  string
	URL   string
	Token string
}
