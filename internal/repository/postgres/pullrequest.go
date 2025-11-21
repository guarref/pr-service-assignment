package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/guarref/pr-service-assignment/internal/domain"
	"github.com/guarref/pr-service-assignment/internal/resperrors"
	"github.com/jmoiron/sqlx"
)

type PullRequestRepository struct {
	db       *sqlx.DB
	userRepo *UserRepository
}

func NewPullRequestRepository(db *sqlx.DB, userRepo *UserRepository) *PullRequestRepository {
	return &PullRequestRepository{
		db:       db,
		userRepo: userRepo,
	}
}

func (prr *PullRequestRepository) CreatePullRequest(ctx context.Context, pr *domain.PullRequest) error {

	tx, err := prr.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction create_pull_request: %w", err)
	}
	defer tx.Rollback()

	creationPullRequestQuery := `INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at) 
		VALUES ($1, $2, $3, $4, NOW(), NULL) ON CONFLICT (pull_request_id) DO NOTHING`

	res, err := tx.ExecContext(ctx, creationPullRequestQuery, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status)
	if err != nil {
		return fmt.Errorf("insert pull request: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected pull_request: %w", err)
	}
	if rows == 0 {
		return resperrors.ErrPullRequestExists
	}

	if len(pr.AssignedReviewers) > 0 {
		insertReviewerQuery := `INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at)
			VALUES ($1, $2, NOW())`

		for _, reviewerID := range pr.AssignedReviewers {
			if _, err := tx.ExecContext(ctx, insertReviewerQuery, pr.PullRequestID, reviewerID); err != nil {
				return fmt.Errorf("addition reviewer %s: %w", reviewerID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction create_pull_request: %w", err)
	}

	return nil
}

func (prr *PullRequestRepository) MergePullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {

	tx, err := prr.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction merge_pull_request: %w", err)
	}
	defer tx.Rollback()

	updateQuery := `UPDATE pull_requests
		SET status = $1,
		    merged_at = COALESCE(merged_at, NOW())
		WHERE pull_request_id = $2
		RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at`

	var pr domain.PullRequest
	if err := tx.GetContext(ctx, &pr, updateQuery, domain.PullRequestMerged, prID); err != nil {
		if err == sql.ErrNoRows {
			return nil, resperrors.ErrPullRequestNotFound
		}
		return nil, fmt.Errorf("update pull request status to MERGED: %w", err)
	}

	reviewersQuery := `SELECT user_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at`

	var rev_users []string
	if err := tx.SelectContext(ctx, &rev_users, reviewersQuery, prID); err != nil {
		return nil, fmt.Errorf("get pull request reviewers: %w", err)
	}
	pr.AssignedReviewers = rev_users

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction merge_pull_request: %w", err)
	}

	return &pr, nil
}

