package service

import (
	"context"

	"github.com/szabolcs/cms/internal/domain"
	"github.com/szabolcs/cms/internal/repository"
)

// VehicleLister lists vehicles.
type VehicleLister interface {
	ListVehicles(ctx context.Context) ([]domain.Vehicle, error)
}

// VehicleService implements VehicleLister.
type VehicleService struct {
	repo repository.VehicleRepository
}

func NewVehicleService(repo repository.VehicleRepository) *VehicleService {
	return &VehicleService{repo: repo}
}

func (s *VehicleService) ListVehicles(ctx context.Context) ([]domain.Vehicle, error) {
	return s.repo.FindAll(ctx)
}
