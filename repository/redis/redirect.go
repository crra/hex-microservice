package redis

import (
	"reflect"
	"time"
)

type redirect struct {
	Code         string `redis:"code"`
	URL          string `redis:"url"`
	CreatedAtStr string `redis:"created_at"`
	CreatedAt    time.Time
}

func (r *redirect) marshalHash(keys []string) map[string]any {
	r.CreatedAtStr = r.CreatedAt.Format(time.RFC3339)

	var values map[string]any
	vs := reflect.ValueOf(r).Elem()

	for _, key := range keys {
		values[key] = vs.FieldByName(key).Interface()
	}

	return values
}
