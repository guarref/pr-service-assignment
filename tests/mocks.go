package tests

import (
	"context"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/repository"
	"github.com/stretchr/testify/mock"
)

// MockTeamRepository is a mock implementation of TeamRepository
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) CreateTeam(ctx context.Context, team *models.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetTeamByName(ctx context.Context, name string) (*models.Team, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) DeactivateUsersAndReassignPRs(ctx context.Context, teamName string, userIDs []string) ([]string, error) {
	args := m.Called(ctx, teamName, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) SetFlagIsActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	args := m.Called(ctx, userID, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetActiveUsersByTeam(ctx context.Context, teamName string, exceptUserID string) ([]*models.User, error) {
	args := m.Called(ctx, teamName, exceptUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

// MockPullRequestRepository is a mock implementation of PullRequestRepository
type MockPullRequestRepository struct {
	mock.Mock
}

func (m *MockPullRequestRepository) CreatePullRequest(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error) {
	args := m.Called(ctx, pr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) MergePullRequestByID(ctx context.Context, prID string) (*models.PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) ReassignToPullRequest(ctx context.Context, prID string, oldUserID string) (*models.PullRequest, string, error) {
	args := m.Called(ctx, prID, oldUserID)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*models.PullRequest), args.String(1), args.Error(2)
}

func (m *MockPullRequestRepository) GetPullRequestByReviewerID(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PullRequestShort), args.Error(1)
}

// MockStatsRepository is a mock implementation of StatsRepository
type MockStatsRepository struct {
	mock.Mock
}

func (m *MockStatsRepository) GetStats(ctx context.Context, top int) (*models.Stats, error) {
	args := m.Called(ctx, top)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stats), args.Error(1)
}

// Verify that mocks implement interfaces
var (
	_ repository.TeamRepository        = (*MockTeamRepository)(nil)
	_ repository.UserRepository        = (*MockUserRepository)(nil)
	_ repository.PullRequestRepository  = (*MockPullRequestRepository)(nil)
	_ repository.StatsRepository        = (*MockStatsRepository)(nil)
)

