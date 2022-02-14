package memory

import "time"

type redirect struct {
	Code      string
	Active    bool
	Token     string
	URL       string
	CreatedAt time.Time
}
