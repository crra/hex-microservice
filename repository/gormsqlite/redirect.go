package gormsqlite

import (
	"time"
)

type redirect struct {
	Code      string `gorm:"primary_key"`
	Token     string
	URL       string
	CreatedAt time.Time
}
