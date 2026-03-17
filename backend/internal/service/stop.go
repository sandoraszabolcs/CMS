package service

import (
	"context"

	"github.com/szabolcs/cms/internal/domain"
	"github.com/szabolcs/cms/internal/repository"
)

// StopLister lists stops.
type StopLister interface {
	ListStops(ctx context.Context) ([]domain.Stop, error)
}

// StopService implements StopLister.
type StopService struct {
	repo repository.StopRepository
}

func NewStopService(repo repository.StopRepository) *StopService {
	return &StopService{repo: repo}
}

func (s *StopService) ListStops(ctx context.Context) ([]domain.Stop, error) {
	return s.repo.FindAll(ctx)
}
