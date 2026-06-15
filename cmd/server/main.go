package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/vishalkumar/ainyx-backend/config"
	"github.com/vishalkumar/ainyx-backend/internal/handler"
	"github.com/vishalkumar/ainyx-backend/internal/logger"
	"github.com/vishalkumar/ainyx-backend/internal/repository"
	"github.com/vishalkumar/ainyx-backend/internal/routes"
	"github.com/vishalkumar/ainyx-backend/internal/service"
)

func main() {
	// ─── Step 1: Load Configuration ───────────────────────────────
	// Read .env file and environment variables into a Config struct.
	// Everything else receives values from this struct — nothing
	// reads environment variables directly.
	cfg := config.Load()

	// ─── Step 2: Initialize Logger ────────────────────────────────
	// Set up Uber Zap before anything else so all subsequent
	// startup steps can log what they're doing.
	logger.Init(cfg.AppEnv)

	// Sync flushes buffered log entries when the app exits.
	// defer ensures this runs even if main() panics.
	defer logger.Sync()

	logger.Log.Info("starting ainyx backend",
		zap.String("env", cfg.AppEnv),
		zap.String("port", cfg.AppPort),
	)

	// ─── Step 3: Connect to Database ──────────────────────────────
	// pgxpool.New creates a connection pool — not a single connection.
	// The pool manages multiple connections automatically,
	// reusing them across concurrent requests.
	pool, err := pgxpool.New(context.Background(), cfg.DBUrl)
	if err != nil {
		logger.Log.Fatal("failed to connect to database",
			zap.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer pool.Close()

	// Ping the database to verify the connection is actually alive.
	// pgxpool.New succeeds even if the DB is unreachable — Ping confirms it.
	if err := pool.Ping(context.Background()); err != nil {
		logger.Log.Fatal("failed to ping database",
			zap.String("error", err.Error()),
		)
		os.Exit(1)
	}

	logger.Log.Info("connected to database successfully")

	// ─── Step 4: Wire Dependencies ────────────────────────────────
	// This is the composition root — we create each layer and inject
	// the layer below it. The dependency chain flows downward:
	// handler → service → repository → database pool
	userRepo := repository.NewUserRepository(pool)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// ─── Step 5: Create Fiber App ─────────────────────────────────
	app := fiber.New(fiber.Config{
		// Custom error handler ensures all errors return
		// our standard {"error": "..."} JSON shape,
		// even unhandled panics or 404s.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "internal server error"

			// Check if it's a Fiber-specific error (has a status code).
			var fiberErr *fiber.Error
			if ok := fiber.ErrUnprocessableEntity; ok != nil {
				if e, ok := err.(*fiber.Error); ok {
					code = e.Code
					msg = e.Message
				}
			}
			_ = fiberErr

			return c.Status(code).JSON(fiber.Map{
				"error": msg,
			})
		},
	})

	// ─── Step 6: Register Routes ──────────────────────────────────
	routes.Setup(app, userHandler)

	// ─── Step 7: Graceful Shutdown ────────────────────────────────
	// Run the server in a goroutine so it doesn't block.
	// We listen for OS signals (Ctrl+C, kill) in the main goroutine
	// and shut down cleanly when one arrives.
	go func() {
		addr := fmt.Sprintf(":%s", cfg.AppPort)
		logger.Log.Info("server listening", zap.String("addr", addr))

		if err := app.Listen(addr); err != nil {
			logger.Log.Error("server error", zap.String("error", err.Error()))
		}
	}()

	// Block until we receive SIGINT (Ctrl+C) or SIGTERM (kill/Docker stop).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("shutting down server...")

	// app.Shutdown() waits for in-flight requests to complete
	// before stopping — no requests are dropped mid-flight.
	if err := app.Shutdown(); err != nil {
		logger.Log.Error("error during shutdown", zap.String("error", err.Error()))
	}

	logger.Log.Info("server stopped cleanly")
}
