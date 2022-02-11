// Package sqlite used the CGO sqlite driver from Yasuhiro Matsumoto (mattn)
// and migrations. It uses no 'storage object' and uses no mapping. The mapping
// is directly performed with the SQL queries.
package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
	"hex-microservice/repository"
	"net/url"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteRepository struct {
	parent context.Context
	db     *sql.DB
}

//go:embed migrations/*.sql
var fs embed.FS

const tableName = "redirects"

// databaseUp migrates the database to the latest schema.
func databaseUp(database *sql.DB) error {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}

	driver, err := sqlite.WithInstance(database, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", d, "sqlite", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		return err
	}

	return nil
}

// New creates a new repository using sqlite as backend.
func New(parent context.Context, config string) (repository.RedirectRepository, repository.Close, error) {
	urlParts, err := url.Parse(config)
	if err != nil {
		return nil, nil, err
	}

	// TODO: use and apply properties
	_ = urlParts

	database, err := sql.Open("sqlite3", config)
	if err != nil {
		return nil, nil, err
	}

	if err := databaseUp(database); err != nil {
		return nil, nil, err
	}

	return &sqliteRepository{
		parent: parent,
		db:     database,
	}, database.Close, nil
}

// LookupFind is the implementation for repository.RedirectRepository#LookupFind.
func (s *sqliteRepository) LookupFind(code string) (lookup.RedirectStorage, error) {
	var red lookup.RedirectStorage

	row := s.db.QueryRow(fmt.Sprintf(`
	SELECT
		code, url, created_at
	FROM '%s'
	WHERE
		code = ?
	`, tableName), code)

	var createdAt string
	if err := row.Scan(&red.Code, &red.URL, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return red, lookup.ErrNotFound
		}

		return red, err
	}

	// Special handling for the timestamp
	var err error
	red.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return red, fmt.Errorf("repository.LookupFind parsing time: %w", err)
	}

	return red, nil
}

// Store is the implementation for repository.RedirectRepository#Store.
func (s *sqliteRepository) Store(red adder.RedirectStorage) error {
	if _, err := s.db.Exec(fmt.Sprintf(`
	INSERT INTO '%s'
		(code, url, token, client_info, created_at)
	VALUES
		(?, ?, ?, ?, ?)
	`, tableName), red.Code, red.URL, red.Token, red.ClientInfo, red.CreatedAt.Format(time.RFC3339)); err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintPrimaryKey) {
				return adder.ErrDuplicate
			}
		}
		return err
	}

	return nil
}

func (r *sqliteRepository) Delete(code, token string) error {
	result, err := r.db.Exec(fmt.Sprintf(`
	DELETE
  FROM '%s'
	WHERE
		code = ? AND token = ?
	`, tableName), code, token)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return deleter.ErrNotFound
	}

	return err
}

func (s *sqliteRepository) DeleteFind(code string) (deleter.RedirectStorage, error) {
	var red deleter.RedirectStorage

	row := s.db.QueryRow(fmt.Sprintf(`
	SELECT
		code, token
  FROM '%s'
	WHERE
		code = ?
	`, tableName), code)

	if err := row.Scan(&red.Code, &red.Token); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return red, lookup.ErrNotFound
		}

		return red, err
	}

	return red, nil
}
