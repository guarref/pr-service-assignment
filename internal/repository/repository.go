package repository

import (
	"context"

	"github.com/guarref/prservice-task/internal/domain"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *domain.Team) error
	GetTeamByName(ctx context.Context, name string) (*domain.Team, error)
}

type UserRepository interface {
	SetFlagIsActive(ctx context.Context, userID string, isActive bool) error
	// CreateOrUpdateUser(ctx context.Context, user *domain.User) error
	// GetUserByID(ctx context.Context, userID string) (*domain.User, error)
}

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr *domain.PullRequest) error
	GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	UpdatePullRequest(ctx context.Context, pr *domain.PullRequest) error
	GetPullRequestByReviewerID(ctx context.Context, userID string) ([]*domain.PullRequest, error)
}

