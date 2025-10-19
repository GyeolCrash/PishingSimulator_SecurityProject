package models

type User struct {
	Username     string `'json:"username"`
	PasswordHash string `json:"-"`
}
