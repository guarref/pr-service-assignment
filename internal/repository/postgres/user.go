package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/guarref/pr-service-assignment/internal/domain"
	"github.com/guarref/pr-service-assignment/internal/resperrors"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) SetFlagIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	
	flagQuery := `UPDATE users
		SET is_active = $2, updated_at = NOW()
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active, created_at, updated_at`

	var user domain.User
	err := ur.db.GetContext(ctx, &user, flagQuery, userID, isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, resperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("error updating is_active field: %w", err)
	}

	return &user, nil
}

func (ur *UserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	
	query := `SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE user_id = $1`

	var user domain.User
	if err := ur.db.GetContext(ctx, &user, query, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, resperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &user, nil
}

func (ur *UserRepository) GetActiveUsersByTeam(ctx context.Context, teamName string, exceptUserID string) ([]*domain.User, error) {
	
	query := `SELECT user_id, username, team_name, is_active, created_at, updated_at
		FROM users
		WHERE team_name = $1 AND is_active = true AND user_id != $2
		ORDER BY user_id`

	var users []*domain.User
	if err := ur.db.SelectContext(ctx, &users, query, teamName, exceptUserID); err != nil {
		return nil, fmt.Errorf("error getting active users: %w", err)
	}

	return users, nil
}
