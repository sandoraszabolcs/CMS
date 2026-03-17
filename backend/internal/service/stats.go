package service

import (
	"context"

	"github.com/szabolcs/cms/internal/domain"
	"github.com/szabolcs/cms/internal/repository"
)

// StatsReader reads aggregated statistics.
type StatsReader interface {
	GetStats(ctx context.Context) (domain.Stats, error)
}

// StatsService implements StatsReader.
type StatsService struct {
	repo repository.StatsRepository
}

func NewStatsService(repo repository.StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetStats(ctx context.Context) (domain.Stats, error) {
	return s.repo.GetStats(ctx)
}
