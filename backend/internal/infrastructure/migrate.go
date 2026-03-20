package infrastructure

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/jmoiron/sqlx"
)

// RunMigrations executes all .sql files in the given directory in sorted order.
// Each file is run inside a transaction and skipped if already applied.
func RunMigrations(db *sqlx.DB, dir string, logger *slog.Logger) error {
	// Create tracking table.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("no migrations directory found, skipping", "dir", dir)
			return nil
		}
		return fmt.Errorf("reading migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		var count int
		if err := db.Get(&count, "SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", f); err != nil {
			return fmt.Errorf("checking migration %s: %w", f, err)
		}
		if count > 0 {
			logger.Debug("migration already applied, skipping", "file", f)
			continue
		}

		sql, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", f, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("beginning tx for %s: %w", f, err)
		}
		if _, err := tx.Exec(string(sql)); err != nil {
			tx.Rollback()
			return fmt.Errorf("running migration %s: %w", f, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", f); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", f, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", f, err)
		}
		logger.Info("applied migration", "file", f)
	}

	return nil
}
