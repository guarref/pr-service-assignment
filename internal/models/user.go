package domain

import "time"

type User struct {
	UserID   string `json:"user_id" db:"user_id"`
	UserName string `json:"username" db:"username"`
	TeamName string `json:"team_name" db:"team_name"`
	IsActive bool   `json:"is_active" db:"is_active"`

	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}
