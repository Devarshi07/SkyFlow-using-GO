package graphql

import (
	"github.com/graphql-go/graphql"
)

var cityType = graphql.NewObject(graphql.ObjectConfig{
	Name: "City",
	Fields: graphql.Fields{
		"id":      &graphql.Field{Type: graphql.String},
		"name":    &graphql.Field{Type: graphql.String},
		"country": &graphql.Field{Type: graphql.String},
		"code":    &graphql.Field{Type: graphql.String},
	},
})

var airportType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Airport",
	Fields: graphql.Fields{
		"id":      &graphql.Field{Type: graphql.String},
		"name":    &graphql.Field{Type: graphql.String},
		"city_id": &graphql.Field{Type: graphql.String},
		"code":    &graphql.Field{Type: graphql.String},
	},
})

var flightType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Flight",
	Fields: graphql.Fields{
		"id":              &graphql.Field{Type: graphql.String},
		"flight_number":   &graphql.Field{Type: graphql.String},
		"origin_id":       &graphql.Field{Type: graphql.String},
		"destination_id":  &graphql.Field{Type: graphql.String},
		"departure_time":  &graphql.Field{Type: graphql.String},
		"arrival_time":    &graphql.Field{Type: graphql.String},
		"price":           &graphql.Field{Type: graphql.Int},
		"seats_total":     &graphql.Field{Type: graphql.Int},
		"seats_available": &graphql.Field{Type: graphql.Int},
	},
})

var connectingFlightType = graphql.NewObject(graphql.ObjectConfig{
	Name: "ConnectingFlight",
	Fields: graphql.Fields{
		"leg1":                 &graphql.Field{Type: flightType},
		"leg2":                 &graphql.Field{Type: flightType},
		"total_price":          &graphql.Field{Type: graphql.Int},
		"discount":             &graphql.Field{Type: graphql.Int},
		"total_duration_hours": &graphql.Field{Type: graphql.Float},
		"layover_minutes":      &graphql.Field{Type: graphql.Int},
		"layover_airport_id":   &graphql.Field{Type: graphql.String},
	},
})

var searchResultType = graphql.NewObject(graphql.ObjectConfig{
	Name: "SearchResult",
	Fields: graphql.Fields{
		"flights":    &graphql.Field{Type: graphql.NewList(flightType)},
		"connecting": &graphql.Field{Type: graphql.NewList(connectingFlightType)},
		"cached":     &graphql.Field{Type: graphql.Boolean},
	},
})

func buildSchema(r *Resolver) (graphql.Schema, error) {
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"cities": &graphql.Field{
				Type:    graphql.NewList(cityType),
				Resolve: r.Cities,
			},
			"airports": &graphql.Field{
				Type:    graphql.NewList(airportType),
				Resolve: r.Airports,
			},
			"searchFlights": &graphql.Field{
				Type: searchResultType,
				Args: graphql.FieldConfigArgument{
					"origin_id":      &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"destination_id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"date":           &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: r.SearchFlights,
			},
			"flight": &graphql.Field{
				Type: flightType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: r.Flight,
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
}
