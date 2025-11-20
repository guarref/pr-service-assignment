package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/guarref/pr-service-assignment/internal/domain"
	"github.com/guarref/pr-service-assignment/internal/resperrors"
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
func (tr *TeamRepository) CreateTeam(ctx context.Context, team *domain.Team) error {
	tx, err := tr.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begining transaction team: %w", err)
	}
	defer tx.Rollback()

	var isExists bool
	checkTeamQuery := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	if err := tx.GetContext(ctx, &isExists, checkTeamQuery, team.TeamName); err != nil {
		return fmt.Errorf("checking for team existence: %w", err)
	}
	if isExists {
		return resperrors.ErrTeamExists
	}

	// создание команды
	creationTeamQuery := `INSERT INTO teams (team_name, created_at, updated_at) VALUES ($1, NOW(), NOW())`
	if _, err := tx.ExecContext(ctx, creationTeamQuery, team.TeamName); err != nil {
		return fmt.Errorf("team creation: %w", err)
	}

	// добавление юзеров
	additionUserQuery := `INSERT INTO users (user_id, username, team_name, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW()) ON CONFLICT (user_id) DO UPDATE SET
			username   = EXCLUDED.username,
			team_name  = EXCLUDED.team_name,
			is_active  = EXCLUDED.is_active,
			updated_at = NOW()`

	for _, m := range team.Members {
		if _, err := tx.ExecContext(ctx, additionUserQuery, m.UserID, m.UserName, team.TeamName, m.IsActive); err != nil {
			return fmt.Errorf("addition user %s: %w", m.UserID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction team: %w", err)
	}

	return nil
}

func (tr *TeamRepository) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	var team domain.Team

	teamQuery := `SELECT team_name, created_at, updated_at
		FROM teams
		WHERE team_name = $1`

	if err := tr.db.GetContext(ctx, &team, teamQuery, teamName); err != nil {
		if err == sql.ErrNoRows {
			return nil, resperrors.ErrTeamNotFound
		}
		return nil, fmt.Errorf("get team: %w", err)
	}

	membersQuery := `SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY username`

	var members []domain.TeamMember
	if err := tr.db.SelectContext(ctx, &members, membersQuery, teamName); err != nil {
		return nil, fmt.Errorf("get team members: %w", err)
	}
	team.Members = members

	return &team, nil
}
