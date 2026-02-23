package main

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/skyflow/skyflow/internal/airports"
	"github.com/skyflow/skyflow/internal/auth"
	"github.com/skyflow/skyflow/internal/bookings"
	"github.com/skyflow/skyflow/internal/cities"
	"github.com/skyflow/skyflow/internal/customers"
	"github.com/skyflow/skyflow/internal/flights"
	gqlhandler "github.com/skyflow/skyflow/internal/graphql"
	"github.com/skyflow/skyflow/internal/payments"
	"github.com/skyflow/skyflow/internal/shared/cache"
	"github.com/skyflow/skyflow/internal/shared/config"
	"github.com/skyflow/skyflow/internal/shared/email"
	"github.com/skyflow/skyflow/internal/shared/events"
	"github.com/skyflow/skyflow/internal/shared/logger"
	mw "github.com/skyflow/skyflow/internal/shared/middleware"
	"github.com/skyflow/skyflow/internal/shared/mongo"
	"github.com/skyflow/skyflow/internal/shared/postgres"
	mqclient "github.com/skyflow/skyflow/internal/shared/rabbitmq"
	"github.com/skyflow/skyflow/internal/shared/redis"
	"github.com/skyflow/skyflow/internal/workers"
	goredis "github.com/redis/go-redis/v9"
	mongodrv "go.mongodb.org/mongo-driver/mongo"
)

func main() {
	_ = godotenv.Load()

	log := logger.Default().WithService("gateway")
	ctx := context.Background()

	dbCfg := config.DB()

	var authStore auth.Store
	var flightStore flights.FlightStore
	var cityStore cities.CityStore
	var airportStore airports.AirportStore
	var customerStore customers.CustomerStore
	var bookingStore bookings.Store

	if dbCfg.Postgres != "" {
		pool, err := postgres.NewPool(ctx, log)
		if err != nil {
			log.Fatal("postgres connect failed", "error", err)
		}
		defer pool.Close()

		authStore = auth.NewPostgresStore(pool)
		flightStore = flights.NewPostgresStore(pool)
		cityStore = cities.NewPostgresStore(pool)
		airportStore = airports.NewPostgresStore(pool)
		customerStore = customers.NewPostgresStore(pool)
		bookingStore = bookings.NewPostgresStore(pool)
		log.Info("using postgres for persistence")
	} else {
		authStore = auth.NewStore()
		flightStore = flights.NewStore()
		cityStore = cities.NewStore()
		airportStore = airports.NewStore()
		customerStore = customers.NewStore()
		bookingStore = nil
		log.Info("using in-memory stores (set DATABASE_URL for postgres)")
	}

	// Redis cache
	var redisCache *cache.RedisCache
	var rdb *goredis.Client
	if dbCfg.Redis != "" {
		var err error
		rdb, err = redis.NewClient(ctx, log)
		if err != nil {
			log.Warn("redis connect failed, disabling cache", "error", err)
		} else {
			defer rdb.Close()
			redisCache = cache.NewRedisCache(rdb, 0)
			flightStore = flights.NewCachedFlightStore(flightStore, redisCache)
			log.Info("redis cache enabled (flight search only)")
		}
	}

	// MongoDB for search logs
	var searchLogColl *mongodrv.Collection
	if dbCfg.Mongo != "" {
		mongoDB, err := mongo.NewClient(ctx, log)
		if err != nil {
			log.Warn("mongodb connect failed, search logging disabled", "error", err)
		} else {
			searchLogColl = mongoDB.Collection("flight_search_events")
			log.Info("mongodb search logging enabled")
		}
	}

	// RabbitMQ
	var mqClient *mqclient.Client
	var eventPublisher *events.Publisher
	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL != "" {
		var err error
		mqClient, err = mqclient.NewClient(rmqURL, log)
		if err != nil {
			log.Warn("rabbitmq connect failed, email notifications disabled", "error", err)
		} else {
			defer mqClient.Close()
			eventPublisher = events.NewPublisher(mqClient, log)

			// Start email worker
			emailSender := email.NewSender(log)
			emailWorker := workers.NewEmailWorker(mqClient, emailSender, log)
			if err := emailWorker.Start(); err != nil {
				log.Warn("email worker start failed", "error", err)
			}
			log.Info("rabbitmq + email worker enabled")
		}
	}

	// Payments
	paymentStripe := payments.NewStripeClient()

	// Services
	authSvc := auth.NewService(authStore)
	if eventPublisher != nil {
		authSvc.SetResetPublisher(eventPublisher)
		authSvc.SetWelcomePublisher(eventPublisher)
	}
	flightSvc := flights.NewService(flightStore, searchLogColl)
	citySvc := cities.NewService(cityStore)
	airportSvc := airports.NewService(airportStore)
	paymentSvc := payments.NewService(paymentStripe)
	customerSvc := customers.NewService(customerStore)

	// Handlers
	authHandler := auth.NewHandler(authSvc, log)
	flightHandler := flights.NewHandler(flightSvc, log)
	cityHandler := cities.NewHandler(citySvc, log)
	airportHandler := airports.NewHandler(airportSvc, log)
	paymentHandler := payments.NewHandler(paymentSvc, log)
	customerHandler := customers.NewHandler(customerSvc, log)

	// Bookings (requires postgres)
	var bookingHandler *bookings.Handler
	if bookingStore != nil {
		bookingSvc := bookings.NewService(bookingStore, flightSvc, paymentSvc)
		if eventPublisher != nil {
			bookingSvc.SetPublisher(eventPublisher)
		}
		bookingHandler = bookings.NewHandler(bookingSvc, log)
	}

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(mw.RequestID)
	r.Use(mw.Recovery(log))
	r.Use(mw.Logging(log))
	r.Use(corsMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", authHandler.Routes)
		r.Route("/flights", flightHandler.Routes)
		r.Route("/cities", cityHandler.Routes)
		r.Route("/airports", airportHandler.Routes)
		r.Route("/payments", paymentHandler.Routes)
		r.Route("/customers", customerHandler.Routes)
		if bookingHandler != nil {
			r.Route("/bookings", bookingHandler.Routes)
		}
	})

	// GraphQL endpoint
	gqlH := gqlhandler.NewHandler(flightSvc, citySvc, airportSvc, rdb, log)
	r.Handle("/graphql", gqlH)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Info("starting gateway", "addr", addr, "stripe_live", paymentStripe.IsLive())
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal("server failed", "error", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		frontendURL := os.Getenv("FRONTEND_URL")

		// Allow the configured frontend URL, localhost for dev, or any origin if not set
		allowed := "*"
		if frontendURL != "" {
			if origin == frontendURL || origin == "http://localhost:5173" || origin == "http://localhost:5176" {
				allowed = origin
			} else {
				allowed = frontendURL
			}
		}

		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}
