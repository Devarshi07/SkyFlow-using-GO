package airports

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, a *Airport) (*Airport, error) {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO airports (name, city_id, code) VALUES ($1, $2::uuid, $3)
		 RETURNING id::text, name, city_id::text, code, created_at`,
		a.Name, a.CityID, a.Code,
	).Scan(&a.ID, &a.Name, &a.CityID, &a.Code, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id string) (*Airport, bool) {
	var a Airport
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, name, city_id::text, code, created_at FROM airports WHERE id::text = $1`, id,
	).Scan(&a.ID, &a.Name, &a.CityID, &a.Code, &a.CreatedAt)
	return &a, err == nil
}

func (s *PostgresStore) List(ctx context.Context) []*Airport {
	rows, err := s.pool.Query(ctx, `SELECT id::text, name, city_id::text, code, created_at FROM airports ORDER BY name`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*Airport
	for rows.Next() {
		var a Airport
		if err := rows.Scan(&a.ID, &a.Name, &a.CityID, &a.Code, &a.CreatedAt); err != nil {
			continue
		}
		out = append(out, &a)
	}
	return out
}

func (s *PostgresStore) Update(ctx context.Context, id string, upd UpdateAirportRequest) (*Airport, bool) {
	a, ok := s.GetByID(ctx, id)
	if !ok {
		return nil, false
	}
	if upd.Name != nil {
		a.Name = *upd.Name
	}
	if upd.CityID != nil {
		a.CityID = *upd.CityID
	}
	if upd.Code != nil {
		a.Code = *upd.Code
	}
	_, err := s.pool.Exec(ctx, `UPDATE airports SET name=$1, city_id=$2::uuid, code=$3 WHERE id::text=$4`,
		a.Name, a.CityID, a.Code, id)
	return a, err == nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) bool {
	r, err := s.pool.Exec(ctx, `DELETE FROM airports WHERE id::text = $1`, id)
	return err == nil && r.RowsAffected() > 0
}
