package msgpack

import (
	"fmt"
	"hex-microservice/shortener"

	"github.com/vmihailenco/msgpack"
)

type Redirect struct{}

func (r *Redirect) Decode(input []byte) (*shortener.Redirect, error) {
	redirect := &shortener.Redirect{}
	if err := msgpack.Unmarshal(input, redirect); err != nil {
		return nil, fmt.Errorf("serializer.Redirect.Decode: %w", err)
	}

	return redirect, nil
}

func (r *Redirect) Encode(input *shortener.Redirect) ([]byte, error) {
	raw, err := msgpack.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("serializer.Redirect.Encode: %w", err)
	}

	return raw, nil
}
