package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/szabolcs/cms/internal/domain"
)

// --- Mock repositories ---

type mockPassengerRepo struct {
	passenger domain.Passenger
	err       error
}

func (m *mockPassengerRepo) FindByCardID(_ context.Context, _ string) (domain.Passenger, error) {
	return m.passenger, m.err
}

type mockValidationRepo struct {
	openCheckin    domain.ValidationEvent
	openCheckinErr error
	insertedEvents []domain.ValidationEvent
	insertErr      error
	txCommitted    bool
}

func (m *mockValidationRepo) FindOpenCheckin(_ context.Context, _ string) (domain.ValidationEvent, error) {
	return m.openCheckin, m.openCheckinErr
}

func (m *mockValidationRepo) InsertEvent(_ context.Context, event domain.ValidationEvent) (domain.ValidationEvent, error) {
	event.ID = int64(len(m.insertedEvents) + 1)
	m.insertedEvents = append(m.insertedEvents, event)
	return event, m.insertErr
}

func (m *mockValidationRepo) InsertEventTx(_ context.Context, _ *sqlx.Tx, event domain.ValidationEvent) (domain.ValidationEvent, error) {
	event.ID = int64(len(m.insertedEvents) + 1)
	m.insertedEvents = append(m.insertedEvents, event)
	return event, m.insertErr
}

func (m *mockValidationRepo) RecentEvents(_ context.Context, _ int) ([]domain.RecentEvent, error) {
	return nil, nil
}

func (m *mockValidationRepo) BeginTx(_ context.Context) (*sqlx.Tx, error) {
	// Return nil tx — our mock InsertEventTx ignores it.
	m.txCommitted = false
	return nil, nil
}

func TestCheckin(t *testing.T) {
	tests := []struct {
		name           string
		passenger      domain.Passenger
		passengerErr   error
		openCheckin    domain.ValidationEvent
		openCheckinErr error
		wantErr        error
		wantEvents     int // expected number of inserted events
	}{
		{
			name:           "no open checkin - just checkin",
			passenger:      domain.Passenger{CardID: "CMS-001", Name: "Ion", Category: domain.CategoryStudent, IsActive: true},
			openCheckinErr: domain.ErrNotFound,
			wantErr:        nil,
			wantEvents:     1,
		},
		{
			name:      "open checkin exists - auto-checkout then checkin",
			passenger: domain.Passenger{CardID: "CMS-001", Name: "Ion", Category: domain.CategoryStudent, IsActive: true},
			openCheckin: domain.ValidationEvent{
				ID:        99,
				CardID:    "CMS-001",
				VehicleID: "BUS-101",
				EventType: domain.EventCheckin,
				StopID:    "S1",
			},
			openCheckinErr: nil,
			wantErr:        nil,
			wantEvents:     2, // auto-checkout + new checkin
		},
		{
			name:       "inactive passenger - error",
			passenger:  domain.Passenger{CardID: "CMS-001", Name: "Ion", Category: domain.CategoryStudent, IsActive: false},
			wantErr:    domain.ErrPassengerInactive,
			wantEvents: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passRepo := &mockPassengerRepo{passenger: tt.passenger, err: tt.passengerErr}
			valRepo := &mockValidationRepo{
				openCheckin:    tt.openCheckin,
				openCheckinErr: tt.openCheckinErr,
			}

			// Use nil redis client — publishEvent will log error but not crash.
			svc := NewValidationService(passRepo, valRepo, nil, slog.Default())

			req := domain.CheckinRequest{
				CardID:    "CMS-001",
				VehicleID: "BUS-101",
				StopID:    "S2",
				Lat:       44.4361,
				Lng:       26.1006,
			}

			_, err := svc.Checkin(context.Background(), req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(valRepo.insertedEvents) != tt.wantEvents {
				t.Fatalf("expected %d events, got %d", tt.wantEvents, len(valRepo.insertedEvents))
			}
		})
	}
}
