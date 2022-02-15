package gormsqlite

import (
	"context"
	"errors"
	"hex-microservice/adder"
	"hex-microservice/invalidator"
	"hex-microservice/lookup"
	"hex-microservice/repository"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type gormSqliteRepository struct {
	parent context.Context
	db     *gorm.DB
}

func New(parent context.Context, url string) (repository.RedirectRepository, repository.Close, error) {
	dsn := strings.TrimPrefix(url, "sqlite://")
	database, err := gorm.Open("sqlite3", dsn)
	if err != nil {
		panic("Failed to connect to database!")
	}

	database.AutoMigrate(&redirect{})

	return &gormSqliteRepository{
		parent: parent,
		db:     database,
	}, database.Close, nil
}

func (g *gormSqliteRepository) Lookup(code string) (lookup.RedirectStorage, error) {
	var red lookup.RedirectStorage
	var stored redirect

	if err := g.db.Where("code = ? and active = ?", code, true).First(&stored).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return red, lookup.ErrNotFound
		}

		return red, err
	}

	return fromRedirectToLookupRedirectStorage(stored), nil
}

// see: https://github.com/go-gorm/gorm/issues/2903
func isDuplicateKeyError(err error) bool {
	return strings.HasPrefix(err.Error(), "UNIQUE constraint failed")
}

func (g *gormSqliteRepository) Store(red adder.RedirectStorage) error {
	store := fromAdderRedirectStorageToRedirect(red)
	store.Active = true

	if err := g.db.Create(store).Error; err != nil {
		if isDuplicateKeyError(err) {
			return adder.ErrDuplicate
		}

		return err
	}

	return nil
}

func (g *gormSqliteRepository) Invalidate(code, token string) error {
	var stored redirect

	if err := g.db.Where("code = ? AND token = ?", code, token).First(&stored).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return invalidator.ErrNotFound
		}

		return err
	}

	return g.db.Model(&stored).Update("active", false).Error
}
