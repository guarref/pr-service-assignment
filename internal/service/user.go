package service

import (
	"context"
	"fmt"

	"github.com/guarref/pr-service-assignment/internal/domain"
	"github.com/guarref/pr-service-assignment/internal/repository"
	"github.com/guarref/pr-service-assignment/internal/resperrors"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (us *UserService) SetFlagIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {

	if userID == "" {
		return nil, resperrors.ErrBadRequest
	}

	user, err := us.userRepo.SetFlagIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, fmt.Errorf("error setting flag is_active for user %s: %w", userID, err)
	}

	return user, nil
}

func (us *UserService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {

	if userID == "" {
		return nil, resperrors.ErrBadRequest
	}

	user, err := us.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user by user_id %s: %w", userID, err)
	}

	return user, nil
}

func (us *UserService) GetActiveUsersByTeam(ctx context.Context, teamName string, exceptUserID string) ([]*domain.User, error) {

	if !IsValidTeamName(teamName) {
		return nil, resperrors.ErrBadRequest
	}

	users, err := us.userRepo.GetActiveUsersByTeam(ctx, teamName, exceptUserID)
	if err != nil {
		return nil, fmt.Errorf("error getting active users by team %s: %w", teamName, err)
	}

	return users, nil
}
