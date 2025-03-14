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
