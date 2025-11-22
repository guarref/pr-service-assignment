package service

import (
	"context"
	"fmt"
	"unicode"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/repository"
	"github.com/guarref/pr-service-assignment/internal/errors"
)

type TeamService struct {
	teamRepo repository.TeamRepository
}

func NewTeamService(teamRepo repository.TeamRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo}
}

func (ts *TeamService) CreateTeam(ctx context.Context, team *models.Team) error {

	if !IsValidTeamName(team.TeamName) {
		return errors.ErrBadRequest
	}
	if len(team.Members) == 0 {
		return errors.ErrBadRequest
	}

	if err := ts.teamRepo.CreateTeam(ctx, team); err != nil {
		return fmt.Errorf("team creation error: %w", err)
	}

	return nil
}

func (ts *TeamService) GetTeamByName(ctx context.Context, teamName string) (*models.Team, error) {

	if !IsValidTeamName(teamName) {
		return nil, errors.ErrBadRequest
	}

	team, err := ts.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("error receiving team: %w", err)
	}

	return team, nil
}

func IsValidTeamName(name string) bool {

	if name == "" {
		return false
	}
	for _, let := range name {
		if !unicode.IsLetter(let) {
			return false
		}
	}

	return true
}
