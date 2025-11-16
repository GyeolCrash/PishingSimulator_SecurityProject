package models

import "time"

type Record struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Scenario  string    `json:"scenario"`
	FilePath  string    `json:"file_path"`
	CreatedAt time.Time `json:"created_at"`
}
