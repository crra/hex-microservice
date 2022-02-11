package memory

import (
	"context"
	"errors"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
	"hex-microservice/repository"
	"sync"
)

var errNotFound = errors.New("not found")

type memoryRepository struct {
	memory map[string]redirect
	m      sync.RWMutex
}

func New(_ context.Context, _ string) (repository.RedirectRepository, repository.Close, error) {
	return &memoryRepository{
		memory: make(map[string]redirect),
		m:      sync.RWMutex{},
	}, func() error { return nil }, nil
}

// findByCode resolves a redirect by it's code.
func (r *memoryRepository) findByCode(code string) (redirect, error) {
	r.m.RLock()
	red, ok := r.memory[code]
	r.m.RUnlock()

	if !ok {
		return red, errNotFound
	}

	return red, nil
}

func (r *memoryRepository) LookupFind(code string) (lookup.RedirectStorage, error) {
	var red lookup.RedirectStorage

	stored, err := r.findByCode(code)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return red, lookup.ErrNotFound
		}

		return red, err
	}

	return fromRedirectToLookupRedirectStorage(stored), nil
}

func (r *memoryRepository) Store(red adder.RedirectStorage) error {
	// Check if already there
	r.m.RLock()
	_, ok := r.memory[red.Code]
	r.m.RUnlock()
	if ok {
		return adder.ErrDuplicate
	}

	r.m.Lock()
	r.memory[red.Code] = fromAdderRedirectStorageToRedirect(red)
	r.m.Unlock()

	return nil
}

func (r *memoryRepository) Delete(code, token string) error {
	r.m.RLock()
	red, ok := r.memory[code]
	r.m.RUnlock()
	if !ok {
		return deleter.ErrNotFound
	}

	if red.Token != token {
		return fmt.Errorf("invalid token: %w", deleter.ErrNotFound)
	}

	r.m.Lock()
	delete(r.memory, code)
	r.m.Unlock()

	return nil
}

func (r *memoryRepository) DeleteFind(code string) (deleter.RedirectStorage, error) {
	var red deleter.RedirectStorage

	stored, err := r.findByCode(code)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return red, deleter.ErrNotFound
		}

		return red, err
	}

	return fromRedirectToDeleterRedirectStorage(stored), nil
}
