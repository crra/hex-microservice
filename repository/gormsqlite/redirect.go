package gormsqlite

import (
	"time"
)

type redirect struct {
	Code      string `gorm:"primary_key"`
	Active    bool
	Token     string
	URL       string
	CreatedAt time.Time
}
