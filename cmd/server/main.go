package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/example/userapi/config"
	"github.com/example/userapi/internal/handler"
	"github.com/example/userapi/internal/logger"
	"github.com/example/userapi/internal/repository"
	"github.com/example/userapi/internal/routes"
	"github.com/example/userapi/internal/service"
)

func main() {
	cfg := config.Load()

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	// Database connection pool.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("unable to create connection pool", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("unable to reach database", zap.Error(err))
	}
	log.Info("connected to database")

	// Dependency wiring.
	repo := repository.NewUserRepository(pool)
	svc := service.NewUserService(repo, log)
	h := handler.NewUserHandler(svc, log)

	app := fiber.New(fiber.Config{
		ErrorHandler:          routes.ErrorHandler,
		DisableStartupMessage: true,
	})
	routes.Register(app, h, log)

	// Run server in a goroutine so we can handle graceful shutdown.
	go func() {
		addr := ":" + cfg.AppPort
		log.Info("starting server", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil && !errors.Is(err, fiber.ErrServiceUnavailable) {
			log.Fatal("server stopped unexpectedly", zap.Error(err))
		}
	}()

	// Wait for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}
	log.Info("server exited cleanly")
}
