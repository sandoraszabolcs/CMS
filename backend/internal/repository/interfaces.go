package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/szabolcs/cms/internal/domain"
)

// PassengerRepository provides access to passenger data.
type PassengerRepository interface {
	FindByCardID(ctx context.Context, cardID string) (domain.Passenger, error)
}

// ValidationRepository provides access to validation events.
type ValidationRepository interface {
	FindOpenCheckin(ctx context.Context, cardID string) (domain.ValidationEvent, error)
	InsertEvent(ctx context.Context, event domain.ValidationEvent) (domain.ValidationEvent, error)
	InsertEventAt(ctx context.Context, event domain.ValidationEvent, at time.Time) (domain.ValidationEvent, error)
	InsertEventTx(ctx context.Context, tx *sqlx.Tx, event domain.ValidationEvent) (domain.ValidationEvent, error)
	RecentEvents(ctx context.Context, limit int) ([]domain.RecentEvent, error)
	CountToday(ctx context.Context) (int, error)
	DeleteAll(ctx context.Context) error
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

// VehicleRepository provides access to vehicle data.
type VehicleRepository interface {
	FindAll(ctx context.Context) ([]domain.Vehicle, error)
	UpdatePosition(ctx context.Context, id string, stopID string, lat, lng float64) error
}

// StopRepository provides access to stop data.
type StopRepository interface {
	FindAll(ctx context.Context) ([]domain.Stop, error)
	FindByID(ctx context.Context, id string) (domain.Stop, error)
}

// ODMatrixRepository provides access to the OD matrix materialized view.
type ODMatrixRepository interface {
	GetAll(ctx context.Context) ([]domain.ODMatrixRow, error)
	Refresh(ctx context.Context) error
}

// StatsRepository provides aggregated statistics.
type StatsRepository interface {
	GetStats(ctx context.Context) (domain.Stats, error)
}
