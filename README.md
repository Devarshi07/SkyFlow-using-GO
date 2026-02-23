# SkyFlow – Flight Booking Platform

A modern, scalable, and cost-effective flight booking system built with a **Go microservices** backend and **React + TypeScript** frontend.

🚀 **Live Demo**:  
https://sky-flow-using-go-nsn8.vercel.app/

> ⚠️ **Note to Recruiters & Reviewers**  
> To keep this project 100% free to host, the Go backend is deployed on Azure with replicas set to `0` (scale-to-zero).  
> The **first API request may take 2–3 minutes** to wake the container.  
> Once running, subsequent requests are fast and responsive.  
> Thank you for your patience!

---

## 📖 Overview

**SkyFlow** is a full-stack flight booking platform demonstrating:

- Microservices architecture in Go  
- REST API design using Chi router  
- PostgreSQL relational modeling  
- Redis caching  
- RabbitMQ event-driven communication  
- Stripe payment integration  
- Google OAuth authentication  
- Dockerized local development  
- Cloud-native, zero-cost deployment strategy  

The backend handles booking workflows, payments, and async email notifications.  
The frontend provides a smooth, real-time flight search and booking experience.

---

# 🐳 Run Locally (Fastest Way)

The easiest way to run the full project without installing Go, Node.js, PostgreSQL, Redis, or RabbitMQ manually is using **Docker**.

---

## ✅ Prerequisites

- Install **Docker Desktop**
- Ensure Docker is running

---

## 1️⃣ Clone the Repository

```bash
git clone https://github.com/SomyaPadhy4501/SkyFlow-using-GO.git
cd SkyFlow-using-GO
```

---

## 2️⃣ Set Environment Variables (Optional)

Create a `.env` file in the root directory:

```env
# Database
DATABASE_URL=postgresql://skyflow:password@postgres:5432/skyflow?sslmode=disable

# Redis
REDIS_URL=redis://redis:6379

# RabbitMQ
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/

# OAuth (Optional for local testing)
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret

# Payments (Optional)
STRIPE_SECRET_KEY=sk_test_...
```

---

## 3️⃣ Start Everything

```bash
docker-compose up --build
```

That’s it 🚀

The entire stack will start:

- Go API  
- PostgreSQL  
- Redis  
- RabbitMQ  
- Frontend  

---

# ☁️ Zero-Cost Cloud Architecture

This system was intentionally architected to demonstrate a **production-ready distributed system using free-tier cloud services.**

| Component     | Technology              | Hosting Provider | Why Chosen |
|--------------|------------------------|------------------|------------|
| Frontend     | React + TypeScript     | Vercel           | Zero-config CI/CD + global CDN |
| Backend API  | Go 1.21+ (Chi Router)  | Azure Container Apps | Scale-to-zero to eliminate idle costs |
| Database     | PostgreSQL             | Neon Serverless  | Generous free tier + pooling |
| Caching      | Redis                  | Upstash          | Serverless, pay-per-request |
| Message Queue| RabbitMQ               | CloudAMQP        | Lightweight async processing |

---

# 🏗️ System Architecture

## 🔁 Data Flow

```text
Frontend (React + TS on Vercel)
        ↓
API Gateway (Go + Chi on Azure)
        ↓
Microservices:
   - Auth
   - Flights
   - Bookings
   - Payments
        ↓
PostgreSQL (Neon) + Redis (Upstash)
        ↓
RabbitMQ
        ↓
Background Workers (Email Processing)
```

---

# 🧠 Architecture Highlights

- Clean separation of services  
- Event-driven async workflows  
- Serverless database + caching  
- Scale-to-zero backend to minimize cost  
- Dockerized full-stack setup  
- Production-ready system design principles  

---

# 🎯 Why This Project Stands Out

This isn’t just a CRUD app.

It demonstrates:

- Real microservices separation  
- Cloud-native deployment strategy  
- Cost-optimized infrastructure decisions  
- DevOps awareness  
- Async processing patterns  
- Production-level architecture thinking  

Perfect for backend, distributed systems, and cloud engineering roles.
