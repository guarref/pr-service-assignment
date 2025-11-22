package repository

import (
	"context"

	"github.com/guarref/pr-service-assignment/internal/models"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *models.Team) error
	GetTeamByName(ctx context.Context, name string) (*models.Team, error)
}

type UserRepository interface {
	SetFlagIsActive(ctx context.Context, userID string, isActive bool) (*models.User, error)
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetActiveUsersByTeam(ctx context.Context, teamName string, exceptUserID string) ([]*models.User, error)
}

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr *models.PullRequest) error
	MergePullRequestByID(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignToPullRequest(ctx context.Context, prID string, oldUserID string) (*models.PullRequest, string, error)
	GetPullRequestByReviewerID(ctx context.Context, userID string) ([]*models.PullRequestShort, error)
}
