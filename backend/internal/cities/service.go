package cities

import (
	"context"

	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
)

type Service struct {
	store CityStore
}

func NewService(store CityStore) *Service {
	return &Service{store: store}
}

func (s *Service) List(ctx context.Context) []*City {
	return s.store.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*City, *apperrors.AppError) {
	c, ok := s.store.GetByID(ctx, id)
	if !ok {
		return nil, apperrors.NotFound("city")
	}
	return c, nil
}

func (s *Service) Create(ctx context.Context, req CreateCityRequest) (*City, *apperrors.AppError) {
	if req.Name == "" || req.Country == "" {
		return nil, apperrors.BadRequest("name and country required")
	}
	c := &City{Name: req.Name, Country: req.Country, Code: req.Code}
	out, err := s.store.Create(ctx, c)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateCityRequest) (*City, *apperrors.AppError) {
	c, ok := s.store.Update(ctx, id, req)
	if !ok {
		return nil, apperrors.NotFound("city")
	}
	return c, nil
}

func (s *Service) Delete(ctx context.Context, id string) *apperrors.AppError {
	if !s.store.Delete(ctx, id) {
		return apperrors.NotFound("city")
	}
	return nil
}