func (prr *PullRequestRepository) ReassignToPullRequest(
	ctx context.Context,
	prID string,
	oldUserID string,
) (*domain.PullRequest, string, error) {
	tx, err := prr.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, "", fmt.Errorf("beginning transaction reassign_pull_request: %w", err)
	}
	defer tx.Rollback()

	prQuery := `SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
		FOR UPDATE`

	var pr domain.PullRequest
	if err := tx.GetContext(ctx, &pr, prQuery, prID); err != nil {
		if err == sql.ErrNoRows {
			return nil, "", resperrors.ErrPullRequestNotFound
		}
		return nil, "", fmt.Errorf("get pull request: %w", err)
	}

	if pr.Status == domain.PullRequestMerged {
		return nil, "", resperrors.ErrPullRequestMerged
	}

	checkReviewerQuery := `SELECT EXISTS(
		SELECT 1 FROM pr_reviewers
		WHERE pull_request_id = $1 AND user_id = $2
	)`

	var isAssigned bool
	if err := tx.GetContext(ctx, &isAssigned, checkReviewerQuery, prID, oldUserID); err != nil {
		return nil, "", fmt.Errorf("check reviewer assigned: %w", err)
	}
	if !isAssigned {
		return nil, "", resperrors.ErrNotAssigned
	}

	// команда заменяемого ревьювера
	teamQuery := `SELECT team_name
		FROM users
		WHERE user_id = $1`

	var oldUserTeam string
	if err := tx.GetContext(ctx, &oldUserTeam, teamQuery, oldUserID); err != nil {
		if err == sql.ErrNoRows {
			return nil, "", resperrors.ErrUserNotFound
		}
		return nil, "", fmt.Errorf("get old user team: %w", err)
	}

	// текущие ревьюверы PR
	reviewersQuery := `SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1`
	var currentReviewers []string
	if err := tx.SelectContext(ctx, &currentReviewers, reviewersQuery, prID); err != nil {
		return nil, "", fmt.Errorf("get current reviewers: %w", err)
	}

	currentSet := make(map[string]struct{}, len(currentReviewers))
	for _, id := range currentReviewers {
		currentSet[id] = struct{}{}
	}

	// кандидаты из команды старого ревьювера: активные, не автор и не уже ревьюверы
	candidatesQuery := `SELECT user_id
		FROM users
		WHERE team_name = $1
		  AND is_active = true
		  AND user_id <> $2
		  AND user_id <> $3`

	var allCandidateIDs []string
	if err := tx.SelectContext(ctx, &allCandidateIDs, candidatesQuery, oldUserTeam, oldUserID, pr.AuthorID); err != nil {
		return nil, "", fmt.Errorf("get candidates for reassign: %w", err)
	}

	available := make([]string, 0, len(allCandidateIDs))
	for _, id := range allCandidateIDs {
		if _, ok := currentSet[id]; ok {
			continue
		}
		available = append(available, id)
	}

	if len(available) == 0 {
		return nil, "", resperrors.ErrNoCandidate
	}

	newReviewerID := pickRandomID(available)

	deleteQuery := `DELETE FROM pr_reviewers
		WHERE pull_request_id = $1 AND user_id = $2`
	if _, err := tx.ExecContext(ctx, deleteQuery, prID, oldUserID); err != nil {
		return nil, "", fmt.Errorf("delete old reviewer: %w", err)
	}

	insertQuery := `INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at)
		VALUES ($1, $2, NOW())`
	if _, err := tx.ExecContext(ctx, insertQuery, prID, newReviewerID); err != nil {
		return nil, "", fmt.Errorf("insert new reviewer %s: %w", newReviewerID, err)
	}

	var updatedReviewers []string
	if err := tx.SelectContext(ctx, &updatedReviewers, reviewersQuery, prID); err != nil {
		return nil, "", fmt.Errorf("get updated reviewers: %w", err)
	}
	pr.AssignedReviewers = updatedReviewers

	if err := tx.Commit(); err != nil {
		return nil, "", fmt.Errorf("commit transaction reassign_pull_request: %w", err)
	}

	return &pr, newReviewerID, nil
}

func pickRandomID(ids []string) string {

	if len(ids) == 0 {
		return ""
	}
	if len(ids) == 1 {
		return ids[0]
	}

	index := rand.Intn(len(ids))

	return ids[index]
}

func (prr *PullRequestRepository) GetPullRequestByReviewerID(ctx context.Context, userID string) ([]*domain.PullRequestShort, error) {

	query := `SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers prev ON pr.pull_request_id = prev.pull_request_id
		WHERE prev.user_id = $1
		ORDER BY pr.created_at DESC`

	var prs []*domain.PullRequestShort
	if err := prr.db.SelectContext(ctx, &prs, query, userID); err != nil {
		return nil, fmt.Errorf("get pull requests by reviewer %s: %w", userID, err)
	}

	if prs == nil {
		prs = []*domain.PullRequestShort{}
	}

	return prs, nil
}
