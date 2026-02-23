package airports

import "context"

type AirportStore interface {
	Create(ctx context.Context, a *Airport) (*Airport, error)
	GetByID(ctx context.Context, id string) (*Airport, bool)
	List(ctx context.Context) []*Airport
	Update(ctx context.Context, id string, upd UpdateAirportRequest) (*Airport, bool)
	Delete(ctx context.Context, id string) bool
}
