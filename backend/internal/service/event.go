package service

import (
	"context"

	"github.com/szabolcs/cms/internal/domain"
	"github.com/szabolcs/cms/internal/repository"
)

// EventLister lists recent validation events.
type EventLister interface {
	RecentEvents(ctx context.Context, limit int) ([]domain.RecentEvent, error)
}

// EventService implements EventLister.
type EventService struct {
	repo repository.ValidationRepository
}

func NewEventService(repo repository.ValidationRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) RecentEvents(ctx context.Context, limit int) ([]domain.RecentEvent, error) {
	return s.repo.RecentEvents(ctx, limit)
}
