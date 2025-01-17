package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"url-shortner/internal/config"
	"url-shortner/internal/storage"
)

const (
	createUrlTableQuery = `
		CREATE TABLE IF NOT EXISTS url (
		    id SERIAL PRIMARY KEY,
		    alias TEXT NOT NULL UNIQUE,
		    url TEXT NOT NULL);
		CREATE INDEX IF NOT EXISTS idx_alias ON url (alias);
	`
	insertUrlQuery = `
		INSERT INTO url (alias, url)
		VALUES ($1, $2)
		RETURNING id;
	`
	getUrlQuery        = `SELECT url FROM url WHERE alias = $1;`
	deleteUrlQuery     = `DELETE FROM url WHERE alias = $1;`
	ErrUniqueViolation = "23505"
)

type Storage struct {
	db *sqlx.DB
}

func New(cfg config.PostgresConfig) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sqlx.Open(
		"postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(createUrlTableQuery)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	stmt, err := s.db.Preparex(insertUrlQuery) // (alias, url)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var id int64

	err = stmt.Get(&id, alias, urlToSave)

	if err != nil {
		var sqlxerr *pq.Error
		if errors.As(err, &sqlxerr) && sqlxerr.Code == ErrUniqueViolation {
			return 0, storage.ErrUrlExist
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	stmt, err := s.db.Preparex(getUrlQuery)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var url string
	err = stmt.Get(&url, alias)

	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrUrlNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	stmt, err := s.db.Preparex(deleteUrlQuery)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
