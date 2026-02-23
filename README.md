# SkyFlow – Flight Booking Platform

A modern, scalable flight booking system built with **Go microservices** backend and **React TypeScript** frontend.

🚀 **Live Demo**: [https://sky-flow-using-go-nsn8.vercel.app/](https://sky-flow-using-go-nsn8.vercel.app/)

> **⚠️ Note**: The application uses free tier deployments with optimized cold start strategies. Initial load may take 3-4 minutes. Subsequent requests respond normally.

---

## Overview

**SkyFlow** is a full-stack flight booking platform that demonstrates modern cloud architecture, microservices patterns, and infrastructure-as-code practices. The backend handles complex booking workflows, payment processing, and event-driven email notifications, while the frontend provides a seamless user experience with real-time flight search.

---

## 🏗️ Architecture

### Backend Stack
- **Runtime**: Go 1.21+
- **API Framework**: Chi Router (HTTP/REST)
- **GraphQL**: GraphQL support with hot caching
- **Databases**:
  - **PostgreSQL**: Primary relational database (users, flights, bookings)
  - **MongoDB**: Event logs and audit trails
  - **Redis**: Distributed caching layer for flights, cities, airports
- **Message Queue**: RabbitMQ for asynchronous email workers
- **Authentication**: JWT + Google OAuth 2.0

### Frontend Stack
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **HTTP Client**: Apollo Client + GraphQL
- **Styling**: CSS

### Data Flow
```
Frontend (React/TS)
    ↓
API Gateway (Chi Router)
    ↓
Microservices (Auth, Flights, Bookings, Payments, etc.)
    ↓
PostgreSQL (Primary Data) + Redis (Cache) + MongoDB (Events)
    ↓
Workers (Email via RabbitMQ)
```

---

## 🚀 Deployment & Infrastructure

### Why Free Tier Services?
This project showcases cost-effective deployment strategies suitable for startups and learning:

| Service | Provider | Reason |
|---------|----------|--------|
| **Frontend** | Vercel | Zero-config deployment, automatic optimizations, global CDN |
| **Backend API** | Azure Container Instances | Flexible free tier, easy scaling, no cold start penalties |
| **Database** | Azure Database for PostgreSQL | Managed service, automated backups, free tier available |
| **Cache** | Redis Cloud | Fast in-memory cache, free tier supports 30MB |
| **Message Queue** | RabbitMQ (Self-hosted) | Open-source, lightweight for event streaming |
| **MongoDB** | MongoDB Atlas | Managed service, free tier with 512MB storage |

### Infrastructure as Code (Terraform)
All cloud infrastructure is defined in **Terraform** (`infra/`):
- Database provisioning
- Container instance configuration
- Redis cluster setup
- Network policies and security groups

Benefits:
- **Reproducibility**: Entire infrastructure can be recreated from code
- **Version Control**: Track infrastructure changes with git
- **Consistency**: Same setup across dev, staging, and production

```bash
cd infra
terraform init
terraform plan
terraform apply
```

---

## 📋 API Endpoints

### Authentication (`/api/v1/auth`)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | User registration |
| POST | `/auth/login` | User login (returns JWT) |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/logout` | User logout |
| POST | `/auth/oauth/google` | Sign in with Google |

### Flights (`/api/v1/flights`)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/flights` | List all flights (cached) |
| GET | `/flights/:id` | Get flight details |
| GET | `/flights/search` | Search flights by route & date |
| POST | `/flights` | Create flight (admin only) |
| PUT | `/flights/:id` | Update flight (admin only) |
| DELETE | `/flights/:id` | Delete flight (admin only) |

### Bookings (`/api/v1/bookings`)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/bookings` | Create new booking |
| GET | `/bookings/:id` | Get booking details |
| GET | `/bookings` | List user bookings |
| PUT | `/bookings/:id` | Update booking |
| DELETE | `/bookings/:id` | Cancel booking |

### Payments (`/api/v1/payments`)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/payments/intent` | Create payment intent (Stripe) |
| POST | `/payments/:id/confirm` | Confirm payment |
| GET | `/payments/:id` | Get payment status |
| POST | `/payments/:id/refund` | Refund payment |
| POST | `/payments/:id/cancel` | Cancel payment |

### Cities & Airports (`/api/v1`)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/cities` | List cities (cached) |
| GET | `/cities/:id` | Get city details |
| GET | `/airports` | List airports (cached) |
| GET | `/airports/:id` | Get airport details |

### Customers (`/api/v1/customers`)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/customers` | Create customer profile |
| GET | `/customers/:id` | Get customer details |
| GET | `/customers/:id/payments` | Payment history |

---

## 🔧 Quick Start

### Local Development

**Prerequisites**: Go 1.21+, Node.js 18+, Docker, PostgreSQL

```bash
# 1. Clone repository
git clone https://github.com/SomyaPadhy4501/SkyFlow-using-GO
cd SkyFlow

# 2. Start backend
cd backend
go run ./cmd/gateway

# 3. Start frontend (in another terminal)
cd frontend
npm install
npm run dev
```

**API Base URL**: `http://localhost:8080`  
**Frontend**: `http://localhost:5173`

### Database Setup

Set environment variables:
```bash
export DATABASE_URL="postgresql://user:password@localhost:5432/skyflow"
export MONGODB_URI="mongodb://localhost:27017"
export REDIS_URL="redis://localhost:6379"
export GOOGLE_CLIENT_ID="your-client-id.apps.googleusercontent.com"
export GOOGLE_CLIENT_SECRET="your-client-secret"
export STRIPE_SECRET_KEY="sk_test_..."
```

Run migrations:
```bash
psql $DATABASE_URL -f backend/migrations/001_init.sql
psql $DATABASE_URL -f backend/migrations/002_bookings.sql
psql $DATABASE_URL -f backend/migrations/003_user_profile.sql
psql $DATABASE_URL -f backend/migrations/004_seed_data.sql
psql $DATABASE_URL -f backend/migrations/005_us_cities_airports.sql
psql $DATABASE_URL -f backend/migrations/006_google_oauth_schema.sql
```

Or via Docker:
```bash
docker-compose up -d
docker exec -i skyflow_postgres psql -U skyflow -d skyflow < backend/migrations/001_init.sql
```

---

## 🔐 Authentication

### Google OAuth Setup

1. Create OAuth 2.0 credentials in [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Set authorized redirect URIs:
   - Local: `http://localhost:5173/login`
   - Production: `https://sky-flow-using-go-nsn8.vercel.app/login`

3. Add to `.env`:
   ```
   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   ```

4. Run OAuth schema migration:
   ```bash
   psql $DATABASE_URL -f backend/migrations/006_google_oauth_schema.sql
   ```

---

## 📦 Project Structure

```
SkyFlow/
├── backend/
│   ├── cmd/gateway/          # API entry point
│   ├── internal/
│   │   ├── auth/             # Authentication logic
│   │   ├── flights/          # Flight management
│   │   ├── bookings/         # Booking operations
│   │   ├── payments/         # Payment processing
│   │   ├── graphql/          # GraphQL resolvers
│   │   └── shared/           # Shared utilities
│   ├── migrations/           # Database schemas
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── pages/            # React pages
│   │   ├── components/       # Reusable components
│   │   ├── api/              # API clients
│   │   └── context/          # Auth context
│   ├── vite.config.ts
│   └── package.json
├── infra/
│   ├── main.tf               # Terraform configuration
│   ├── variables.tf
│   └── terraform.tfvars
└── docker-compose.yml        # Local development setup
```

---

## 📊 Technology Highlights

- **Microservices Architecture**: Modular, independent services for scalability
- **Caching Strategy**: Redis integration for flights, cities, and airports
- **Event-Driven**: RabbitMQ for asynchronous email notifications
- **GraphQL**: Modern API alternative to REST
- **Infrastructure as Code**: Full Terraform configuration
- **CI/CD Ready**: Docker containerization for easy deployment
- **Type-Safe**: Go interfaces and TypeScript for robust code
