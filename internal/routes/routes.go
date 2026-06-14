package routes

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/asimar007/userapi/internal/handler"
	"github.com/asimar007/userapi/internal/middleware"
)

// Register attaches all routes and middleware to the Fiber app.
func Register(app *fiber.App, h *handler.UserHandler, log *zap.Logger) {
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(log))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	users := app.Group("/users")
	users.Post("/", h.Create)
	users.Get("/", h.List)
	users.Get("/:id", h.Get)
	users.Put("/:id", h.Update)
	users.Delete("/:id", h.Delete)
}

// ErrorHandler is a centralized Fiber error handler that renders a consistent
// JSON error envelope and preserves explicit fiber status codes.
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "internal server error"

	var fe *fiber.Error
	if e, ok := err.(*fiber.Error); ok {
		fe = e
		code = e.Code
		msg = e.Message
	}
	_ = fe

	return c.Status(code).JSON(fiber.Map{
		"error":  msg,
		"status": code,
	})
}
