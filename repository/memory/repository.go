package memory

import (
	"context"
	"errors"
	"hex-microservice/adder"
	"hex-microservice/invalidator"
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
func (r *memoryRepository) findActiveByCode(code string) (redirect, error) {
	r.m.RLock()
	red, ok := r.memory[code]
	r.m.RUnlock()

	if !ok || !red.Active {
		return red, errNotFound
	}

	return red, nil
}

func (r *memoryRepository) findActiveByCodeAndToken(code, token string) (redirect, error) {
	r.m.RLock()
	red, ok := r.memory[code]
	r.m.RUnlock()

	if !ok || !red.Active || red.Token != token {
		return red, errNotFound
	}

	return red, nil
}

func (r *memoryRepository) Lookup(code string) (lookup.RedirectStorage, error) {
	var red lookup.RedirectStorage

	stored, err := r.findActiveByCode(code)
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

	store := fromAdderRedirectStorageToRedirect(red)
	store.Active = true

	r.m.Lock()
	r.memory[red.Code] = store
	r.m.Unlock()

	return nil
}

func (r *memoryRepository) Invalidate(code, token string) error {
	store, err := r.findActiveByCodeAndToken(code, token)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return invalidator.ErrNotFound
		}

		return err
	}

	store.Active = false

	r.m.Lock()
	r.memory[store.Code] = store
	r.m.Unlock()

	return nil
}
