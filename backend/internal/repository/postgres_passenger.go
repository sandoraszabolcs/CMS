package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/szabolcs/cms/internal/domain"
)

type postgresPassengerRepo struct {
	db *sqlx.DB
}

// NewPostgresPassengerRepo creates a PassengerRepository backed by PostgreSQL.
func NewPostgresPassengerRepo(db *sqlx.DB) PassengerRepository {
	return &postgresPassengerRepo{db: db}
}

func (r *postgresPassengerRepo) FindByCardID(ctx context.Context, cardID string) (domain.Passenger, error) {
	var p domain.Passenger
	err := r.db.GetContext(ctx, &p, "SELECT card_id, name, category, is_active FROM passengers WHERE card_id = $1", cardID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Passenger{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Passenger{}, err
	}
	return p, nil
}
