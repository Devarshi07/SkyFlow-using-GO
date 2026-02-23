package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/skyflow/skyflow/internal/airports"
	"github.com/skyflow/skyflow/internal/cities"
	"github.com/skyflow/skyflow/internal/flights"
)

type Resolver struct {
	flightSvc  *flights.Service
	citySvc    *cities.Service
	airportSvc *airports.Service
	hotCache   *HotRouteCache
}

func NewResolver(flightSvc *flights.Service, citySvc *cities.Service, airportSvc *airports.Service, hotCache *HotRouteCache) *Resolver {
	return &Resolver{
		flightSvc:  flightSvc,
		citySvc:    citySvc,
		airportSvc: airportSvc,
		hotCache:   hotCache,
	}
}

func (r *Resolver) Cities(p graphql.ResolveParams) (interface{}, error) {
	return r.citySvc.List(p.Context), nil
}

func (r *Resolver) Airports(p graphql.ResolveParams) (interface{}, error) {
	return r.airportSvc.List(p.Context), nil
}

func (r *Resolver) Flight(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)
	f, appErr := r.flightSvc.GetByID(p.Context, id)
	if appErr != nil {
		return nil, appErr
	}
	return f, nil
}

func (r *Resolver) SearchFlights(p graphql.ResolveParams) (interface{}, error) {
	origin := p.Args["origin_id"].(string)
	dest := p.Args["destination_id"].(string)
	date := p.Args["date"].(string)

	if r.hotCache != nil {
		if cached, ok := r.hotCache.Get(origin, dest, date); ok {
			return map[string]interface{}{
				"flights":    cached.Flights,
				"connecting": cached.Connecting,
				"cached":     true,
			}, nil
		}
		r.hotCache.RecordSearch(origin, dest, date)
	}

	result := r.flightSvc.SearchWithConnecting(p.Context, origin, dest, date)

	if r.hotCache != nil {
		r.hotCache.MaybeCache(origin, dest, date, result)
	}

	return map[string]interface{}{
		"flights":    result.Flights,
		"connecting": result.Connecting,
		"cached":     false,
	}, nil
}
