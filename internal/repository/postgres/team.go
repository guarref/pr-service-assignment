package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/errors"
	"github.com/jmoiron/sqlx"
)

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// если есть команда, то TEAM_EXISTS, если нет, то создаем
// если есть пользователь, то обновляем данные, если нет, то создаем
func (tr *TeamRepository) CreateTeam(ctx context.Context, team *models.Team) error {

	tx, err := tr.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error begining transaction team: %w", err)
	}
	defer tx.Rollback()

	var isExists bool
	checkTeamQuery := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`

	if err := tx.GetContext(ctx, &isExists, checkTeamQuery, team.TeamName); err != nil {
		return fmt.Errorf("error checking for team existence: %w", err)
	}
	if isExists {
		return errors.ErrTeamExists
	}

	creationTeamQuery := `INSERT INTO teams (team_name, created_at, updated_at) VALUES ($1, NOW(), NOW())`

	if _, err := tx.ExecContext(ctx, creationTeamQuery, team.TeamName); err != nil {
		return fmt.Errorf("error team creation: %w", err)
	}

	additionUserQuery := `INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW()) ON CONFLICT (user_id) 
		DO UPDATE SET
		username   = EXCLUDED.username,
		team_name  = EXCLUDED.team_name,
		is_active  = EXCLUDED.is_active,
		updated_at = NOW()`

	for _, m := range team.Members {
		if _, err := tx.ExecContext(ctx, additionUserQuery, m.UserID, m.UserName, team.TeamName, m.IsActive); err != nil {
			return fmt.Errorf("error addition user with id %s: %w", m.UserID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction team: %w", err)
	}

	return nil
}

func (tr *TeamRepository) GetTeamByName(ctx context.Context, teamName string) (*models.Team, error) {

	var team models.Team

	teamQuery := `SELECT team_name, created_at, updated_at
		FROM teams
		WHERE team_name = $1`

	if err := tr.db.GetContext(ctx, &team, teamQuery, teamName); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("error gettig team: %w", err)
	}

	membersQuery := `SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY username`

	var members []models.TeamMember
	if err := tr.db.SelectContext(ctx, &members, membersQuery, teamName); err != nil {
		return nil, fmt.Errorf("error getting team members: %w", err)
	}
	team.Members = members

	return &team, nil
}
