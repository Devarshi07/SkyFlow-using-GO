package config

import (
	"os"
)

// DB holds database connection strings from environment.
// When unset, returns "" so callers can skip DB setup (in-memory mode).
func DB() struct {
	Postgres string
	Mongo    string
	Redis    string
} {
	return struct {
		Postgres string
		Mongo    string
		Redis    string
	}{
		Postgres: os.Getenv("DATABASE_URL"),
		Mongo:    os.Getenv("MONGODB_URI"),
		Redis:    os.Getenv("REDIS_URL"),
	}
}

