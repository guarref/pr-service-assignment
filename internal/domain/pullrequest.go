package domain

import (
	"database/sql"
	"time"
)

type PullRequestStatus string

const (
	PullRequestOpen   PullRequestStatus = "OPEN"
	PullRequestMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string            `json:"pull_request_id" db:"pull_request_id"`
	PullRequestName   string            `json:"pull_request_name" db:"pull_request_name"`
	AuthorID          string            `json:"author_id" db:"author_id"`
	Status            PullRequestStatus `json:"status" db:"status"`
	AssignedReviewers []string          `json:"assigned_reviewers" db:"-"`

	CreatedAt *time.Time    `json:"createdAt,omitempty" db:"created_at"`
	MergedAt  *sql.NullTime `json:"mergedAt,omitempty" db:"merged_at"`
}

type PullRequestShort struct {
	PullRequestID   string            `json:"pull_request_id"`
	PullRequestName string            `json:"pull_request_name"`
	AuthorID        string            `json:"author_id"`
	Status          PullRequestStatus `json:"status"`
}
