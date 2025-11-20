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

// GetByID — получить PR + список ревьюверов.
func (prr *PullRequestRepository) GetByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest

	prQuery := `
		SELECT
			pull_request_id,
			pull_request_name,
			author_id,
			status,
			created_at,
			merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`
	if err := prr.db.GetContext(ctx, &pr, prQuery, prID); err != nil {
		if err == sql.ErrNoRows {
			return nil, resperrors.ErrPullRequestNotFound
		}
		return nil, fmt.Errorf("get pull request: %w", err)
	}

	reviewersQuery := `
		SELECT user_id
		FROM pr_reviewers
		WHERE pull_request_id = $1
		ORDER BY assigned_at
	`
	var reviewers []string
	if err := prr.db.SelectContext(ctx, &reviewers, reviewersQuery, prID); err != nil {
		return nil, fmt.Errorf("get pull request reviewers: %w", err)
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

// Update — обновить PR и полностью пересоздать список ревьюверов.
func (prr *PullRequestRepository) Update(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := prr.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction pull_request update: %w", err)
	}
	defer tx.Rollback()

	updatePRQuery := `
		UPDATE pull_requests
		SET
			pull_request_name = $2,
			author_id         = $3,
			status            = $4,
			created_at        = $5,
			merged_at         = $6
		WHERE pull_request_id = $1
	`
	res, err := tx.ExecContext(
		ctx,
		updatePRQuery,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return fmt.Errorf("update pull request: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected pull_request update: %w", err)
	}
	if rows == 0 {
		return resperrors.ErrPullRequestNotFound
	}

	// чистим старых ревьюверов
	deleteReviewersQuery := `
		DELETE FROM pr_reviewers
		WHERE pull_request_id = $1
	`
	if _, err := tx.ExecContext(ctx, deleteReviewersQuery, pr.PullRequestID); err != nil {
		return fmt.Errorf("delete pull request reviewers: %w", err)
	}

	// вставляем новых ревьюверов
	if len(pr.AssignedReviewers) > 0 {
		insertReviewerQuery := `
			INSERT INTO pr_reviewers (pull_request_id, user_id, assigned_at)
			VALUES ($1, $2, NOW())
		`
		for _, reviewerID := range pr.AssignedReviewers {
			if _, err := tx.ExecContext(ctx, insertReviewerQuery, pr.PullRequestID, reviewerID); err != nil {
				return fmt.Errorf("insert pull request reviewer %s: %w", reviewerID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction pull_request update: %w", err)
	}

	return nil
}

// GetByReviewerID — PR'ы, где пользователь назначен ревьювером.
// Потом в сервисе/хендлере ты из них сделаешь PullRequestShort для /users/getReview.
func (prr *PullRequestRepository) GetByReviewerID(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	query := `
		SELECT
			pr.pull_request_id,
			pr.pull_request_name,
			pr.author_id,
			pr.status,
			pr.created_at,
			pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers rev
			ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.user_id = $1
		ORDER BY pr.created_at DESC
	`

	var prs []*domain.PullRequest
	if err := prr.db.SelectContext(ctx, &prs, query, userID); err != nil {
		return nil, fmt.Errorf("get pull requests by reviewer %s: %w", userID, err)
	}

	return prs, nil
}
