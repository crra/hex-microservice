package repository

import (
	"hex-microservice/adder"
	"hex-microservice/lookup"
)

// RedirectRepository provides a storage abstraction for service needs.
type RedirectRepository interface {
	// Lookup returns the storage representation of the redirect for the lookup service.
	Lookup(code string) (lookup.RedirectStorage, error)
	// Store persists a redirect from the adder service.
	Store(redirect adder.RedirectStorage) error
	// Delete deletes a stored redirect.
	Invalidate(code, token string) error
}

type Close func() error
