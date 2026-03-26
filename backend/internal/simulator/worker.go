package simulator

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/szabolcs/cms/internal/domain"
	"github.com/szabolcs/cms/internal/infrastructure"
	"github.com/szabolcs/cms/internal/repository"
)

// Worker defines the simulator lifecycle.
type Worker interface {
	Start(ctx context.Context) error
	Stop() error
}

// Resetter allows the simulator to trigger a data reset.
type Resetter interface {
	Reset(ctx context.Context) error
}

// Deps holds the simulator's dependencies.
type Deps struct {
	Validations repository.ValidationRepository
	Vehicles    repository.VehicleRepository
	Stops       repository.StopRepository
	Passengers  repository.PassengerRepository
	Resetter    Resetter
	Redis       *redis.Client
	Logger      *slog.Logger
	Interval    time.Duration
}

type simulator struct {
	deps   Deps
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Route state: circular S1→S2→...→S6→S1
	stops []domain.Stop

	// Passenger cache: card_id → passenger.
	passengers map[string]domain.Passenger

	// Vehicle state: current stop index per vehicle.
	mu             sync.Mutex
	vehicleStopIdx map[string]int
	checkedIn      map[string]bool // card_id → is checked in
}

// New creates a new simulator Worker.
func New(deps Deps) Worker {
	return &simulator{
		deps:           deps,
		passengers:     make(map[string]domain.Passenger),
		vehicleStopIdx: make(map[string]int),
		checkedIn:      make(map[string]bool),
	}
}

func (s *simulator) Start(ctx context.Context) error {
	// Load stops for route.
	stops, err := s.deps.Stops.FindAll(ctx)
	if err != nil {
		return err
	}
	s.stops = stops

	// Load vehicles and initialize stop indices.
	vehicles, err := s.deps.Vehicles.FindAll(ctx)
	if err != nil {
		return err
	}
	for _, v := range vehicles {
		for i, st := range s.stops {
			if st.ID == v.CurrentStopID {
				s.vehicleStopIdx[v.ID] = i
				break
			}
		}
	}

	// Load passengers.
	cardIDs := []string{"CMS-001", "CMS-002", "CMS-003", "CMS-004", "CMS-005", "CMS-006", "CMS-007", "CMS-008"}
	for _, id := range cardIDs {
		p, err := s.deps.Passengers.FindByCardID(ctx, id)
		if err != nil {
			return err
		}
		s.passengers[id] = p
	}

	ctx, s.cancel = context.WithCancel(ctx)
	s.wg.Add(1)
	go s.run(ctx, vehicles)
	return nil
}

func (s *simulator) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
	return nil
}

// isPeakHour returns true for hours 7-9 and 16-18.
func isPeakHour(h int) bool {
	return (h >= 7 && h <= 9) || (h >= 16 && h <= 18)
}

const maxTripsBeforeReset = 1000

func (s *simulator) run(ctx context.Context, vehicles []domain.Vehicle) {
	defer s.wg.Done()
	ticker := time.NewTicker(s.deps.Interval)
	defer ticker.Stop()

	cardIDs := make([]string, 0, len(s.passengers))
	for id := range s.passengers {
		cardIDs = append(cardIDs, id)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Auto-reset at threshold.
			count, err := s.deps.Validations.CountToday(ctx)
			if err == nil && count >= maxTripsBeforeReset {
				s.deps.Logger.Info("simulator: trip limit reached, resetting", "count", count)
				if err := s.deps.Resetter.Reset(ctx); err != nil {
					s.deps.Logger.Error("simulator: reset failed", "error", err)
				}
				continue
			}

			// Generate more events during peak hours.
			n := 1
			if isPeakHour(time.Now().UTC().Hour()) {
				n = 3 + rand.IntN(3) // 3-5 events
			}
			for range n {
				s.tick(ctx, vehicles, cardIDs)
			}
		}
	}
}

func (s *simulator) tick(ctx context.Context, vehicles []domain.Vehicle, passengers []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Pick random vehicle and advance it.
	v := vehicles[rand.IntN(len(vehicles))]
	stopIdx := (s.vehicleStopIdx[v.ID] + 1) % len(s.stops)
	s.vehicleStopIdx[v.ID] = stopIdx
	stop := s.stops[stopIdx]

	// Update vehicle position.
	if err := s.deps.Vehicles.UpdatePosition(ctx, v.ID, stop.ID, stop.Lat, stop.Lng); err != nil {
		s.deps.Logger.Error("simulator: failed to update vehicle", "error", err)
		return
	}

	// Pick random passenger and generate appropriate event.
	cardID := passengers[rand.IntN(len(passengers))]
	passenger := s.passengers[cardID]
	var event domain.ValidationEvent

	if s.checkedIn[cardID] {
		// Generate checkout.
		event = domain.ValidationEvent{
			CardID:    cardID,
			VehicleID: v.ID,
			EventType: domain.EventCheckout,
			StopID:    stop.ID,
			Lat:       stop.Lat,
			Lng:       stop.Lng,
		}
		s.checkedIn[cardID] = false
	} else {
		// Generate checkin.
		event = domain.ValidationEvent{
			CardID:    cardID,
			VehicleID: v.ID,
			EventType: domain.EventCheckin,
			StopID:    stop.ID,
			Lat:       stop.Lat,
			Lng:       stop.Lng,
		}
		s.checkedIn[cardID] = true
	}

	inserted, err := s.deps.Validations.InsertEvent(ctx, event)
	if err != nil {
		s.deps.Logger.Error("simulator: failed to insert event", "error", err)
		return
	}

	s.deps.Logger.Info("simulator: event generated",
		"event_type", inserted.EventType,
		"card_id", inserted.CardID,
		"vehicle_id", inserted.VehicleID,
		"stop", stop.Name,
	)

	// Publish enriched event to Redis.
	rich := domain.RecentEvent{
		ValidationEvent:   inserted,
		PassengerName:     passenger.Name,
		PassengerCategory: passenger.Category,
		StopName:          stop.Name,
	}
	data, err := json.Marshal(rich)
	if err != nil {
		s.deps.Logger.Error("simulator: failed to marshal event", "error", err)
		return
	}
	if err := s.deps.Redis.Publish(ctx, infrastructure.RedisChannelValidationEvents, data).Err(); err != nil {
		s.deps.Logger.Error("simulator: failed to publish event", "error", err)
	}
}
