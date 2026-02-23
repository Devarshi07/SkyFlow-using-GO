package airports

import (
	"context"

	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
)

type Service struct {
	store AirportStore
}

func NewService(store AirportStore) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context) []*Airport {
	return s.store.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Airport, *apperrors.AppError) {
	a, ok := s.store.GetByID(ctx, id)
	if !ok {
		return nil, apperrors.NotFound("airport")
	}
	return a, nil
}

func (s *Service) Create(ctx context.Context, req CreateAirportRequest) (*Airport, *apperrors.AppError) {
	if req.Name == "" || req.CityID == "" {
		return nil, apperrors.BadRequest("name and city_id required")
	}
	a := &Airport{Name: req.Name, CityID: req.CityID, Code: req.Code}
	out, err := s.store.Create(ctx, a)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateAirportRequest) (*Airport, *apperrors.AppError) {
	a, ok := s.store.Update(ctx, id, req)
	if !ok {
		return nil, apperrors.NotFound("airport")
	}
	return a, nil
}

func (s *Service) Delete(ctx context.Context, id string) *apperrors.AppError {
	if !s.store.Delete(ctx, id) {
		return apperrors.NotFound("airport")
	}
	return nil
}
