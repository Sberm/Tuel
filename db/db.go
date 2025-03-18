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

func CreateTablesIfNone(db *sql.DB, toolTable string, toolsetTable string) {
	fmt.Println("creating tables", toolTable, toolsetTable)
	q := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id INTEGER PRIMARY KEY,
	name TEXT,
	descr TEXT
);
	`, toolsetTable)
	_, err := db.Exec(q)
	if err != nil {
		log.Fatal("create toolset table failed", err)
	}

	q = fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	id INTEGER PRIMARY KEY,
	name TEXT,
	descr TEXT,
	toolsetid INTEGER,
	FOREIGN KEY(toolsetid) REFERENCES %s(id)
);
	`, toolTable, toolsetTable)
	_, err = db.Exec(q)
	if err != nil {
		log.Fatal("create tool table failed", err)
	}
}
