package storage

import (
	"time"
)

func CreateRecording(userID int, scenarioKey string, filePath string) error {
	stmt, err := db.Prepare("INSERT INTO recordings(user_id, scenario_key, file_path, created_at) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, scenarioKey, filePath, time.Now())
	return err
}
