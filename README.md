# SkyFlow – Flight Booking API

Go microservices backend for a flight booking system.
It may take time to load the cities as I am using the Azure 0 replica stratergy to keep the cost to 0 so please bear with it for 3 - 4 mins. Also the API might feel slow (based on location) because the Free tier didnt have any servers in US which was available in my Free Tier plan. You will get the Email and everything it will just take 30 - 50 seconds thats all.

## Structure

```
SkyFlow/
├── backend/          # Go API backend
│   ├── cmd/gateway/
│   ├── internal/
│   ├── migrations/
│   └── ...
└── README.md
```

## Quick Start

```bash
cd backend
go run ./cmd/gateway
```

API base URL: `http://localhost:8080`

## Database Configuration

Set these environment variables (or use `.env`). Database details will be provided separately.

### Migrations

Run migrations in order (with PostgreSQL running):

```bash
psql $DATABASE_URL -f backend/migrations/001_init.sql
psql $DATABASE_URL -f backend/migrations/002_bookings.sql
psql $DATABASE_URL -f backend/migrations/003_user_profile.sql
psql $DATABASE_URL -f backend/migrations/004_seed_data.sql
psql $DATABASE_URL -f backend/migrations/005_us_cities_airports.sql  # US routes for frontend
```

Or via Docker: `docker exec -i <postgres_container> psql -U skyflow -d skyflow < backend/migrations/005_us_cities_airports.sql`

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string |
| `MONGODB_URI` | MongoDB connection string |
| `REDIS_URL` | Redis connection string |
| `STRIPE_SECRET_KEY` | Stripe API key (optional) |

- **All empty** → In-memory mode (no persistence)
- **All set** → PostgreSQL + Redis cache + MongoDB event logs

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` – User login
- `POST /api/v1/auth/register` – User registration
- `POST /api/v1/auth/refresh` – Refresh token
- `POST /api/v1/auth/logout` – User logout

### Flights
- `GET /api/v1/flights` – List flights
- `GET /api/v1/flights/:id` – Get flight
- `POST /api/v1/flights` – Create flight
- `PUT /api/v1/flights/:id` – Update flight
- `DELETE /api/v1/flights/:id` – Delete flight
- `GET /api/v1/flights/search?origin_id=&destination_id=&date=` – Search flights

### Cities, Airports
- Full CRUD at `/api/v1/cities`, `/api/v1/airports`

### Payments
- `POST /api/v1/payments/intent` – Create payment intent
- `POST /api/v1/payments/:id/confirm` – Confirm payment
- `GET /api/v1/payments/:id` – Get payment
- `POST /api/v1/payments/:id/refund` – Refund
- `POST /api/v1/payments/:id/cancel` – Cancel
- `GET /api/v1/payments/methods` – Supported methods

### Customers
- `POST /api/v1/customers` – Create customer
- `GET /api/v1/customers/:id` – Get customer
- `GET /api/v1/customers/:id/payments` – Payment history

## Google OAuth (Sign in with Google)

### Setup

1. **Backend `.env`** – set:
   ```
   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   ```

2. **Google Cloud Console** – [Credentials](https://console.cloud.google.com/apis/credentials) → your OAuth 2.0 Client ID → **Authorized redirect URIs**:
   - `http://localhost:5173/login` (add your production URL when deploying)

3. **Run OAuth schema migration** (if using Postgres):
   ```bash
   make migrate-oauth
   # or: psql $DATABASE_URL -f backend/migrations/006_google_oauth_schema.sql
   ```

**If you get "Internal server error"** after Google sign-in:

1. **Run the schema migration** (required for Postgres):
   ```bash
   psql $DATABASE_URL -f backend/migrations/006_google_oauth_schema.sql
   ```
   Or run 001, 003, then 006.

2. **Set `DEBUG=1`** in `backend/.env` and restart the backend – the UI will show the actual error.

3. **Ensure env vars** in `backend/.env`:
   ```
   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   ```

4. **No database?** Unset `DATABASE_URL` to use in-memory storage – Google OAuth will work without Postgres.
