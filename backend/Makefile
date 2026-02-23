.PHONY: build run run-dev test docker migrate

build:
	go build -o bin/gateway ./cmd/gateway

run: build
	./bin/gateway

run-dev:
	go run ./cmd/gateway

test:
	go test ./...

# Run DB migrations (requires DATABASE_URL)
migrate:
	@test -n "$$DATABASE_URL" || (echo "Set DATABASE_URL" && exit 1)
	for f in migrations/*.sql; do echo "Running $$f..."; psql $$DATABASE_URL -f $$f; done

# Fix Google OAuth (run if sign-in fails with Internal server error)
migrate-oauth:
	@test -n "$$DATABASE_URL" || (echo "Set DATABASE_URL" && exit 1)
	psql $$DATABASE_URL -f migrations/006_google_oauth_schema.sql

docker:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down
