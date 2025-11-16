package storage

import (
	"PishingSimulator_SecurityProject/internal/models"
	"time"
)

func CreateRecords(userID int, scenarioKey string, filePath string) error {
	stmt, err := db.Prepare("INSERT INTO records(user_id, scenario_key, file_path, created_at) VALUES(?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, scenarioKey, filePath, time.Now())
	return err
}

func GetRecordsByUserID(userID int) ([]models.Record, error) {
	query := `
		SELECT id, scenario_key, file_path, created_at 
		FROM records 
		WHERE user_id = ? 
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var r models.Record
		var createdStr string // SQLite는 시간을 문자열로 저장함

		if err := rows.Scan(&r.ID, &r.Scenario, &r.FilePath, &createdStr); err != nil {
			return nil, err
		}

		// 시간 파싱 (SQLite 포맷에 따라 수정 필요할 수 있음)
		parsedTime, _ := time.Parse("2006-01-02 15:04:05", createdStr)
		r.CreatedAt = parsedTime

		records = append(records, r)
	}
	return records, nil
}
