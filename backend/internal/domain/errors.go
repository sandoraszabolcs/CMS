package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrOpenCheckinExists = errors.New("open checkin exists")
	ErrPassengerInactive = errors.New("passenger is not active")
)
