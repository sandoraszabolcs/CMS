package domain

import "time"

// PassengerCategory represents the CMS beneficiary type per Romanian legislation.
type PassengerCategory string

const (
	CategoryRegular    PassengerCategory = "regular"    // regular passenger, no subsidy
	CategoryStudent    PassengerCategory = "student"    // Legea 198/2023
	CategoryUniversity PassengerCategory = "university" // Legea 199/2023
	CategoryPensioner  PassengerCategory = "pensioner"  // Legea 147/2000
	CategoryDisabled   PassengerCategory = "disabled"   // Legea 448/2006
	CategoryVeteran    PassengerCategory = "veteran"    // Legea 44/1994
)

// EventType represents a validation event direction.
type EventType string

const (
	EventCheckin  EventType = "checkin"
	EventCheckout EventType = "checkout"
)

// Stop represents a transit stop on a route.
type Stop struct {
	ID   string  `db:"id"   json:"id"`
	Name string  `db:"name" json:"name"`
	Lat  float64 `db:"lat"  json:"lat"`
	Lng  float64 `db:"lng"  json:"lng"`
}

// Vehicle represents a bus operating on a line.
type Vehicle struct {
	ID            string    `db:"id"              json:"id"`
	Line          string    `db:"line"            json:"line"`
	CurrentStopID string    `db:"current_stop_id" json:"current_stop_id"`
	Lat           float64   `db:"lat"             json:"lat"`
	Lng           float64   `db:"lng"             json:"lng"`
	UpdatedAt     time.Time `db:"updated_at"      json:"updated_at"`
}

// Passenger represents a CMS card holder — subsidized or regular.
type Passenger struct {
	CardID   string            `db:"card_id"   json:"card_id"`
	Name     string            `db:"name"      json:"name"`
	Category PassengerCategory `db:"category"  json:"category"`
	IsActive bool              `db:"is_active" json:"is_active"`
}

// ValidationEvent represents a checkin or checkout at a stop.
type ValidationEvent struct {
	ID        int64     `db:"id"         json:"id"`
	CardID    string    `db:"card_id"    json:"card_id"`
	VehicleID string    `db:"vehicle_id" json:"vehicle_id"`
	EventType EventType `db:"event_type" json:"event_type"`
	StopID    string    `db:"stop_id"    json:"stop_id"`
	Lat       float64   `db:"lat"        json:"lat"`
	Lng       float64   `db:"lng"        json:"lng"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// RecentEvent extends ValidationEvent with passenger info for the feed.
type RecentEvent struct {
	ValidationEvent
	PassengerName     string            `db:"passenger_name"     json:"passenger_name"`
	PassengerCategory PassengerCategory `db:"passenger_category" json:"passenger_category"`
	StopName          string            `db:"stop_name"          json:"stop_name"`
}

// ODMatrixRow represents an aggregated origin-destination pair.
type ODMatrixRow struct {
	OriginStop      string `db:"origin_stop"      json:"origin_stop"`
	OriginName      string `db:"origin_name"      json:"origin_name"`
	DestinationStop string `db:"destination_stop" json:"destination_stop"`
	DestinationName string `db:"destination_name" json:"destination_name"`
	TripCount       int    `db:"trip_count"       json:"trip_count"`
}

// Stats holds dashboard-level statistics.
type Stats struct {
	TotalTripsToday        int            `json:"total_trips_today"`
	MostPopularOrigin      string         `json:"most_popular_origin"`
	MostPopularDestination string         `json:"most_popular_destination"`
	TripsByCategory        map[string]int `json:"trips_by_category"`
	TripsByHour            map[int]int    `json:"trips_by_hour"`
}

// CheckinRequest holds the incoming checkin data.
type CheckinRequest struct {
	CardID    string `json:"card_id"    binding:"required"`
	VehicleID string `json:"vehicle_id" binding:"required"`
	StopID    string `json:"stop_id"    binding:"required"`
}

// CheckoutRequest holds the incoming checkout data.
type CheckoutRequest struct {
	CardID    string `json:"card_id"    binding:"required"`
	VehicleID string `json:"vehicle_id" binding:"required"`
	StopID    string `json:"stop_id"    binding:"required"`
}
