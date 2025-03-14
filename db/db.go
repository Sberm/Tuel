package db;

import(
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"fmt"
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

func CreateTable(db *sql.DB) {
	// TODO: check if table exists before creating one
	q := `
CREATE TABLE tool (
	id INTEGER PRIMARY KEY,
	name TEXT,
	desc TEST
);
	`
	_, err := db.Query(q)
	if err != nil {
		log.Fatal(err)
	}
}
