package flights

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, f *Flight) (*Flight, error) {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO flights (flight_number, origin_id, destination_id, departure_time, arrival_time, price, seats_total, seats_available)
		 VALUES ($1, $2::uuid, $3::uuid, $4, $5, $6, $7, $8)
		 RETURNING id::text, flight_number, origin_id::text, destination_id::text, departure_time, arrival_time, price, seats_total, seats_available, created_at`,
		f.FlightNumber, f.OriginID, f.DestinationID, f.DepartureTime, f.ArrivalTime, f.Price, f.SeatsTotal, f.SeatsTotal,
	).Scan(&f.ID, &f.FlightNumber, &f.OriginID, &f.DestinationID, &f.DepartureTime, &f.ArrivalTime, &f.Price, &f.SeatsTotal, &f.SeatsAvailable, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id string) (*Flight, bool) {
	var f Flight
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, flight_number, origin_id::text, destination_id::text, departure_time, arrival_time, price, seats_total, seats_available, created_at
		 FROM flights WHERE id::text = $1`,
		id,
	).Scan(&f.ID, &f.FlightNumber, &f.OriginID, &f.DestinationID, &f.DepartureTime, &f.ArrivalTime, &f.Price, &f.SeatsTotal, &f.SeatsAvailable, &f.CreatedAt)
	return &f, err == nil
}

