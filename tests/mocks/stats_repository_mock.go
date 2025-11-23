package mocks

import (
	"context"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/stretchr/testify/mock"
)

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

