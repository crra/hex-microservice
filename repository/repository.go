package repository

import (
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
)

// RedirectRepository provides a storage abstraction for service needs.
type RedirectRepository interface {
	// LookupFind returns the storage representation of the redirect for the lookup service.
	LookupFind(code string) (lookup.RedirectStorage, error)
	// Store persists a redirect from the adder service.
	Store(redirect adder.RedirectStorage) error
	// DeleteFind returns the storage representation of the redirect for the deleter service.
	DeleteFind(code string) (deleter.RedirectStorage, error)
	// Delete deletes a stored redirect.
	Delete(code, token string) error
}
