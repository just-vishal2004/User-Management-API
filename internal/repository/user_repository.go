package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/vishalkumar/ainyx-backend/db/sqlc/generated"
)

// ErrNotFound is returned when a requested resource does not exist.
// We define our own error here so the service layer can check for it
// without knowing anything about pgx or SQL internals.
var ErrNotFound = errors.New("record not found")

// UserRepository handles all database operations for users.
// It wraps the SQLC-generated Queries struct and translates
// database-level errors into application-level errors.
type UserRepository struct {
	queries *db.Queries
}

// NewUserRepository creates a new UserRepository.
// It takes a *pgxpool.Pool — a connection pool that manages
// multiple database connections efficiently under load.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		queries: db.New(pool),
	}
}

// Create inserts a new user into the database.
// Returns the created user including the auto-generated id.
func (r *UserRepository) Create(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	user, err := r.queries.CreateUser(ctx, arg)
	if err != nil {
		return db.User{}, fmt.Errorf("repository.Create: %w", err)
	}
	return user, nil
}

// GetByID fetches a single user by their id.
// Returns ErrNotFound if no user exists with that id.
func (r *UserRepository) GetByID(ctx context.Context, id int32) (db.User, error) {
	user, err := r.queries.GetUser(ctx, id)
	if err != nil {
		// pgx returns pgx.ErrNoRows when a query returns no results.
		// We translate this to our own ErrNotFound so the service
		// layer never needs to import pgx just to check this error.
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrNotFound
		}
		return db.User{}, fmt.Errorf("repository.GetByID: %w", err)
	}
	return user, nil
}

// Update modifies an existing user's name and dob.
// Returns ErrNotFound if no user exists with that id.
func (r *UserRepository) Update(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	user, err := r.queries.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, ErrNotFound
		}
		return db.User{}, fmt.Errorf("repository.Update: %w", err)
	}
	return user, nil
}

// Delete removes a user by their id.
// Returns ErrNotFound if no user exists with that id.
func (r *UserRepository) Delete(ctx context.Context, id int32) error {
	err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("repository.Delete: %w", err)
	}
	return nil
}

// List fetches a paginated list of users.
// offset = (page - 1) * pageSize, calculated in the service layer.
func (r *UserRepository) List(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	users, err := r.queries.ListUsers(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("repository.List: %w", err)
	}
	return users, nil
}

// Count returns the total number of users in the database.
// Used by the service layer to calculate total pages for pagination.
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("repository.Count: %w", err)
	}
	return count, nil
}
