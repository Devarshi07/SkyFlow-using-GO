package flights

import (
	"context"

	apperrors "github.com/skyflow/skyflow/internal/shared/errors"
	"github.com/skyflow/skyflow/internal/shared/events"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	store     FlightStore
	searchLog *mongo.Collection
}

func NewService(store FlightStore, searchLog *mongo.Collection) *Service {
	return &Service{store: store, searchLog: searchLog}
}

func (s *Service) List(ctx context.Context) []*Flight {
	return s.store.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Flight, *apperrors.AppError) {
	f, ok := s.store.GetByID(ctx, id)
	if !ok {
		return nil, apperrors.NotFound("flight")
	}
	return f, nil
}

func (s *Service) Create(ctx context.Context, req CreateFlightRequest) (*Flight, *apperrors.AppError) {
	if req.FlightNumber == "" || req.OriginID == "" || req.DestinationID == "" {
		return nil, apperrors.BadRequest("flight_number, origin_id, destination_id required")
	}
	if req.SeatsTotal <= 0 {
		req.SeatsTotal = 100
	}
	f := &Flight{
		FlightNumber:  req.FlightNumber,
		OriginID:      req.OriginID,
		DestinationID: req.DestinationID,
		DepartureTime: req.DepartureTime,
		ArrivalTime:   req.ArrivalTime,
		Price:         req.Price,
		SeatsTotal:    req.SeatsTotal,
	}
	out, err := s.store.Create(ctx, f)
	if err != nil {
		return nil, apperrors.Internal(err)
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateFlightRequest) (*Flight, *apperrors.AppError) {
	f, ok := s.store.Update(ctx, id, req)
	if !ok {
		return nil, apperrors.NotFound("flight")
	}
	return f, nil
}

func (s *Service) Delete(ctx context.Context, id string) *apperrors.AppError {
	if !s.store.Delete(ctx, id) {
		return apperrors.NotFound("flight")
	}
	return nil
}

func (s *Service) Search(ctx context.Context, origin, dest, date string) []*Flight {
	results := s.store.Search(ctx, origin, dest, date)
	if s.searchLog != nil {
		events.LogFlightSearch(ctx, s.searchLog, origin, dest, date, len(results))
	}
	return results
}

func (s *Service) SearchWithConnecting(ctx context.Context, origin, dest, date string) *SearchResult {
	direct := s.Search(ctx, origin, dest, date)
	connecting := s.store.SearchConnecting(ctx, origin, dest, date)
	if direct == nil {
		direct = []*Flight{}
	}
	if connecting == nil {
		connecting = []*ConnectingFlight{}
	}
	return &SearchResult{
		Flights:    direct,
		Connecting: connecting,
	}
}

func (s *Service) Store() FlightStore {
	return s.store
}
