package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/guarref/pr-service-assignment/internal/domain"
	"github.com/guarref/pr-service-assignment/internal/resperrors"
	"github.com/jmoiron/sqlx"
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *sqlx.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (prr *PullRequestRepository) CreatePullRequest(ctx context.Context, pr *domain.PullRequest) error {

	tx, err := prr.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction pull_request: %w", err)
	}
	defer tx.Rollback()

	creationPullRequestQuery := `INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at, merged_at) 
		VALUES ($1, $2, $3, $4, NOW(), NULL) ON CONFLICT (pull_request_id) DO NOTHING`

	res, err := tx.ExecContext(ctx, creationPullRequestQuery,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
	)
	if err != nil {
		return fmt.Errorf("insert pull request: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected pull_request: %w", err)
	}
	if rows == 0 {
		// конфликт по PRIMARY KEY → PR уже существует
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
		return fmt.Errorf("commit transaction pull_request: %w", err)
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

	var reviewers []string
	if err := tx.SelectContext(ctx, &reviewers, reviewersQuery, prID); err != nil {
		return nil, fmt.Errorf("get pull request reviewers: %w", err)
	}
	pr.AssignedReviewers = reviewers

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction merge_pull_request: %w", err)
	}

	return &pr, nil
}

func (prr *PullRequestRepository) ReassignToPullRequest(ctx context.Context, prID string, oldUserID string) (*domain.PullRequest, error) {

	tx, err := prr.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction reassign_pull_request: %w", err)
	}
	defer tx.Rollback()

	prQuery := `SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
		FOR UPDATE`

	var pr domain.PullRequest
	if err := tx.GetContext(ctx, &pr, prQuery, prID); err != nil {
		if err == sql.ErrNoRows {
			return nil, resperrors.ErrPullRequestNotFound
		}
		return nil, fmt.Errorf("get pull request: %w", err)
	}

	if pr.Status == domain.PullRequestMerged {
		return nil, resperrors.ErrPullRequestMerged
	}

	checkReviewerQuery := `SELECT EXISTS(
		SELECT 1 FROM pr_reviewers
		WHERE pull_request_id = $1 AND user_id = $2
	)`

	var isAssigned bool
	if err := tx.GetContext(ctx, &isAssigned, checkReviewerQuery, prID, oldUserID); err != nil {
		return nil, fmt.Errorf("check reviewer assigned: %w", err)
	}
	if !isAssigned {
		return nil, resperrors.ErrNotAssigned
	}

	candidateQuery := `SELECT cand.user_id
		FROM users cand
		JOIN users old_user ON old_user.user_id = $2
		JOIN pull_requests pr ON pr.pull_request_id = $1
		WHERE cand.team_name = old_user.team_name
		  AND cand.is_active = true
		  AND cand.user_id <> old_user.user_id
		  AND cand.user_id <> pr.author_id
		  AND NOT EXISTS (
				SELECT 1 FROM pr_reviewers r
				WHERE r.pull_request_id = $1 AND r.user_id = cand.user_id
		  )
		ORDER BY random()
		LIMIT 1`

	var newReviewerID string
	if err := tx.GetContext(ctx, &newReviewerID, candidateQuery, prID, oldUserID); err != nil {
		if err == sql.ErrNoRows {
			return nil, resperrors.ErrNoCandidate
		}
		return nil, fmt.Errorf("get replacement candidate: %w", err)
	}

	deleteQuery := `DELETE FROM pr_reviewers
		WHERE pull_request_id = $1 AND user_id = $2`

	if _, err := tx.ExecContext(ctx, deleteQuery, prID, oldUserID); err != nil {
		return nil, fmt.Errorf("delete old reviewer: %w", err)
	}

	insertQuery := `INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at)
		VALUES ($1, $2, NOW())`

	if _, err := tx.ExecContext(ctx, insertQuery, prID, newReviewerID); err != nil {
		return nil, fmt.Errorf("insert new reviewer %s: %w", newReviewerID, err)
	}

	reviewersQuery := `SELECT user_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at`

	var reviewers []string
	if err := tx.SelectContext(ctx, &reviewers, reviewersQuery, prID); err != nil {
		return nil, fmt.Errorf("get updated reviewers: %w", err)
	}
	pr.AssignedReviewers = reviewers

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction reassign_pull_request: %w", err)
	}

	return &pr, nil
}

func (prr *PullRequestRepository) GetPullRequestByReviewerID(ctx context.Context, userID string) ([]*domain.PullRequest, error) {

	query := `SELECT
			pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status,
			pr.created_at,
			pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.user_id = $1
		ORDER BY pr.created_at DESC`

	var prs []*domain.PullRequest
	if err := prr.db.SelectContext(ctx, &prs, query, userID); err != nil {
		return nil, fmt.Errorf("get pull requests by reviewer %s: %w", userID, err)
	}

	if prs == nil {
		prs = []*domain.PullRequest{}
	}

	return prs, nil
}
