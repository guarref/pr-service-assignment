package service

import (
	"context"
	"fmt"

	"github.com/guarref/pr-service-assignment/internal/errs"
	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/repository"
)

var Limit = 10

type StatsService struct {
	repo repository.StatsRepository
}

func NewStatsService(repo repository.StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetStats(ctx context.Context, top *int) (*models.Stats, error) {

	var topRew int

	if top != nil {
		if *top <= 0 {
			return nil, errs.ErrBadRequest
		}
		topRew = min(*top, Limit)
	}

	stats, err := s.repo.GetStats(ctx, topRew)
	if err != nil {
		return nil, fmt.Errorf("error getting stats: %w", err)
	}

	return stats, nil
}
