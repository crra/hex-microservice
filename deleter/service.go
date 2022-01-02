package deleter

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"
)

// ErrNotFound signals that the desired redirect is not found
var (
	ErrNotFound = errors.New("redirect not found")
)

// Repository defines the method the service expects from
// a repository implementation.
type Repository interface {
	DeleteFind(code string) (RedirectStorage, error)
	Delete(code, token string) error
}

// Service describes the method the service offers.
type Service interface {
	// Lookup takes a token to deletes a Redirect.
	// Raises an error if the entry couldn't be deleted.
	Delete(q RedirectQuery) error
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

// Delete deletes a redirect by the given token.
func (s *service) Delete(q RedirectQuery) error {
	stored, err := s.repository.DeleteFind(q.Code)
	if err != nil {
		return err
	}

	if q.Token != stored.Token {
		return fmt.Errorf("invalid token: %w", ErrNotFound)
	}

	if err := s.repository.Delete(q.Code, q.Token); err != nil {
		return err
	}

	return nil
}
