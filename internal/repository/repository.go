package repository

import (
	"context"

	"github.com/guarref/pr-service-assignment/internal/domain"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *domain.Team) error
	GetTeamByName(ctx context.Context, name string) (*domain.Team, error)
}

type UserRepository interface {
	SetFlagIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamName string, exceptUserID string) ([]*domain.User, error)
}

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr *domain.PullRequest) error
	MergePullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	ReassignToPullRequest(ctx context.Context, prID string, oldUserID string) (*domain.PullRequest, string, error)
	GetPullRequestByReviewerID(ctx context.Context, userID string) ([]*domain.PullRequestShort, error)
}
