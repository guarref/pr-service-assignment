package models

type TopReviewer struct {
	UserID      string `json:"user_id"      db:"user_id"`
	UserName    string `json:"username"     db:"username"`
	ReviewCount int    `json:"review_count" db:"review_count"`
}

type Stats struct {
	TotalTeams        int            `json:"total_teams"         db:"total_teams"`
	TotalUsers        int            `json:"total_users"         db:"total_users"`
	ActiveUsers       int            `json:"active_users"        db:"active_users"`
	TotalPullRequests int            `json:"total_pull_requests" db:"total_pull_requests"`
	OpenPullRequests  int            `json:"open_pull_requests"  db:"open_pull_requests"`
	TopReviewers      []*TopReviewer `json:"top_reviewers"       db:"-"`
}
