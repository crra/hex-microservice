package repository_test

import (
	"context"
	"hex-microservice/adder"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"hex-microservice/repository"
	"hex-microservice/repository/gormsqlite"
	"hex-microservice/repository/memory"
	"hex-microservice/repository/sqlite"
	"testing"

	"github.com/stretchr/testify/assert"
)

var repositoryImplementations = []struct {
	name   string
	config string
	new    func(context.Context, string) (repository.RedirectRepository, repository.Close, error)
}{
	{"memory", "", memory.New},
	{
		"sqlite",
		"file::memory:?cache=shared&_journal_mode=WAL&_foreign_keys=true",
		sqlite.New,
	},
	{
		"gorm + sqlite",
		//"file::memory:?cache=shared&_journal_mode=WAL&_foreign_keys=true",
		":memory:",
		gormsqlite.New,
	},
}

func TestLookupNonExisting(t *testing.T) {
	ctx := context.Background()

	const code = "code"
	for _, ri := range repositoryImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			repo, close, err := ri.new(ctx, ri.config)
			if assert.NoError(t, err) {
				defer close()

				_, err := repo.Lookup(code)
				assert.ErrorIs(t, err, lookup.ErrNotFound)
			}
		})
	}
}

func TestStoreAndReadBack(t *testing.T) {
	ctx := context.Background()

	const (
		code  = "code"
		token = "token"
		url   = "https://example.com"
	)

	for _, ri := range repositoryImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			repo, close, err := ri.new(ctx, ri.config)
			if assert.NoError(t, err) {
				defer close()

				err = repo.Store(adder.RedirectStorage{
					Code:  code,
					Token: token,
					URL:   url,
				})
				if assert.NoError(t, err) {
					lookedUp, err := repo.Lookup(code)
					if assert.NoError(t, err) {
						assert.Equal(t, code, lookedUp.Code)
					}
				}
			}
		})
	}
}

func TestStoreTwice(t *testing.T) {
	ctx := context.Background()

	const (
		code  = "code"
		token = "token"
		url   = "https://example.com"
	)

	for _, ri := range repositoryImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			repo, close, err := ri.new(ctx, ri.config)
			if assert.NoError(t, err) {
				defer close()

				err = repo.Store(adder.RedirectStorage{
					Code:  code,
					Token: token,
					URL:   url,
				})
				if assert.NoError(t, err) {
					err = repo.Store(adder.RedirectStorage{
						Code:  code,
						Token: token,
						URL:   url,
					})

					assert.ErrorIs(t, err, adder.ErrDuplicate)
				}
			}
		})
	}
}

func TestInvalidateNonExisting(t *testing.T) {
	ctx := context.Background()

	const (
		code  = "code"
		token = "token"
	)

	for _, ri := range repositoryImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			repo, close, err := ri.new(ctx, ri.config)
			if assert.NoError(t, err) {
				defer close()

				err := repo.Invalidate(code, token)
				assert.ErrorIs(t, err, invalidator.ErrNotFound)
			}
		})
	}
}

func TestInvalidateInvalidToken(t *testing.T) {
	ctx := context.Background()

	const (
		code         = "code"
		token        = "token"
		invalidToken = token + token
		url          = "https://example.com"
	)

	for _, ri := range repositoryImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			repo, close, err := ri.new(ctx, ri.config)
			if assert.NoError(t, err) {
				defer close()

				err = repo.Store(adder.RedirectStorage{
					Code:  code,
					Token: token,
					URL:   url,
				})
				if assert.NoError(t, err) {
					err := repo.Invalidate(code, invalidToken)
					assert.ErrorIs(t, err, invalidator.ErrNotFound)
				}
			}
		})
	}
}

func TestInvalidate(t *testing.T) {
	ctx := context.Background()

	const (
		code  = "code"
		token = "token"
		url   = "https://example.com"
	)

	for _, ri := range repositoryImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			repo, close, err := ri.new(ctx, ri.config)
			if assert.NoError(t, err) {
				defer close()

				err = repo.Store(adder.RedirectStorage{
					Code:  code,
					Token: token,
					URL:   url,
				})
				if assert.NoError(t, err) {
					err := repo.Invalidate(code, token)
					if assert.NoError(t, err) {
						_, err := repo.Lookup(code)
						assert.ErrorIs(t, err, lookup.ErrNotFound)
					}
				}
			}
		})
	}
}

func TestInvalidateAndAdd(t *testing.T) {
	ctx := context.Background()

	const (
		code  = "code"
		token = "token"
		url   = "https://example.com"
	)

	for _, ri := range repositoryImplementations {
		ri := ri // pin

		t.Run(ri.name, func(t *testing.T) {
			t.Parallel()

			repo, close, err := ri.new(ctx, ri.config)
			if assert.NoError(t, err) {
				defer close()

				err = repo.Store(adder.RedirectStorage{
					Code:  code,
					Token: token,
					URL:   url,
				})
				if assert.NoError(t, err) {
					err := repo.Invalidate(code, token)
					if assert.NoError(t, err) {
						err = repo.Store(adder.RedirectStorage{
							Code:  code,
							Token: token,
							URL:   url,
						})

						assert.ErrorIs(t, err, adder.ErrDuplicate)
					}
				}
			}
		})
	}
}
