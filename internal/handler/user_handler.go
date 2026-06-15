package handler

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/vishalkumar/ainyx-backend/internal/logger"
	"github.com/vishalkumar/ainyx-backend/internal/models"
	"github.com/vishalkumar/ainyx-backend/internal/repository"
	"github.com/vishalkumar/ainyx-backend/internal/service"
)

// UserHandler holds the service and validator instances.
// All HTTP handlers for user operations are methods on this struct.
type UserHandler struct {
	service  *service.UserService
	validate *validator.Validate
}

// NewUserHandler creates a new UserHandler with the given service.
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		service:  svc,
		validate: validator.New(),
	}
}

// ─── Helpers ───────────────────────────────────────────────────────────────

// parseID extracts and validates an integer ID from a URL parameter.
// Returns a 400 Bad Request if the parameter is not a valid integer.
func (h *UserHandler) parseID(c *fiber.Ctx) (int32, error) {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid user ID")
	}
	return int32(id), nil
}

// validationError formats validator errors into a readable string.
// Instead of returning raw validator internals, we build a clean
// message like "name is required; dob must be a valid date"
func validationError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		msg := ""
		for i, fe := range ve {
			if i > 0 {
				msg += "; "
			}
			msg += formatFieldError(fe)
		}
		return msg
	}
	return err.Error()
}

// formatFieldError converts a single field validation error
// into a human-readable message.
func formatFieldError(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + fe.Param() + " characters"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	case "datetime":
		return field + " must be a valid date in YYYY-MM-DD format"
	default:
		return field + " is invalid"
	}
}

// respond sends a JSON response with the given status code and body.
func respond(c *fiber.Ctx, status int, body interface{}) error {
	return c.Status(status).JSON(body)
}

// respondError sends a standardized JSON error response.
func respondError(c *fiber.Ctx, status int, message string) error {
	return respond(c, status, models.ErrorResponse{Error: message})
}

// ─── Handlers ──────────────────────────────────────────────────────────────

// CreateUser handles POST /users
// Expects: {"name": "Alice", "dob": "1990-05-10"}
// Returns: 201 Created with the created user
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest

	// Parse the JSON request body into our request struct.
	// BodyParser returns an error if the body is malformed JSON.
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, fiber.StatusBadRequest, "invalid request body")
	}

	// Validate the parsed struct against its validation tags.
	if err := h.validate.Struct(req); err != nil {
		return respondError(c, fiber.StatusBadRequest, validationError(err))
	}

	user, err := h.service.CreateUser(c.Context(), req)
	if err != nil {
		logger.Log.Error("failed to create user",
			zap.String("error", err.Error()),
		)
		return respondError(c, fiber.StatusInternalServerError, "failed to create user")
	}

	return respond(c, fiber.StatusCreated, user)
}

// GetUser handles GET /users/:id
// Returns: 200 OK with the user, or 404 if not found
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := h.parseID(c)
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "invalid user ID")
	}

	user, err := h.service.GetUser(c.Context(), id)
	if err != nil {
		// Check if the error chain contains ErrNotFound.
		// errors.Is traverses the chain created by fmt.Errorf("%w").
		if errors.Is(err, repository.ErrNotFound) {
			return respondError(c, fiber.StatusNotFound, "user not found")
		}
		logger.Log.Error("failed to get user",
			zap.Int32("user_id", id),
			zap.String("error", err.Error()),
		)
		return respondError(c, fiber.StatusInternalServerError, "failed to get user")
	}

	return respond(c, fiber.StatusOK, user)
}

// UpdateUser handles PUT /users/:id
// Expects: {"name": "Alice Updated", "dob": "1991-03-15"}
// Returns: 200 OK with the updated user, or 404 if not found
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := h.parseID(c)
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "invalid user ID")
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return respondError(c, fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.validate.Struct(req); err != nil {
		return respondError(c, fiber.StatusBadRequest, validationError(err))
	}

	user, err := h.service.UpdateUser(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return respondError(c, fiber.StatusNotFound, "user not found")
		}
		logger.Log.Error("failed to update user",
			zap.Int32("user_id", id),
			zap.String("error", err.Error()),
		)
		return respondError(c, fiber.StatusInternalServerError, "failed to update user")
	}

	return respond(c, fiber.StatusOK, user)
}

// DeleteUser handles DELETE /users/:id
// Returns: 204 No Content on success, or 404 if not found
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := h.parseID(c)
	if err != nil {
		return respondError(c, fiber.StatusBadRequest, "invalid user ID")
	}

	err = h.service.DeleteUser(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return respondError(c, fiber.StatusNotFound, "user not found")
		}
		logger.Log.Error("failed to delete user",
			zap.Int32("user_id", id),
			zap.String("error", err.Error()),
		)
		return respondError(c, fiber.StatusInternalServerError, "failed to delete user")
	}

	// 204 No Content — success with no response body.
	// This is the correct HTTP status for a successful DELETE.
	return c.SendStatus(fiber.StatusNoContent)
}

// ListUsers handles GET /users
// Supports pagination: ?page=1&page_size=10
// Returns: 200 OK with paginated users and metadata
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	var query models.PaginationQuery

	// QueryParser reads ?page=1&page_size=10 into the struct
	// using the `query:` struct tags we defined in models.
	if err := c.QueryParser(&query); err != nil {
		return respondError(c, fiber.StatusBadRequest, "invalid query parameters")
	}

	result, err := h.service.ListUsers(c.Context(), query)
	if err != nil {
		logger.Log.Error("failed to list users",
			zap.String("error", err.Error()),
		)
		return respondError(c, fiber.StatusInternalServerError, "failed to list users")
	}

	return respond(c, fiber.StatusOK, result)
}
