package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/szabolcs/cms/internal/domain"
)

type postgresValidationRepo struct {
	db *sqlx.DB
}

// NewPostgresValidationRepo creates a ValidationRepository backed by PostgreSQL.
func NewPostgresValidationRepo(db *sqlx.DB) ValidationRepository {
	return &postgresValidationRepo{db: db}
}

func (r *postgresValidationRepo) FindOpenCheckin(ctx context.Context, cardID string) (domain.ValidationEvent, error) {
	// Find the most recent checkin that has no subsequent checkout.
	query := `
		SELECT ve.id, ve.card_id, ve.vehicle_id, ve.event_type, ve.stop_id, ve.lat, ve.lng, ve.created_at
		FROM validation_events ve
		WHERE ve.card_id = $1
		  AND ve.event_type = 'checkin'
		  AND NOT EXISTS (
		      SELECT 1 FROM validation_events co
		      WHERE co.card_id = ve.card_id
		        AND co.event_type = 'checkout'
		        AND co.created_at > ve.created_at
		  )
		ORDER BY ve.created_at DESC
		LIMIT 1`

	var ev domain.ValidationEvent
	err := r.db.GetContext(ctx, &ev, query, cardID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ValidationEvent{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.ValidationEvent{}, err
	}
	return ev, nil
}

func (r *postgresValidationRepo) InsertEvent(ctx context.Context, event domain.ValidationEvent) (domain.ValidationEvent, error) {
	query := `
		INSERT INTO validation_events (card_id, vehicle_id, event_type, stop_id, lat, lng)
		VALUES (:card_id, :vehicle_id, :event_type, :stop_id, :lat, :lng)
		RETURNING id, created_at`

	rows, err := r.db.NamedQueryContext(ctx, query, event)
	if err != nil {
		return domain.ValidationEvent{}, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&event.ID, &event.CreatedAt); err != nil {
			return domain.ValidationEvent{}, err
		}
	}
	return event, nil
}

func (r *postgresValidationRepo) InsertEventTx(ctx context.Context, tx *sqlx.Tx, event domain.ValidationEvent) (domain.ValidationEvent, error) {
	query := `
		INSERT INTO validation_events (card_id, vehicle_id, event_type, stop_id, lat, lng)
		VALUES (:card_id, :vehicle_id, :event_type, :stop_id, :lat, :lng)
		RETURNING id, created_at`

	rows, err := tx.NamedQuery(query, event)
	if err != nil {
		return domain.ValidationEvent{}, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&event.ID, &event.CreatedAt); err != nil {
			return domain.ValidationEvent{}, err
		}
	}
	return event, nil
}

func (r *postgresValidationRepo) RecentEvents(ctx context.Context, limit int) ([]domain.RecentEvent, error) {
	query := `
		SELECT ve.id, ve.card_id, ve.vehicle_id, ve.event_type, ve.stop_id, ve.lat, ve.lng, ve.created_at,
		       p.name AS passenger_name, p.category AS passenger_category, s.name AS stop_name
		FROM validation_events ve
		JOIN passengers p ON p.card_id = ve.card_id
		JOIN stops s ON s.id = ve.stop_id
		ORDER BY ve.created_at DESC
		LIMIT $1`

	var events []domain.RecentEvent
	err := r.db.SelectContext(ctx, &events, query, limit)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *postgresValidationRepo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}
