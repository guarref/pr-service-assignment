package mocks

import (
	"context"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/stretchr/testify/mock"
)

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
	return args.Get(0).(*models.PullRequest), args.Get(1).(string), args.Error(2)
}

func (m *MockPullRequestRepository) GetPullRequestByReviewerID(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.PullRequestShort), args.Error(1)
}

