package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/szabolcs/cms/internal/domain"
)

type postgresVehicleRepo struct {
	db *sqlx.DB
}

// NewPostgresVehicleRepo creates a VehicleRepository backed by PostgreSQL.
func NewPostgresVehicleRepo(db *sqlx.DB) VehicleRepository {
	return &postgresVehicleRepo{db: db}
}

func (r *postgresVehicleRepo) FindAll(ctx context.Context) ([]domain.Vehicle, error) {
	var vehicles []domain.Vehicle
	err := r.db.SelectContext(ctx, &vehicles, "SELECT id, line, current_stop_id, lat, lng, updated_at FROM vehicles ORDER BY id")
	if err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (r *postgresVehicleRepo) UpdatePosition(ctx context.Context, id string, stopID string, lat, lng float64) error {
	query := `UPDATE vehicles SET current_stop_id = $1, lat = $2, lng = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, stopID, lat, lng, id)
	return err
}
