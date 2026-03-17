package service

import (
	"context"

	"github.com/szabolcs/cms/internal/domain"
	"github.com/szabolcs/cms/internal/repository"
)

// ODMatrixReader reads the OD matrix.
type ODMatrixReader interface {
	GetODMatrix(ctx context.Context) ([]domain.ODMatrixRow, error)
}

// ODMatrixService implements ODMatrixReader.
type ODMatrixService struct {
	repo repository.ODMatrixRepository
}

func NewODMatrixService(repo repository.ODMatrixRepository) *ODMatrixService {
	return &ODMatrixService{repo: repo}
}

func (s *ODMatrixService) GetODMatrix(ctx context.Context) ([]domain.ODMatrixRow, error) {
	return s.repo.GetAll(ctx)
}
