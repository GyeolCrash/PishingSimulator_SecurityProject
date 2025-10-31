package storage

import (
	"PishingSimulator_SecurityProject/internal/models"
	"database/sql"
	"errors"

	"modernc.org/sqlite"
)

var ErrUsernameExists = errors.New("username already exists")

func CreateUser(username, passwordHash string) error {
	stmt, err := db.Prepare("INSERT INTO users(username, password_hash) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, passwordHash)
	if err != nil {
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code() == 2067 {
				return ErrUsernameExists
			}
		}
		return err
	}
	return nil
}

func GetUserByUsername(username string) (models.User, error) {
	var user models.User
	var id int

	row := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?", username)

	if err := row.Scan(&id, &user.Username, &user.PasswordHash); err != nil {
		if err == sql.ErrNoRows {
			return user, err // no selected user
		}
		return user, err // unexpected error
	}

	return user, nil
}

func GetUserIDByUsername(username string) (int, error) {
	var id int
	row := db.QueryRow("SELECT id FROM users WHERE username = ?", username)
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, err
		}
		return 0, err
	}
	return id, nil
}
