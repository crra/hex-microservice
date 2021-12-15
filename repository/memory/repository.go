package memory

import (
	"hex-microservice/shortener"
)

type memoryRepository struct {
	memory map[string]shortener.Redirect
}

func New(_ string) (shortener.Repository, error) {
	return &memoryRepository{
		memory: make(map[string]shortener.Redirect),
	}, nil
}

func (r *memoryRepository) Find(code string) (*shortener.Redirect, error) {
	v, ok := r.memory[code]

	if !ok {
		return nil, shortener.ErrRedirectNotFound
	}

	return &v, nil
}

func (r *memoryRepository) Store(redirect *shortener.Redirect) error {
	r.memory[redirect.Code] = *redirect

	return nil
}
