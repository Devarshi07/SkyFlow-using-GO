package graphql

import (
	"net/http"

	"github.com/graphql-go/handler"
	"github.com/redis/go-redis/v9"
	"github.com/skyflow/skyflow/internal/airports"
	"github.com/skyflow/skyflow/internal/cities"
	"github.com/skyflow/skyflow/internal/flights"
	"github.com/skyflow/skyflow/internal/shared/logger"
)

func NewHandler(
	flightSvc *flights.Service,
	citySvc *cities.Service,
	airportSvc *airports.Service,
	rdb *redis.Client,
	log *logger.Logger,
) http.Handler {
	hotCache := NewHotRouteCache(rdb)
	resolver := NewResolver(flightSvc, citySvc, airportSvc, hotCache)

	schema, err := buildSchema(resolver)
	if err != nil {
		log.Fatal("failed to build graphql schema", "error", err)
	}

	return handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
}
