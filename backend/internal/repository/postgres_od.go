package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/szabolcs/cms/internal/domain"
)

type postgresODMatrixRepo struct {
	db *sqlx.DB
}

// NewPostgresODMatrixRepo creates an ODMatrixRepository backed by PostgreSQL.
func NewPostgresODMatrixRepo(db *sqlx.DB) ODMatrixRepository {
	return &postgresODMatrixRepo{db: db}
}

func (r *postgresODMatrixRepo) GetAll(ctx context.Context) ([]domain.ODMatrixRow, error) {
	var rows []domain.ODMatrixRow
	err := r.db.SelectContext(ctx, &rows, "SELECT origin_stop, origin_name, destination_stop, destination_name, trip_count FROM od_matrix ORDER BY trip_count DESC")
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *postgresODMatrixRepo) Refresh(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY od_matrix")
	return err
}