func (s *PostgresStore) List(ctx context.Context) []*Flight {
	rows, err := s.pool.Query(ctx,
		`SELECT id::text, flight_number, origin_id::text, destination_id::text, departure_time, arrival_time, price, seats_total, seats_available, created_at FROM flights ORDER BY departure_time LIMIT 200`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*Flight
	for rows.Next() {
		var f Flight
		if err := rows.Scan(&f.ID, &f.FlightNumber, &f.OriginID, &f.DestinationID, &f.DepartureTime, &f.ArrivalTime, &f.Price, &f.SeatsTotal, &f.SeatsAvailable, &f.CreatedAt); err != nil {
			continue
		}
		out = append(out, &f)
	}
	return out
}

func (s *PostgresStore) Update(ctx context.Context, id string, upd UpdateFlightRequest) (*Flight, bool) {
	f, ok := s.GetByID(ctx, id)
	if !ok {
		return nil, false
	}
	if upd.FlightNumber != nil {
		f.FlightNumber = *upd.FlightNumber
	}
	if upd.OriginID != nil {
		f.OriginID = *upd.OriginID
	}
	if upd.DestinationID != nil {
		f.DestinationID = *upd.DestinationID
	}
	if upd.DepartureTime != nil {
		f.DepartureTime = *upd.DepartureTime
	}
	if upd.ArrivalTime != nil {
		f.ArrivalTime = *upd.ArrivalTime
	}
	if upd.Price != nil {
		f.Price = *upd.Price
	}
	if upd.SeatsTotal != nil {
		f.SeatsTotal = *upd.SeatsTotal
	}
	if upd.SeatsAvailable != nil {
		f.SeatsAvailable = *upd.SeatsAvailable
	}
	_, err := s.pool.Exec(ctx,
		`UPDATE flights SET flight_number=$1, origin_id=$2::uuid, destination_id=$3::uuid, departure_time=$4, arrival_time=$5, price=$6, seats_total=$7, seats_available=$8 WHERE id::text=$9`,
		f.FlightNumber, f.OriginID, f.DestinationID, f.DepartureTime, f.ArrivalTime, f.Price, f.SeatsTotal, f.SeatsAvailable, id)
	return f, err == nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) bool {
	r, err := s.pool.Exec(ctx, `DELETE FROM flights WHERE id::text = $1`, id)
	return err == nil && r.RowsAffected() > 0
}

func (s *PostgresStore) Search(ctx context.Context, origin, dest, date string) []*Flight {
	q := `SELECT id::text, flight_number, origin_id::text, destination_id::text, departure_time, arrival_time, price, seats_total, seats_available, created_at FROM flights WHERE 1=1`
	args := []interface{}{}
	n := 1
	if origin != "" {
		q += fmt.Sprintf(` AND origin_id::text = $%d`, n)
		args = append(args, origin)
		n++
	}
	if dest != "" {
		q += fmt.Sprintf(` AND destination_id::text = $%d`, n)
		args = append(args, dest)
		n++
	}
	if date != "" {
		q += fmt.Sprintf(` AND DATE(departure_time) = $%d`, n)
		args = append(args, date)
		n++
	}
	q += ` ORDER BY departure_time`

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*Flight
	for rows.Next() {
		var f Flight
		if err := rows.Scan(&f.ID, &f.FlightNumber, &f.OriginID, &f.DestinationID, &f.DepartureTime, &f.ArrivalTime, &f.Price, &f.SeatsTotal, &f.SeatsAvailable, &f.CreatedAt); err != nil {
			continue
		}
		out = append(out, &f)
	}
	return out
}

func (s *PostgresStore) SearchConnecting(ctx context.Context, origin, dest, date string) []*ConnectingFlight {
	if origin == "" || dest == "" || date == "" {
		return nil
	}

	q := `
	SELECT
		f1.id::text, f1.flight_number, f1.origin_id::text, f1.destination_id::text, f1.departure_time, f1.arrival_time, f1.price, f1.seats_total, f1.seats_available, f1.created_at,
		f2.id::text, f2.flight_number, f2.origin_id::text, f2.destination_id::text, f2.departure_time, f2.arrival_time, f2.price, f2.seats_total, f2.seats_available, f2.created_at,
		f1.destination_id::text AS layover_airport_id
	FROM flights f1
	JOIN flights f2 ON f1.destination_id = f2.origin_id
	WHERE f1.origin_id::text = $1
	  AND f2.destination_id::text = $2
	  AND DATE(f1.departure_time) = $3
	  AND f2.departure_time > f1.arrival_time + INTERVAL '45 minutes'
	  AND f2.departure_time < f1.arrival_time + INTERVAL '8 hours'
	  AND EXTRACT(EPOCH FROM (f2.arrival_time - f1.departure_time))/3600 <= 23
	  AND f1.seats_available > 0
	  AND f2.seats_available > 0
	ORDER BY (f1.price + f2.price) ASC
	LIMIT 20`

	rows, err := s.pool.Query(ctx, q, origin, dest, date)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var out []*ConnectingFlight
	for rows.Next() {
		var leg1, leg2 Flight
		var layoverAirport string
		if err := rows.Scan(
			&leg1.ID, &leg1.FlightNumber, &leg1.OriginID, &leg1.DestinationID, &leg1.DepartureTime, &leg1.ArrivalTime, &leg1.Price, &leg1.SeatsTotal, &leg1.SeatsAvailable, &leg1.CreatedAt,
			&leg2.ID, &leg2.FlightNumber, &leg2.OriginID, &leg2.DestinationID, &leg2.DepartureTime, &leg2.ArrivalTime, &leg2.Price, &leg2.SeatsTotal, &leg2.SeatsAvailable, &leg2.CreatedAt,
			&layoverAirport,
		); err != nil {
			continue
		}

		totalDur := leg2.ArrivalTime.Sub(leg1.DepartureTime).Hours()
		layoverMin := int(leg2.DepartureTime.Sub(leg1.ArrivalTime).Minutes())
		combinedPrice := leg1.Price + leg2.Price

		discount := 0
		if totalDur > 10 {
			discount = 30
		} else if totalDur >= 6 {
			discount = 15
		}

		discountedPrice := int64(math.Round(float64(combinedPrice) * (1 - float64(discount)/100)))

		out = append(out, &ConnectingFlight{
			Leg1:             &leg1,
			Leg2:             &leg2,
			TotalPrice:       discountedPrice,
			Discount:         discount,
			TotalDurationHrs: math.Round(totalDur*100) / 100,
			LayoverMinutes:   layoverMin,
			LayoverAirportID: layoverAirport,
		})
	}
	return out
}
