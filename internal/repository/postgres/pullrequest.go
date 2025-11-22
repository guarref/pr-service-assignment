package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/guarref/pr-service-assignment/internal/errors"
	"github.com/jmoiron/sqlx"
)

type PullRequestRepository struct {
	db       *sqlx.DB
	userRepo *UserRepository
}

func NewPullRequestRepository(db *sqlx.DB, userRepo *UserRepository) *PullRequestRepository {
	return &PullRequestRepository{db: db, userRepo: userRepo}
}

func (prr *PullRequestRepository) CreatePullRequest(ctx context.Context, pr *models.PullRequest) (*models.PullRequest, error) {

	tx, err := prr.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction create_pull_request: %w", err)
	}
	defer tx.Rollback()

	creationPullRequestQuery := `INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
		VALUES ($1, $2, $3, $4, NOW(), NULL)
		ON CONFLICT (pull_request_id) DO NOTHING
		RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at`

	var newPR models.PullRequest

	err = tx.GetContext(ctx, &newPR, creationPullRequestQuery, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrPullRequestExists
		}
		return nil, fmt.Errorf("error pull request creation: %w", err)
	}

	if len(pr.AssignedReviewers) > 0 {
		reviewerIns := `INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at)
			VALUES ($1, $2, NOW())`

		for _, reviewerID := range pr.AssignedReviewers {
			if _, err := tx.ExecContext(ctx, reviewerIns, newPR.PullRequestID, reviewerID); err != nil {
				return nil, fmt.Errorf("error addition reviewer %s: %w", reviewerID, err)
			}
		}
	}

	newPR.AssignedReviewers = append([]string(nil), pr.AssignedReviewers...)

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction create_pull_request: %w", err)
	}

	return &newPR, nil
}

func (prr *PullRequestRepository) MergePullRequestByID(ctx context.Context, prID string) (*models.PullRequest, error) {

	tx, err := prr.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction merge_pull_request: %w", err)
	}
	defer tx.Rollback()

	updateQuery := `UPDATE pull_requests
		SET status = $1,
		merged_at = COALESCE(merged_at, NOW())
		WHERE pull_request_id = $2
		RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at`

	var pr models.PullRequest
	if err := tx.GetContext(ctx, &pr, updateQuery, models.PullRequestMerged, prID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrPullRequestNotFound
		}
		return nil, fmt.Errorf("error updating pull request status to MERGED: %w", err)
	}

	reviewersQuery := `SELECT user_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at`

	var rev_users []string
	if err := tx.SelectContext(ctx, &rev_users, reviewersQuery, prID); err != nil {
		return nil, fmt.Errorf("error getting pull request reviewers: %w", err)
	}
	pr.AssignedReviewers = rev_users

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction merge_pull_request: %w", err)
	}

	return &pr, nil
}

func (prr *PullRequestRepository) ReassignToPullRequest(ctx context.Context, prID string, oldUserID string) (*models.PullRequest, string, error) {

	tx, err := prr.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error beginning transaction reassign_pull_request: %w", err)
	}
	defer tx.Rollback()

	prQuery := `SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
		FOR UPDATE`

	var pr models.PullRequest
	if err := tx.GetContext(ctx, &pr, prQuery, prID); err != nil {
		if err == sql.ErrNoRows {
			return nil, "", errors.ErrPullRequestNotFound
		}
		return nil, "", fmt.Errorf("error getting pull request: %w", err)
	}

	if pr.Status == models.PullRequestMerged {
		return nil, "", errors.ErrPullRequestMerged
	}

	checkReviewerQuery := `SELECT EXISTS(SELECT 1 FROM pr_reviewers
		WHERE pull_request_id = $1 AND user_id = $2)`

	var isAssigned bool
	if err := tx.GetContext(ctx, &isAssigned, checkReviewerQuery, prID, oldUserID); err != nil {
		return nil, "", fmt.Errorf("error checking reviewer assigned: %w", err)
	}
	if !isAssigned {
		return nil, "", errors.ErrNotAssigned
	}

	teamQuery := `SELECT team_name
		FROM users
		WHERE user_id = $1`

	var oldUserTeam string
	if err := tx.GetContext(ctx, &oldUserTeam, teamQuery, oldUserID); err != nil {
		if err == sql.ErrNoRows {
			return nil, "", errors.ErrUserNotFound
		}
		return nil, "", fmt.Errorf("error getting old user team: %w", err)
	}

	reviewersQuery := `SELECT user_id 
		FROM pr_reviewers 
		WHERE pull_request_id = $1`

	var currentReviewers []string
	if err := tx.SelectContext(ctx, &currentReviewers, reviewersQuery, prID); err != nil {
		return nil, "", fmt.Errorf("error getting current reviewers: %w", err)
	}

	currentSet := make(map[string]struct{}, len(currentReviewers))
	for _, id := range currentReviewers {
		currentSet[id] = struct{}{}
	}

	usersQuery := `SELECT user_id
		FROM users
		WHERE team_name = $1 AND is_active = true AND user_id <> $2 AND user_id <> $3`

	var allUsersIDs []string
	if err := tx.SelectContext(ctx, &allUsersIDs, usersQuery, oldUserTeam, oldUserID, pr.AuthorID); err != nil {
		return nil, "", fmt.Errorf("error getting users for reassign: %w", err)
	}

	accessible := make([]string, 0, len(allUsersIDs))
	for _, id := range allUsersIDs {
		if _, ok := currentSet[id]; ok {
			continue
		}
		accessible = append(accessible, id)
	}

	if len(accessible) == 0 {
		return nil, "", errors.ErrNoCandidate
	}

	newReviewerID := getRandomID(accessible)

	deleteQuery := `DELETE FROM pr_reviewers
		WHERE pull_request_id = $1 AND user_id = $2`

	if _, err := tx.ExecContext(ctx, deleteQuery, prID, oldUserID); err != nil {
		return nil, "", fmt.Errorf("error delete old reviewer: %w", err)
	}

	reviewerIns := `INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at)
		VALUES ($1, $2, NOW())`
	if _, err := tx.ExecContext(ctx, reviewerIns, prID, newReviewerID); err != nil {
		return nil, "", fmt.Errorf("error addition reviewer %s: %w", newReviewerID, err)
	}

	var updatedReviewers []string
	if err := tx.SelectContext(ctx, &updatedReviewers, reviewersQuery, prID); err != nil {
		return nil, "", fmt.Errorf("error getting updated reviewers: %w", err)
	}
	pr.AssignedReviewers = updatedReviewers

	if err := tx.Commit(); err != nil {
		return nil, "", fmt.Errorf("error committing transaction reassign_pull_request: %w", err)
	}

	return &pr, newReviewerID, nil
}

func getRandomID(ids []string) string {

	if len(ids) == 0 {
		return ""
	}
	if len(ids) == 1 {
		return ids[0]
	}

	index := rand.Intn(len(ids))

	return ids[index]
}

func (prr *PullRequestRepository) GetPullRequestByReviewerID(ctx context.Context, userID string) ([]*models.PullRequestShort, error) {

	query := `SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id,pr.status
        FROM pull_requests pr
        INNER JOIN pr_reviewers prev ON pr.pull_request_id = prev.pull_request_id
        WHERE prev.user_id = $1
        ORDER BY pr.created_at DESC`

	var prs []*models.PullRequestShort
	if err := prr.db.SelectContext(ctx, &prs, query, userID); err != nil {
		return nil, fmt.Errorf("error getting pull requests by reviewer %s: %w", userID, err)
	}

	if prs == nil {
		prs = []*models.PullRequestShort{}
	}

	return prs, nil
}
