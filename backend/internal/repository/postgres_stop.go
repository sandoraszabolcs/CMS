package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/szabolcs/cms/internal/domain"
)

type postgresStopRepo struct {
	db *sqlx.DB
}

// NewPostgresStopRepo creates a StopRepository backed by PostgreSQL.
func NewPostgresStopRepo(db *sqlx.DB) StopRepository {
	return &postgresStopRepo{db: db}
}

func (r *postgresStopRepo) FindAll(ctx context.Context) ([]domain.Stop, error) {
	var stops []domain.Stop
	err := r.db.SelectContext(ctx, &stops, "SELECT id, name, lat, lng FROM stops ORDER BY id")
	if err != nil {
		return nil, err
	}
	return stops, nil
}

func (r *postgresStopRepo) FindByID(ctx context.Context, id string) (domain.Stop, error) {
	var stop domain.Stop
	err := r.db.GetContext(ctx, &stop, "SELECT id, name, lat, lng FROM stops WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Stop{}, domain.ErrNotFound
		}
		return domain.Stop{}, err
	}
	return stop, nil
}
