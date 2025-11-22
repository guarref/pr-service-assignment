package domain

import "time"

type TeamMember struct {
	UserID   string `json:"user_id" db:"user_id"`
	UserName string `json:"username" db:"username"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name" db:"team_name"`
	Members  []TeamMember `json:"members" db:"-"`

	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}
