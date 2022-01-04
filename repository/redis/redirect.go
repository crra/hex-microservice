package redis

import (
	"reflect"
	"time"
)

type redirect struct {
	Code         string `redis:"code"`
	Token        string `redis:"token"`
	URL          string `redis:"url"`
	CreatedAtStr string `redis:"created_at"`
	CreatedAt    time.Time
}

func (r *redirect) marshalHash(mapping map[string]string) map[string]any {
	r.CreatedAtStr = r.CreatedAt.Format(time.RFC3339)

	values := make(map[string]any)

	vs := reflect.ValueOf(r)
	for vs.Kind() == reflect.Ptr {
		vs = vs.Elem()
	}

	for targetName, fieldName := range mapping {
		values[targetName] = vs.FieldByName(fieldName).Interface()
	}

	return values
}
