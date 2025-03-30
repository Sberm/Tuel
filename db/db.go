package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func ConnectDB() *sql.DB {
	dsn := "db/tuel.db"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Database %s opened\n", dsn)
	return db
}

func CreateTablesIfNone(db *sql.DB, toolTable string, toolsetTable string, toolsetRelTable string) {
	q := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id INTEGER PRIMARY KEY,
	name TEXT,
	descr TEXT
);
	`, toolTable)
	_, err := db.Exec(q)
	if err != nil {
		log.Fatal("create tool table failed", err)
	}

	q = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id INTEGER PRIMARY KEY,
	name TEXT,
	descr TEXT
);
	`, toolsetTable)
	_, err = db.Exec(q)
	if err != nil {
		log.Fatal("create toolset table failed", err)
	}

	q = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	tool_id INTEGER,
	toolset_id INTEGER,
	PRIMARY KEY (tool_id, toolset_id),
	FOREIGN KEY (tool_id) REFERENCES %s (id) ON DELETE CASCADE,
	FOREIGN KEY (toolset_id) REFERENCES %s (id) ON DELETE CASCADE
);
	`, toolsetRelTable, toolTable, toolsetTable)
	_, err = db.Exec(q)
	if err != nil {
		log.Fatal("create tool & toolset relation table failed", err)
	}

	fmt.Println("tables", toolTable, toolsetTable, toolsetRelTable, "created")
}

func TurnOnForeignKey(db *sql.DB) {
	_, err := db.Exec("PRAGMA foreign_keys = ON;");
	if err != nil {
		log.Println("foreign key contraint cannot be turned on")
	}
}
