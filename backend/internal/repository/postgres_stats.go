package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/szabolcs/cms/internal/domain"
)

type postgresStatsRepo struct {
	db *sqlx.DB
}

// NewPostgresStatsRepo creates a StatsRepository backed by PostgreSQL.
func NewPostgresStatsRepo(db *sqlx.DB) StatsRepository {
	return &postgresStatsRepo{db: db}
}

func (r *postgresStatsRepo) GetStats(ctx context.Context) (domain.Stats, error) {
	stats := domain.Stats{
		TripsByCategory: make(map[string]int),
		TripsByHour:     make(map[int]int),
	}

	// Total completed trips today (count of checkouts today).
	err := r.db.GetContext(ctx, &stats.TotalTripsToday, `
		SELECT COUNT(*) FROM validation_events
		WHERE event_type = 'checkout' AND created_at::date = CURRENT_DATE`)
	if err != nil {
		return domain.Stats{}, err
	}

	// Most popular origin from OD matrix.
	var origin sql.NullString
	err = r.db.GetContext(ctx, &origin, `
		SELECT origin_name FROM od_matrix ORDER BY trip_count DESC LIMIT 1`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.Stats{}, err
	}
	stats.MostPopularOrigin = origin.String

	// Most popular destination from OD matrix.
	var dest sql.NullString
	err = r.db.GetContext(ctx, &dest, `
		SELECT destination_name FROM od_matrix ORDER BY trip_count DESC LIMIT 1`)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.Stats{}, err
	}
	stats.MostPopularDestination = dest.String

	// Trips by category.
	type catCount struct {
		Category string `db:"category"`
		Count    int    `db:"count"`
	}
	var cats []catCount
	err = r.db.SelectContext(ctx, &cats, `
		SELECT p.category, COUNT(*) AS count
		FROM validation_events ve
		JOIN passengers p ON p.card_id = ve.card_id
		WHERE ve.event_type = 'checkout' AND ve.created_at::date = CURRENT_DATE
		GROUP BY p.category`)
	if err != nil {
		return domain.Stats{}, err
	}
	for _, c := range cats {
		stats.TripsByCategory[c.Category] = c.Count
	}

	// Trips by hour of day.
	type hourCount struct {
		Hour  int `db:"hour"`
		Count int `db:"count"`
	}
	var hours []hourCount
	err = r.db.SelectContext(ctx, &hours, `
		SELECT EXTRACT(HOUR FROM created_at)::int AS hour, COUNT(*) AS count
		FROM validation_events
		WHERE created_at::date = CURRENT_DATE
		GROUP BY 1
		ORDER BY 1`)
	if err != nil {
		return domain.Stats{}, err
	}
	for _, h := range hours {
		stats.TripsByHour[h.Hour] = h.Count
	}

	return stats, nil
}
