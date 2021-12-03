package json

import (
	"encoding/json"
	"fmt"
	"hex-microservice/shortener"
)

type Redirect struct{}

func (r *Redirect) Decode(input []byte) (*shortener.Redirect, error) {
	redirect := &shortener.Redirect{}
	if err := json.Unmarshal(input, redirect); err != nil {
		return nil, fmt.Errorf("serializer.Redirect.Decode: %w", err)
	}

	return redirect, nil
}

func (r *Redirect) Encode(input *shortener.Redirect) ([]byte, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("serializer.Redirect.Encode: %w", err)
	}

	return raw, nil
}
