package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"url-shortner/internal/config"
	"url-shortner/internal/stats"
	"url-shortner/internal/storage"
)

const (
	insertUrlQuery = `
		INSERT INTO url (alias, url)
		VALUES ($1, $2)
		RETURNING id;
	`
	getUrlQuery      = `SELECT url FROM url WHERE alias = $1;`
	deleteUrlQuery   = `DELETE FROM url WHERE alias = $1;`
	getUrlCountQuery = `SELECT COUNT(*) FROM url;`

	getLastPeakRate = `SELECT day_peak FROM analytics WHERE id = 1`
	updateStats     = `
		INSERT INTO analytics (id, total_url_count, url_per_min, day_peak) 
		VALUES (1, $1, $2, $3) 
		ON CONFLICT (id) 
		DO UPDATE SET total_url_count = EXCLUDED.total_url_count, 
					  url_per_min = EXCLUDED.url_per_min, 
					  day_peak = EXCLUDED.day_peak;
	`
	resetPeakRate = "UPDATE analytics SET day_peak = 0 WHERE id = 1"

	ErrUniqueViolation = "23505"
)

type Storage struct {
	db *sqlx.DB
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sqlx.Open(
		"postgres",
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.DBUser, cfg.DBPassword, cfg.Database, cfg.SSLMode),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = initMigrations(cfg, db)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func initMigrations(cfg *config.Config, db *sqlx.DB) error {
	const op = "storage.postgres.initMigrations"

	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	isExist, err := isTableExists(db, "url")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if isExist && cfg.IsReload {
		err = goose.Reset(db.DB, cfg.MigrationsPath)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	err = goose.Up(db.DB, cfg.MigrationsPath)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func isTableExists(db *sqlx.DB, tableName string) (bool, error) {
	const op = "storage.postgres.isTableExists"

	var exists bool
	query := `SELECT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = $1
    )`

	err := db.Get(&exists, query, tableName)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return exists, nil
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

func (s *Storage) GetLastPeakRate() (int, error) {
	const op = "storage.postgres.GetLastPeakRate"

	stmt, err := s.db.Preparex(getLastPeakRate)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var peakRate int
	err = stmt.Get(&peakRate)

	if errors.Is(err, sql.ErrNoRows) {
		return 0, storage.ErrUrlNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return peakRate, nil
}

func (s *Storage) GetURLCount() (int, error) {
	const op = "storage.postgres.GetURLCount"

	stmt, err := s.db.Preparex(getUrlCountQuery)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var count int
	err = stmt.Get(&count)

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return count, nil
}

func (s *Storage) UpdateStats(newStats stats.Statistic) error {
	const op = "storage.postgres.UpdateStats"

	stmt, err := s.db.Preparex(updateStats)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(newStats.TotalURLCount, newStats.UrlPerMinute, newStats.DayPeak)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) ResetPeakRate() error {
	const op = "storage.postgres.ResetPeakRate"

	stmt, err := s.db.Preparex(resetPeakRate)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// TODO: create method for getting stats
//func (s *Storage) GetStats() (stats.Statistic, error) {
//	const op = "storage.postgres.GetStats"
//}
