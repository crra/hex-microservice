// Package adder offers a service to add a redirect.
package adder

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/teris-io/shortid"
	validate "gopkg.in/dealancer/validate.v2"
)

var (
	ErrRedirectInvalid = errors.New("Redirect Invalid")
	ErrDuplicate       = errors.New("Redirect already exists")
)

// Repository defines the methods the service expects from
// a repository implementation.
type Repository interface {
	Store(RedirectStorage) error
}

// Service describes the methods the service offers.
type Service interface {
	// Add takes a list of redirects for persistence.
	// May raise errors if the redirect is not valid.
	Add(...RedirectCommand) ([]RedirectResult, error)
}

// service implements the Service interface and holds
// references.
type service struct {
	logger     logr.Logger
	repository Repository
}

// New creates a new adder service.
func New(l logr.Logger, r Repository) Service {
	return &service{
		logger:     l,
		repository: r,
	}
}

// Add takes a list of redirect commands and persists them.
func (s *service) Add(redirects ...RedirectCommand) ([]RedirectResult, error) {
	results := make([]RedirectResult, len(redirects))

	for i, redirect := range redirects {
		if err := validate.Validate(redirect); err != nil {
			return results, fmt.Errorf("service.Redirect: %w", ErrRedirectInvalid)
		}

		// TODO: custom code
		code, err := shortid.Generate()
		if err != nil {
			return results, err
		}

		// storage view
		store := RedirectStorage{
			Code:       code,
			URL:        redirect.URL,
			Token:      strings.Replace(uuid.New().String(), "-", "", -1),
			ClientInfo: redirect.ClientInfo,
			CreatedAt:  time.Now(),
		}
		if err := s.repository.Store(store); err != nil {
			return results, err
		}

		// result view
		results[i], err = storageToResult(store)
		if err != nil {
			return results, err
		}
	}

	return results, nil
}
