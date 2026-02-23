package events

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type FlightSearchEvent struct {
	OriginID      string    `bson:"origin_id"`
	DestinationID string    `bson:"destination_id"`
	Date          string    `bson:"date"`
	ResultsCount  int       `bson:"results_count"`
	Timestamp     time.Time `bson:"timestamp"`
}

// LogFlightSearch writes a search event to MongoDB (fire-and-forget, no error returned to caller)
func LogFlightSearch(ctx context.Context, coll *mongo.Collection, origin, dest, date string, count int) {
	doc := FlightSearchEvent{
		OriginID:      origin,
		DestinationID: dest,
		Date:          date,
		ResultsCount:  count,
		Timestamp:     time.Now(),
	}
	_, _ = coll.InsertOne(ctx, doc)
}
