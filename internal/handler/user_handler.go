package handler

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/asimar007/userapi/internal/models"
	"github.com/asimar007/userapi/internal/service"
)

// UserHandler wires HTTP requests to the user service.
type UserHandler struct {
	svc      *service.UserService
	validate *validator.Validate
	log      *zap.Logger
}

// NewUserHandler builds a UserHandler.
func NewUserHandler(svc *service.UserService, log *zap.Logger) *UserHandler {
	return &UserHandler{
		svc:      svc,
		validate: validator.New(),
		log:      log,
	}
}

// Create handles POST /users.
func (h *UserHandler) Create(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationError(err)
	}
	res, err := h.svc.Create(c.Context(), req.Name, req.DOB)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create user")
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

// Get handles GET /users/:id.
func (h *UserHandler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	res, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not fetch user")
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

// List handles GET /users with optional pagination (?page=&limit=).
func (h *UserHandler) List(c *fiber.Ctx) error {
	page := parseQueryInt(c, "page", 1)
	limit := parseQueryInt(c, "limit", 20)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	users, total, err := h.svc.List(c.Context(), int32(limit), int32(offset))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list users")
	}

	// Expose pagination metadata via headers while keeping the documented
	// response body shape (a plain JSON array of users).
	c.Set("X-Total-Count", strconv.FormatInt(total, 10))
	c.Set("X-Page", strconv.Itoa(page))
	c.Set("X-Limit", strconv.Itoa(limit))
	return c.Status(fiber.StatusOK).JSON(users)
}

// Update handles PUT /users/:id.
func (h *UserHandler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationError(err)
	}
	res, err := h.svc.Update(c.Context(), id, req.Name, req.DOB)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "could not update user")
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

// Delete handles DELETE /users/:id.
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete user")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// --- helpers ---

func parseID(c *fiber.Ctx) (int32, error) {
	raw := c.Params("id")
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || v <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid user id")
	}
	return int32(v), nil
}

func parseQueryInt(c *fiber.Ctx, key string, def int) int {
	raw := c.Query(key)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return v
}

func validationError(err error) error {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) && len(verrs) > 0 {
		fe := verrs[0]
		msg := "validation failed on field '" + fe.Field() + "' (rule: " + fe.Tag() + ")"
		return fiber.NewError(fiber.StatusBadRequest, msg)
	}
	return fiber.NewError(fiber.StatusBadRequest, "validation failed")
}
