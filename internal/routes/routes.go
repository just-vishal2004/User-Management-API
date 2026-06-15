package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/vishalkumar/ainyx-backend/internal/handler"
	"github.com/vishalkumar/ainyx-backend/internal/middleware"
)

// Setup registers all middleware and routes on the Fiber app.
// It is called once in main.go after all dependencies are initialized.
func Setup(app *fiber.App, userHandler *handler.UserHandler) {
	// ─── Global Middleware ─────────────────────────────────────────
	// These run on every request, in the order they are registered.
	// RequestID must come before Logger so the logger can read the ID.
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())

	// ─── Health Check ──────────────────────────────────────────────
	// A simple endpoint that returns 200 OK.
	// Used by Docker, Kubernetes, and load balancers to verify
	// the service is alive. Always include this in production APIs.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	// ─── API Routes ────────────────────────────────────────────────
	// All user routes are grouped under /api/v1.
	// Versioning the API (/v1/) means we can introduce /v2/ later
	// without breaking existing clients.
	api := app.Group("/api/v1")

	users := api.Group("/users")
	users.Post("/", userHandler.CreateUser)
	users.Get("/", userHandler.ListUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)
}
