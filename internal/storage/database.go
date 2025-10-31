package storage

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func InitDB() {
	var err error

	db, err = sql.Open("sqlite", "./phising_simulator.db")
	if err != nil {
		log.Fatal("InitDB(): Failed to open databse :", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("storage.InitDB(): Failed to connect to database: ", err)
	}

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT, 
			"username" TEXT NOT NULL UNIQUE,
			"password_hash" TEXT NOT NULL
	);`
	createRecordingsTable := `
	CREATE TABLE IF NOT EXISTS recordings (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"user_id" INTEGER NOT NULL,
			"scenario_key" TEXT,
			"file_path" TEXT NOT NULL,
			"created_at" DATETIME NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
	)`

	if _, err := db.Exec(createUsersTable); err != nil {
		log.Fatalf("InitDB(): Failed to create users table: %v", err)
	}
	if _, err := db.Exec(createRecordingsTable); err != nil {
		log.Fatalf("InitDB(): Failed to create recrodings table: %v", err)
	}
	log.Println("InitDB(): Init and create table successfully!")

}
