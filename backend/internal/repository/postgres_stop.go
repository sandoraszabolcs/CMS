package repository

import (
	"context"

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
