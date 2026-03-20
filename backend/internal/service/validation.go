package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/szabolcs/cms/internal/domain"
	"github.com/szabolcs/cms/internal/infrastructure"
	"github.com/szabolcs/cms/internal/repository"
)

// Validator defines the business operations for check-in/checkout.
type Validator interface {
	Checkin(ctx context.Context, req domain.CheckinRequest) (domain.ValidationEvent, error)
	Checkout(ctx context.Context, req domain.CheckoutRequest) (domain.ValidationEvent, error)
}

// ValidationService handles checkin/checkout business logic.
type ValidationService struct {
	passengers  repository.PassengerRepository
	validations repository.ValidationRepository
	stops       repository.StopRepository
	rdb         *redis.Client
	logger      *slog.Logger
}

// NewValidationService creates a ValidationService with its dependencies.
func NewValidationService(
	passengers repository.PassengerRepository,
	validations repository.ValidationRepository,
	stops repository.StopRepository,
	rdb *redis.Client,
	logger *slog.Logger,
) *ValidationService {
	return &ValidationService{
		passengers:  passengers,
		validations: validations,
		stops:       stops,
		rdb:         rdb,
		logger:      logger,
	}
}

// Checkin validates a passenger and records a checkin event.
// If an open checkin exists, it auto-checkouts first within a transaction.
func (s *ValidationService) Checkin(ctx context.Context, req domain.CheckinRequest) (domain.ValidationEvent, error) {
	passenger, err := s.passengers.FindByCardID(ctx, req.CardID)
	if err != nil {
		return domain.ValidationEvent{}, err
	}
	if !passenger.IsActive {
		return domain.ValidationEvent{}, domain.ErrPassengerInactive
	}

	stop, err := s.stops.FindByID(ctx, req.StopID)
	if err != nil {
		return domain.ValidationEvent{}, err
	}

	tx, err := s.validations.BeginTx(ctx)
	if err != nil {
		return domain.ValidationEvent{}, err
	}
	defer func() {
		if err != nil && tx != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.logger.Error("rollback failed", "error", rbErr)
			}
		}
	}()

	// Check for open checkin and auto-checkout if needed.
	openCheckin, findErr := s.validations.FindOpenCheckin(ctx, req.CardID)
	if findErr != nil && !errors.Is(findErr, domain.ErrNotFound) {
		err = findErr
		return domain.ValidationEvent{}, err
	}
	if findErr == nil {
		// Auto-checkout at the current stop.
		autoCheckout := domain.ValidationEvent{
			CardID:    req.CardID,
			VehicleID: openCheckin.VehicleID,
			EventType: domain.EventCheckout,
			StopID:    stop.ID,
			Lat:       stop.Lat,
			Lng:       stop.Lng,
		}
		checkoutEvent, insertErr := s.validations.InsertEventTx(ctx, tx, autoCheckout)
		if insertErr != nil {
			err = insertErr
			return domain.ValidationEvent{}, err
		}
		s.logger.Info("auto-checkout performed",
			"card_id", req.CardID,
			"stop_id", stop.ID,
			"checkout_id", checkoutEvent.ID,
		)
		s.publishEvent(ctx, checkoutEvent, passenger, stop)
	}

	// Insert the new checkin.
	checkinEvent := domain.ValidationEvent{
		CardID:    req.CardID,
		VehicleID: req.VehicleID,
		EventType: domain.EventCheckin,
		StopID:    stop.ID,
		Lat:       stop.Lat,
		Lng:       stop.Lng,
	}
	result, err := s.validations.InsertEventTx(ctx, tx, checkinEvent)
	if err != nil {
		return domain.ValidationEvent{}, err
	}

	if tx != nil {
		if err = tx.Commit(); err != nil {
			return domain.ValidationEvent{}, err
		}
	}

	s.publishEvent(ctx, result, passenger, stop)
	return result, nil
}

// Checkout records a checkout event for a passenger.
func (s *ValidationService) Checkout(ctx context.Context, req domain.CheckoutRequest) (domain.ValidationEvent, error) {
	passenger, err := s.passengers.FindByCardID(ctx, req.CardID)
	if err != nil {
		return domain.ValidationEvent{}, err
	}
	if !passenger.IsActive {
		return domain.ValidationEvent{}, domain.ErrPassengerInactive
	}

	stop, err := s.stops.FindByID(ctx, req.StopID)
	if err != nil {
		return domain.ValidationEvent{}, err
	}

	event := domain.ValidationEvent{
		CardID:    req.CardID,
		VehicleID: req.VehicleID,
		EventType: domain.EventCheckout,
		StopID:    stop.ID,
		Lat:       stop.Lat,
		Lng:       stop.Lng,
	}
	result, err := s.validations.InsertEvent(ctx, event)
	if err != nil {
		return domain.ValidationEvent{}, err
	}

	s.publishEvent(ctx, result, passenger, stop)
	return result, nil
}

func (s *ValidationService) publishEvent(ctx context.Context, event domain.ValidationEvent, passenger domain.Passenger, stop domain.Stop) {
	if s.rdb == nil {
		return
	}
	rich := domain.RecentEvent{
		ValidationEvent:   event,
		PassengerName:     passenger.Name,
		PassengerCategory: passenger.Category,
		StopName:          stop.Name,
	}
	data, err := json.Marshal(rich)
	if err != nil {
		s.logger.Error("failed to marshal event for publish", "error", err)
		return
	}
	if err := s.rdb.Publish(ctx, infrastructure.RedisChannelValidationEvents, data).Err(); err != nil {
		s.logger.Error("failed to publish event to redis", "error", err)
	}
}
