package memory

import "time"

type redirect struct {
	Code      string
	Token     string
	URL       string
	CreatedAt time.Time
}
