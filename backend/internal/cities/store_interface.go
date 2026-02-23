package cities

import "context"

type CityStore interface {
	Create(ctx context.Context, c *City) (*City, error)
	GetByID(ctx context.Context, id string) (*City, bool)
	List(ctx context.Context) []*City
	Update(ctx context.Context, id string, upd UpdateCityRequest) (*City, bool)
	Delete(ctx context.Context, id string) bool
}
