package lookup

import "time"

// RedirectStorage is the storage view for the lookup service.
type RedirectStorage struct {
	Code      string
	URL       string
	CreatedAt time.Time
}

// RedirectQuery is the request query of the lookup service.
type RedirectQuery struct {
	Code string
}

// RedirectResult is the result of the lookup service.
type RedirectResult struct {
	Code      string
	URL       string
	CreatedAt time.Time
}
