package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"time"
	"url-shortner/internal/config"
	"url-shortner/internal/models/database"
	"url-shortner/internal/stats"
	"url-shortner/internal/storage"
)

const (
	isTableExistsQuery = `SELECT EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_name = $1
    )`

	insertUrlQuery = `
		INSERT INTO url (alias, url)
		VALUES ($1, $2)
		RETURNING id;
	`
	getUrlQuery      = `SELECT url FROM url WHERE alias = $1;`
	deleteUrlQuery   = `DELETE FROM url WHERE alias = $1;`
	getUrlCountQuery = `SELECT COUNT(*) FROM url;`

	getLastPeakRate = `SELECT day_peak, updated_at FROM analytics WHERE id = 1;`
	getStats        = "SELECT total_url_count, leaders, day_peak FROM analytics WHERE id = 1;"
	updateStats     = `
		INSERT INTO analytics (id, total_url_count, leaders, day_peak, updated_at) 
		VALUES (1, $1, $2, $3, $4) 
		ON CONFLICT (id) 
		DO UPDATE SET total_url_count = EXCLUDED.total_url_count, 
					  leaders = EXCLUDED.leaders,
					  day_peak = EXCLUDED.day_peak,
					  updated_at = EXCLUDED.updated_at;
	`
	resetPeakRate = `UPDATE analytics SET day_peak = 0 WHERE id = 1;`
	getLeaders    = `
		SELECT 
			SPLIT_PART(url, '/', 1) AS resource,
			COUNT(*) AS url_count
		FROM url
		GROUP BY resource
		ORDER BY url_count DESC
		LIMIT 3;
	`

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
	query := isTableExistsQuery

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

func (s *Storage) GetLastPeakRate() (stats.DayPeakStatistic, error) {
	const op = "storage.postgres.GetLastPeakRate"

	stmt, err := s.db.Preparex(getLastPeakRate)
	if err != nil {
		return stats.DayPeakStatistic{}, fmt.Errorf("%s: %w", op, err)
	}

	var peakRateStats stats.DayPeakStatistic
	err = stmt.Get(&peakRateStats)

	if errors.Is(err, sql.ErrNoRows) {
		return stats.DayPeakStatistic{}, storage.ErrUrlNotFound
	}
	if err != nil {
		return stats.DayPeakStatistic{}, fmt.Errorf("%s: %w", op, err)
	}

	return peakRateStats, nil
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

func (s *Storage) GetResourcesLeaders() ([]stats.ResourceInfo, error) {
	const op = "storage.postgres.GetResourcesLeaders"

	stmt, err := s.db.Preparex(getLeaders)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var resources []stats.ResourceInfo
	err = stmt.Select(&resources)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resources, nil
}

func (s *Storage) UpdateStats(newStats database.Statistic) error {
	const op = "storage.postgres.UpdateStats"

	stmt, err := s.db.Preparex(updateStats)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	now := time.Now()
	_, err = stmt.Exec(newStats.TotalURLCount, newStats.LeadersJSON, newStats.DayPeak, now)
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

func (s *Storage) GetStats() (database.Statistic, error) {
	const op = "storage.postgres.GetStats"

	stmt, err := s.db.Preparex(getStats)
	if err != nil {
		return database.Statistic{}, fmt.Errorf("%s: %w", op, err)
	}

	var statistics database.Statistic
	err = stmt.Get(&statistics)
	if err != nil {
		return database.Statistic{}, fmt.Errorf("%s: %w", op, err)
	}

	return statistics, nil
}
