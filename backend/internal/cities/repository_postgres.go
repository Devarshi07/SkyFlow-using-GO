package cities

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

func (s *PostgresStore) Create(ctx context.Context, c *City) (*City, error) {
	err := s.pool.QueryRow(ctx,
		`INSERT INTO cities (name, country, code) VALUES ($1, $2, $3)
		 RETURNING id::text, name, country, code, created_at`,
		c.Name, c.Country, c.Code,
	).Scan(&c.ID, &c.Name, &c.Country, &c.Code, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *PostgresStore) GetByID(ctx context.Context, id string) (*City, bool) {
	var c City
	err := s.pool.QueryRow(ctx,
		`SELECT id::text, name, country, code, created_at FROM cities WHERE id::text = $1`, id,
	).Scan(&c.ID, &c.Name, &c.Country, &c.Code, &c.CreatedAt)
	return &c, err == nil
}

func (s *PostgresStore) List(ctx context.Context) []*City {
	rows, err := s.pool.Query(ctx, `SELECT id::text, name, country, code, created_at FROM cities ORDER BY name`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []*City
	for rows.Next() {
		var c City
		if err := rows.Scan(&c.ID, &c.Name, &c.Country, &c.Code, &c.CreatedAt); err != nil {
			continue
		}
		out = append(out, &c)
	}
	return out
}

func (s *PostgresStore) Update(ctx context.Context, id string, upd UpdateCityRequest) (*City, bool) {
	c, ok := s.GetByID(ctx, id)
	if !ok {
		return nil, false
	}
	if upd.Name != nil {
		c.Name = *upd.Name
	}
	if upd.Country != nil {
		c.Country = *upd.Country
	}
	if upd.Code != nil {
		c.Code = *upd.Code
	}
	_, err := s.pool.Exec(ctx, `UPDATE cities SET name=$1, country=$2, code=$3 WHERE id::text=$4`,
		c.Name, c.Country, c.Code, id)
	return c, err == nil
}

func (s *PostgresStore) Delete(ctx context.Context, id string) bool {
	r, err := s.pool.Exec(ctx, `DELETE FROM cities WHERE id::text = $1`, id)
	return err == nil && r.RowsAffected() > 0
}
