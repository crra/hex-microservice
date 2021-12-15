package shortener

import (
	"errors"
	"fmt"
	"time"

	validate "gopkg.in/dealancer/validate.v2"

	"github.com/go-logr/logr"
	"github.com/teris-io/shortid"
)

var (
	ErrRedirectNotFound = errors.New("Redirect Not Found")
	ErrRedirectInvalid  = errors.New("Redirect Invalid")
)

type service struct {
	log        logr.Logger
	repository Repository
}

func NewRedirectService(log logr.Logger, redirectRepo Repository) Service {
	return &service{
		log:        log,
		repository: redirectRepo,
	}
}

func (r *service) Find(code string) (*Redirect, error) {
	return r.repository.Find(code)
}

func (r *service) Store(redirect *Redirect) error {
	if err := validate.Validate(redirect); err != nil {
		return fmt.Errorf("service.Redirect: %w", ErrRedirectInvalid)
	}

	redirect.Code = shortid.MustGenerate()
	redirect.CreatedAt = time.Now().UTC().Unix()

	return r.repository.Store(redirect)
}
