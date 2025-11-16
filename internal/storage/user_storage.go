package storage

import (
	"PishingSimulator_SecurityProject/internal/models"
	"database/sql"
	"errors"

	"modernc.org/sqlite"
)

var ErrUsernameExists = errors.New("username already exists")

func CreateUser(username, passwordHash string, profile models.UserProfile) error {
	stmt, err := db.Prepare("INSERT INTO users(username, password_hash, name, age, gender) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, passwordHash, profile.Name, profile.Age, profile.Gender)
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

	row := db.QueryRow("SELECT id, username, password_hash, name, age, gender FROM users WHERE username = ?", username)

	var nullAge sql.NullInt64
	var nullName, nullGender sql.NullString

	if err := row.Scan(
		&id, &user.Username,
		&user.PasswordHash,
		&nullName,
		&nullAge,
		&nullGender,
	); err != nil {
		if err == sql.ErrNoRows {
			return user, err // no selected user
		}
		return user, err // unexpected error
	}

	if nullName.Valid {
		user.Profile.Name = nullName.String
	}
	if nullAge.Valid {
		user.Profile.Age = int(nullAge.Int64)
	}
	if nullGender.Valid {
		user.Profile.Gender = nullGender.String
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
