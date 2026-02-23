#!/usr/bin/env bash
# SkyFlow full startup script
set -e

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

echo "=== 1. Starting databases (PostgreSQL, Redis, RabbitMQ) ==="
docker compose up -d

echo ""
echo "=== 2. Waiting for PostgreSQL to be ready ==="
sleep 3
PG_CONTAINER=$(docker compose ps -q postgres)
until docker exec "$PG_CONTAINER" pg_isready -U skyflow -d skyflow 2>/dev/null; do
  echo "Waiting for PostgreSQL..."
  sleep 2
done

echo ""
echo "=== 3. Running migrations ==="
cd "$ROOT/backend"
export DATABASE_URL="postgres://skyflow:skyflow@localhost:5433/skyflow?sslmode=disable"
for f in migrations/*.sql; do
  echo "Running $f..."
  docker exec -i "$PG_CONTAINER" psql -U skyflow -d skyflow < "$f" 2>/dev/null || true
done

echo ""
echo "=== 4. Starting backend (http://localhost:8080) ==="
cd "$ROOT/backend"
export DATABASE_URL="postgres://skyflow:skyflow@localhost:5433/skyflow?sslmode=disable"
export REDIS_URL="redis://localhost:6380"
export RABBITMQ_URL="amqp://skyflow:skyflow@localhost:5672/"
go run ./cmd/gateway &
BACKEND_PID=$!

echo ""
echo "=== 5. Starting frontend (http://localhost:5173) ==="
cd "$ROOT/frontend"
npm run dev &
FRONTEND_PID=$!

echo ""
echo "=== SkyFlow is starting ==="
echo "  Backend:  http://localhost:8080"
echo "  Frontend: http://localhost:5173"
echo "  RabbitMQ: http://localhost:15672 (skyflow/skyflow)"
echo ""
echo "Press Ctrl+C to stop all services"
trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT TERM
wait
