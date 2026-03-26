package service

import (
	"context"

	"github.com/szabolcs/cms/internal/repository"
)

// Resetter resets simulation data.
type Resetter interface {
	Reset(ctx context.Context) error
}

// ResetService implements Resetter.
type ResetService struct {
	validations repository.ValidationRepository
	odMatrix    repository.ODMatrixRepository
}

func NewResetService(
	validations repository.ValidationRepository,
	odMatrix repository.ODMatrixRepository,
) *ResetService {
	return &ResetService{
		validations: validations,
		odMatrix:    odMatrix,
	}
}

func (s *ResetService) Reset(ctx context.Context) error {
	if err := s.validations.DeleteAll(ctx); err != nil {
		return err
	}
	return s.odMatrix.Refresh(ctx)
}
