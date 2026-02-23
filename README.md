# SkyFlow – Flight Booking Platform

A modern, scalable, and cost-effective flight booking system built with a **Go microservices** backend and **React TypeScript** frontend.

🚀 **Live Demo**: [https://sky-flow-using-go-nsn8.vercel.app/](https://sky-flow-using-go-nsn8.vercel.app/)

> **⚠️ Note to Recruiters & Reviewers**: To keep this project 100% free to host, the Go backend is deployed on Azure with replicas set to `0` (scale-to-zero). **The initial load or API request may take 2-3 minutes to wake the server.** Once awake, subsequent requests will be fast and highly responsive. Thank you for your patience!

---

## 📖 Overview

**SkyFlow** is a full-stack flight booking platform that demonstrates modern cloud architecture, microservices patterns, and infrastructure-as-code practices. The backend handles complex booking workflows, payment processing, and event-driven email notifications, while the frontend provides a seamless user experience with real-time flight search.

---
## Run this locally to get the best speed
The easiest way to run this project locally without installing Go, Node.js, or configuring any databases is by using **Docker**. This allows you to spin up the entire full-stack environment with a single command.

### Prerequisites
* [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed and running on your machine.

### 1. Clone the Repository
```bash
git clone [https://github.com/SomyaPadhy4501/SkyFlow-using-GO](https://github.com/SomyaPadhy4501/SkyFlow-using-GO)
cd SkyFlow

# Local Docker Environment Variables
DATABASE_URL="postgresql://skyflow:password@postgres:5432/skyflow?sslmode=disable"
REDIS_URL="redis://redis:6379"
RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/"

# OAuth & Payments (Optional for basic local viewing)
GOOGLE_CLIENT_ID="your-client-id.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="your-client-secret"
STRIPE_SECRET_KEY="sk_test_..."

docker-compose up --build
___

## ☁️ Zero-Cost Cloud Architecture

This project was intentionally architected to demonstrate how to build and deploy a production-ready, distributed system using entirely free-tier cloud services. 

Here is how the infrastructure is broken down:

| Component | Technology | Hosting Provider | Why I Chose It |
|-----------|------------|------------------|----------------|
| **Frontend** | React + TypeScript | **Vercel** | Zero-config deployment, global CDN, and automatic CI/CD from GitHub. |
| **Backend API** | Go 1.21+ (Chi Router) | **Azure (Container Apps/Instances)** | Configured to scale-to-zero (`replicas: 0`) to avoid recurring cloud costs while demonstrating containerized deployment. |
| **Database** | PostgreSQL | **Neon Serverless Postgres** | Generous free tier for relational data, easy connection pooling, and automated backups. |
| **Caching** | Redis | **Upstash Serverless Redis** | Purpose-built for serverless environments. Its per-request pricing and scale-to-zero nature perfectly complement the Azure backend without incurring idle costs. |
| **Message Queue** | RabbitMQ | **CloudAMQP (or self-hosted)** | Lightweight event streaming for asynchronous background workers (like email processing). |

---


## 🏗️ System Architecture

### Data Flow
```text
Frontend (React/TS on Vercel)
    ↓
API Gateway (Go / Chi Router on Azure)
    ↓
Microservices (Auth, Flights, Bookings, Payments)
    ↓
PostgreSQL (Neon) + Redis (Upstash) 
    ↓
Workers (Asynchronous Email via RabbitMQ)

## 🐳 Running Locally (Recruiter & Reviewer Friendly)
