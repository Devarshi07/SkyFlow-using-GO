package flights

import "context"

type FlightStore interface {
	Create(ctx context.Context, f *Flight) (*Flight, error)
	GetByID(ctx context.Context, id string) (*Flight, bool)
	List(ctx context.Context) []*Flight
	Update(ctx context.Context, id string, upd UpdateFlightRequest) (*Flight, bool)
	Delete(ctx context.Context, id string) bool
	Search(ctx context.Context, origin, dest, date string) []*Flight
	SearchConnecting(ctx context.Context, origin, dest, date string) []*ConnectingFlight
}
