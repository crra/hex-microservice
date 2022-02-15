// Package lookup offers a service to lookup redirects.
package lookup

import (
	"errors"

	"github.com/go-logr/logr"
)

// ErrNotFound signals that the desired redirect is not found
var ErrNotFound = errors.New("redirect not found")

// Repository defines the method the service expects from
// a repository implementation.
type Repository interface {
	Lookup(code string) (RedirectStorage, error)
}

// Service describes the method the service offers.
type Service interface {
	// Lookup takes a code to lookup a Redirect.
	// Raises an error if no redirect is associated with that code.
	Lookup(q RedirectQuery) (RedirectResult, error)
}

// service implements the Service interface and holds
// references.
type service struct {
	logger     logr.Logger
	repository Repository
}

// New creates a new lookup service.
func New(l logr.Logger, r Repository) Service {
	return &service{
		logger:     l,
		repository: r,
	}
}

// Lookup resolves a given code to a redirect
func (s *service) Lookup(q RedirectQuery) (RedirectResult, error) {
	var r RedirectResult
	stored, err := s.repository.Lookup(q.Code)
	if err != nil {
		return r, err
	}

	return fromRedirectStorageToRedirectResult(stored), nil
}
