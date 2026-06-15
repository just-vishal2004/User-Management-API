package service

import (
	"context"
	"fmt"
	"math"
	"time"

	db "github.com/vishalkumar/ainyx-backend/db/sqlc/generated"
	"github.com/vishalkumar/ainyx-backend/internal/models"
	"github.com/vishalkumar/ainyx-backend/internal/repository"
)

// UserService contains all business logic for user operations.
// It depends on the repository for data access but never touches
// the database directly.
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService creates a new UserService with the given repository.
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// ─── Age Calculation ───────────────────────────────────────────────────────

// CalculateAge returns the current age in years for a given date of birth.
// It uses time.Now() as the reference point.
// Exported so it can be called from handlers or other packages if needed.
func CalculateAge(dob time.Time) int {
	return calculateAgeAt(dob, time.Now())
}

// calculateAgeAt calculates age relative to a given reference time.
// Unexported — used internally and in tests for deterministic results.
// Separating this from CalculateAge lets us test with a fixed "today"
// without time.Now() making test results change over real time.
func calculateAgeAt(dob, now time.Time) int {
	years := now.Year() - dob.Year()

	birthdayThisYear := time.Date(
		now.Year(),
		dob.Month(),
		dob.Day(),
		0, 0, 0, 0,
		now.Location(),
	)

	if now.Before(birthdayThisYear) {
		years--
	}

	return years
}

// ─── Response Builder ──────────────────────────────────────────────────────

// toUserResponse converts a SQLC-generated db.User into a models.UserResponse.
// This is where age is calculated and dob is formatted as a string.
func toUserResponse(user db.User) models.UserResponse {
	return models.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		Dob:  models.FormatDob(user.Dob),
		Age:  CalculateAge(user.Dob),
	}
}

// ─── Service Methods ───────────────────────────────────────────────────────

// CreateUser validates the request, parses dob, and creates a new user.
func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	// Parse the dob string into time.Time.
	// The layout "2006-01-02" is Go's reference time — always this exact date.
	dob, err := time.Parse("2006-01-02", req.Dob)
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("invalid date format: %w", err)
	}

	// Ensure dob is not in the future.
	if dob.After(time.Now()) {
		return models.UserResponse{}, fmt.Errorf("date of birth cannot be in the future")
	}

	user, err := s.repo.Create(ctx, db.CreateUserParams{
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("service.CreateUser: %w", err)
	}

	return toUserResponse(user), nil
}

// GetUser fetches a single user by id and returns them with calculated age.
func (s *UserService) GetUser(ctx context.Context, id int32) (models.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("service.GetUser: %w", err)
	}
	return toUserResponse(user), nil
}

// UpdateUser updates an existing user's name and dob.
func (s *UserService) UpdateUser(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error) {
	dob, err := time.Parse("2006-01-02", req.Dob)
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("invalid date format: %w", err)
	}

	if dob.After(time.Now()) {
		return models.UserResponse{}, fmt.Errorf("date of birth cannot be in the future")
	}

	user, err := s.repo.Update(ctx, db.UpdateUserParams{
		ID:   id,
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		return models.UserResponse{}, fmt.Errorf("service.UpdateUser: %w", err)
	}

	return toUserResponse(user), nil
}

// DeleteUser removes a user by id.
func (s *UserService) DeleteUser(ctx context.Context, id int32) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("service.DeleteUser: %w", err)
	}
	return nil
}

// ListUsers returns a paginated list of users with metadata.
func (s *UserService) ListUsers(ctx context.Context, query models.PaginationQuery) (models.PaginatedResponse, error) {
	// Apply defaults and enforce limits on pagination parameters.
	page, pageSize := normalizePagination(query)

	// Calculate SQL OFFSET from page number.
	// Page 1 = offset 0, Page 2 = offset pageSize, etc.
	offset := (page - 1) * pageSize

	users, err := s.repo.List(ctx, db.ListUsersParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return models.PaginatedResponse{}, fmt.Errorf("service.ListUsers: %w", err)
	}

	// Get total count for pagination metadata.
	total, err := s.repo.Count(ctx)
	if err != nil {
		return models.PaginatedResponse{}, fmt.Errorf("service.ListUsers count: %w", err)
	}

	// Convert each db.User to models.UserResponse with calculated age.
	responses := make([]models.UserResponse, len(users))
	for i, user := range users {
		responses[i] = toUserResponse(user)
	}

	// Calculate total number of pages.
	// math.Ceil ensures we round up — 11 users with page size 10 = 2 pages.
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return models.PaginatedResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ─── Helpers ───────────────────────────────────────────────────────────────

// normalizePagination applies defaults and enforces max limits
// on pagination parameters to prevent abuse.
func normalizePagination(query models.PaginationQuery) (page, pageSize int) {
	page = query.Page
	pageSize = query.PageSize

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 10
	}

	// Cap page size at 100 to prevent clients from
	// requesting thousands of rows in one request.
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}
