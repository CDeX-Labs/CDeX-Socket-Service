package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CDeX-Labs/CDeX-Socket-Service/config"
	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/auth"
	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/handlers"
	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/hub"
	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/kafka"
	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/metrics"
	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/middleware"
	"github.com/CDeX-Labs/CDeX-Socket-Service/internal/presence"
	redisclient "github.com/CDeX-Labs/CDeX-Socket-Service/internal/redis"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()

	cfg := config.Load()

	if cfg.Server.Env == "development" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	logger.Info().
		Str("port", cfg.Server.Port).
		Str("env", cfg.Server.Env).
		Str("version", "1.0.0").
		Msg("Starting Socket Service")

	if cfg.JWT.Secret == "" {
		logger.Fatal().Msg("JWT_SECRET is required")
	}

	redisClient, err := redisclient.NewClient(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redisClient.Close()

	appMetrics := metrics.New()

	wsHub := hub.NewHub(logger)
	go wsHub.Run()

	jwtValidator := auth.NewJWTValidator(cfg.JWT.Secret)

	redisPubSub := redisclient.NewPubSub(redisClient, func(envelope *redisclient.PubSubEnvelope) {
		if envelope.TargetRoom != "" {
			wsHub.SendToRoom(envelope.TargetRoom, envelope.Message)
		} else if envelope.TargetUser != "" {
			wsHub.SendToUser(envelope.TargetUser, envelope.Message)
		} else {
			wsHub.Broadcast(envelope.Message)
		}
	}, logger)

	if err := redisPubSub.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start Redis PubSub")
	}
	defer redisPubSub.Stop()

	presenceManager := presence.NewManager(redisClient, redisPubSub.GetInstanceID(), logger)

	kafkaConsumer := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.ConsumerGroup,
		cfg.Kafka.Topics,
		logger,
	)

	kafkaHandlers := kafka.NewHandlers(wsHub, logger)
	kafkaHandlers.RegisterAll(kafkaConsumer)
	kafkaConsumer.Start()
	defer kafkaConsumer.Stop()

	wsHandler := handlers.NewWebSocketHandler(wsHub, presenceManager, logger)

	rateLimiter := middleware.NewRateLimiter(100, time.Minute, logger)

	mux := http.NewServeMux()
	mux.Handle("/ws", auth.AuthMiddleware(jwtValidator)(wsHandler))
	mux.HandleFunc("/health", handlers.HealthHandler())
	mux.HandleFunc("/ready", handlers.ReadyHandler(wsHub))

	var handler http.Handler = mux
	handler = middleware.CORS(middleware.DefaultCORSConfig())(handler)
	handler = rateLimiter.Middleware(handler)
	handler = middleware.Recovery(logger)(handler)
	handler = middleware.Logging(logger)(handler)

	if cfg.Metrics.Enabled {
		go func() {
			metricsMux := http.NewServeMux()
			metricsMux.Handle("/metrics", promhttp.Handler())
			metricsServer := &http.Server{
				Addr:    ":" + cfg.Metrics.Port,
				Handler: metricsMux,
			}
			logger.Info().Str("port", cfg.Metrics.Port).Msg("Metrics server started")
			if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
				logger.Error().Err(err).Msg("Metrics server error")
			}
		}()
	}

	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		logger.Info().Str("port", cfg.Server.Port).Msg("WebSocket server started")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info().Str("signal", sig.String()).Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server stopped gracefully")

	_ = appMetrics
}
