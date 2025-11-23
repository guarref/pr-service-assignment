package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/guarref/pr-service-assignment/internal/errs"
	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
		return errs.ErrTeamExists
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
			return nil, errs.ErrTeamNotFound
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

func (r *TeamRepository) DeactivateUsersAndReassignPRs(ctx context.Context, teamName string, userIDs []string) ([]string, error) {

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error begining transaction team: %w", err)
	}
	defer tx.Rollback()

	var isExists bool
	checkTeamQuery := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	if err := tx.GetContext(ctx, &isExists, checkTeamQuery, teamName); err != nil {
		return nil, fmt.Errorf("error checking for team existence: %w", err)
	}
	if !isExists {
		return nil, errs.ErrNotFound
	}

	var usersToDeactivate []string

	// выбор конкретных пользователей для деактивации
	if len(userIDs) > 0 {
		deactivateQuery := `SELECT user_id 
			FROM users 
			WHERE team_name = $1 
			AND user_id = ANY($2) 
			AND is_active = true`

		if err := tx.SelectContext(ctx, &usersToDeactivate, deactivateQuery, teamName, pq.Array(userIDs)); err != nil {
			return nil, fmt.Errorf("error getting users to deactivate: %w", err)
		}
	} else {
		deactivateQuery := `SELECT user_id 
			FROM users 
			WHERE team_name = $1 
			AND is_active = true`

		if err := tx.SelectContext(ctx, &usersToDeactivate, deactivateQuery, teamName); err != nil {
			return nil, fmt.Errorf("error getting all team users: %w", err)
		}
	}

	if len(usersToDeactivate) == 0 {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("error committing transaction team: %w", err)
		}
		return []string{}, nil
	}

	type prInfo struct {
		PullRequestID string `db:"pull_request_id"`
		AuthorID      string `db:"author_id"`
	}

	affectedPRsQuery := `SELECT DISTINCT pr.pull_request_id, pr.author_id
		FROM pull_requests pr
		INNER JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE pr.status = 'OPEN' 
		AND rev.user_id = ANY($1)`

	var affectedPRs []prInfo
	if err := tx.SelectContext(ctx, &affectedPRs, affectedPRsQuery, pq.Array(usersToDeactivate)); err != nil {
		return nil, fmt.Errorf("error getting affected PRs: %w", err)
	}

	for _, pr := range affectedPRs {
		var currentReviewers []string
		reviewersQuery := `SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1`
		if err := tx.SelectContext(ctx, &currentReviewers, reviewersQuery, pr.PullRequestID); err != nil {
			return nil, fmt.Errorf("error getting current reviewers for PR %s: %w", pr.PullRequestID, err)
		}

		var reviewersToReplace []string
		for _, reviewerID := range currentReviewers {
			for _, deactivateID := range usersToDeactivate {
				if reviewerID == deactivateID {
					reviewersToReplace = append(reviewersToReplace, reviewerID)
					break
				}
			}
		}

		if len(reviewersToReplace) == 0 {
			continue
		}

		var authorTeam string
		teamQuery := `SELECT team_name FROM users WHERE user_id = $1`
		if err := tx.GetContext(ctx, &authorTeam, teamQuery, pr.AuthorID); err != nil {
			return nil, fmt.Errorf("error getting author team: %w", err)
		}

		excludeUsers := append(currentReviewers, pr.AuthorID)
		excludeUsers = append(excludeUsers, usersToDeactivate...)

		usersQuery := `SELECT user_id 
			FROM users 
			WHERE team_name = $1 
			AND is_active = true 
			AND user_id != ALL($2)
			ORDER BY RANDOM()`

		var users []string
		if err := tx.SelectContext(ctx, &users, usersQuery, authorTeam, pq.Array(excludeUsers)); err != nil {
			return nil, fmt.Errorf("error getting users: %w", err)
		}

		userIdx := 0
		for _, oldReviewerID := range reviewersToReplace {
			deleteQuery := `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2`

			if _, err := tx.ExecContext(ctx, deleteQuery, pr.PullRequestID, oldReviewerID); err != nil {
				return nil, fmt.Errorf("error removing reviewer %s from PR %s: %w", oldReviewerID, pr.PullRequestID, err)
			}

			if userIdx < len(users) {
				newReviewerID := users[userIdx]
				userIdx++

				insertQuery := `INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at) VALUES ($1, $2, NOW())`

				if _, err := tx.ExecContext(ctx, insertQuery, pr.PullRequestID, newReviewerID); err != nil {
					return nil, fmt.Errorf("error assigning new reviewer %s to PR %s: %w", newReviewerID, pr.PullRequestID, err)
				}
			}
		}
	}

	deactivateQuery := `UPDATE users 
		SET is_active = false, updated_at = NOW() 
		WHERE user_id = ANY($1)`

	if _, err := tx.ExecContext(ctx, deactivateQuery, pq.Array(usersToDeactivate)); err != nil {
		return nil, fmt.Errorf("error deactivating users: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction team: %w", err)
	}

	return usersToDeactivate, nil
}
