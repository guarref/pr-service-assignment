package postgres

import (
	"context"
	"fmt"

	"github.com/guarref/pr-service-assignment/internal/models"
	"github.com/jmoiron/sqlx"
)

type StatsRepository struct {
	db *sqlx.DB
}

func NewStatsRepository(db *sqlx.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (sr *StatsRepository) GetStats(ctx context.Context, top int) (*models.Stats, error) {

	var stats models.Stats

	countsQuery := `SELECT
    	(SELECT COUNT(*) FROM teams) AS total_teams,
    	(SELECT COUNT(*) FROM users) AS total_users,
    	(SELECT COUNT(*) FROM users WHERE is_active = TRUE) AS active_users,
    	(SELECT COUNT(*) FROM pull_requests) AS total_pull_requests,
    	(SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN') AS open_pull_requests`

	if err := sr.db.GetContext(ctx, &stats, countsQuery); err != nil {
		return nil, fmt.Errorf("error getting stats: %w", err)
	}

	topReviewersQuery := `SELECT u.user_id, u.username, COUNT(DISTINCT prr.pull_request_id) AS review_count
		FROM users u
		JOIN pr_reviewers prr ON u.user_id = prr.user_id
		GROUP BY u.user_id, u.username
		ORDER BY review_count DESC, u.user_id
		LIMIT $1`

	var reviewers []*models.TopReviewer

	if err := sr.db.SelectContext(ctx, &reviewers, topReviewersQuery, top); err != nil {
		return nil, fmt.Errorf("error getting top reviewers: %w", err)
	}

	if reviewers == nil {
		reviewers = []*models.TopReviewer{}
	}

	stats.TopReviewers = reviewers

	return &stats, nil
}
