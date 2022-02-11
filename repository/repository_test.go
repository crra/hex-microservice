package repository_test

import (
	"context"
	"hex-microservice/adder"
	"hex-microservice/deleter"
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
		// 	//"file::memory:?cache=shared&_journal_mode=WAL&_foreign_keys=true",
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

				_, err := repo.LookupFind(code)
				assert.ErrorIs(t, err, lookup.ErrNotFound)
			}
		})
	}
}

func TestRedirectAddAndReadBack(t *testing.T) {
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
					lookupFind, err := repo.LookupFind(code)
					if assert.NoError(t, err) {
						deleteFind, err := repo.DeleteFind(code)
						if assert.NoError(t, err) {
							assert.Equal(t, code, lookupFind.Code)
							assert.Equal(t, code, deleteFind.Code)
						}
					}
				}
			}
		})
	}
}

func TestRedirectAddTwice(t *testing.T) {
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

func TestRedirectRemoveNonExisting(t *testing.T) {
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

				err := repo.Delete(code, token)
				assert.ErrorIs(t, err, deleter.ErrNotFound)
			}
		})
	}
}

func TestRedirectRemove(t *testing.T) {
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

					err := repo.Delete(code, token)
					if assert.NoError(t, err) {
						_, err := repo.LookupFind(code)
						assert.ErrorIs(t, err, lookup.ErrNotFound)
					}
				}
			}
		})
	}
}
