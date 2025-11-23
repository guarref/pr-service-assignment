package service

import (
	"context"
	"fmt"
	"unicode"

	"github.com/guarref/pr-service-assignment/internal/errs"
	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/repository"
)

type TeamService struct {
	teamRepo repository.TeamRepository
}

func NewTeamService(teamRepo repository.TeamRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo}
}

func (ts *TeamService) CreateTeam(ctx context.Context, team *models.Team) error {

	if !IsValidTeamName(team.TeamName) {
		return errs.ErrBadRequest
	}
	if len(team.Members) == 0 {
		return errs.ErrBadRequest
	}

	if err := ts.teamRepo.CreateTeam(ctx, team); err != nil {
		return fmt.Errorf("team creation error: %w", err)
	}

	return nil
}

func (ts *TeamService) GetTeamByName(ctx context.Context, teamName string) (*models.Team, error) {

	if !IsValidTeamName(teamName) {
		return nil, errs.ErrBadRequest
	}

	team, err := ts.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("error receiving team: %w", err)
	}

	return team, nil
}

func (ts *TeamService) DeactivateUsersAndReassignPRs(ctx context.Context, teamName string, userIDs []string) ([]string, error) {

	if !IsValidTeamName(teamName) {
		return nil, errs.ErrBadRequest
	}

	deactivated, err := ts.teamRepo.DeactivateUsersAndReassignPRs(ctx, teamName, userIDs)
	if err != nil {
		return nil, fmt.Errorf("error deactivating users for team %s: %w", teamName, err)
	}

	return deactivated, nil
}

func IsValidTeamName(name string) bool {

	if name == "" {
		return false
	}

	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			continue
		}
		return false
	}

	return true
}
