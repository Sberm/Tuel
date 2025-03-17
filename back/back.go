package back;

import (
	"net/http"
	"database/sql"
	"log"
	"tuel/db"
)

// write tool item
func put(w http.ResponseWriter, r *http.Request) {
	log.Print("put")
}

// get tool info
func get(w http.ResponseWriter, r *http.Request) {
	log.Print("get")
}

// put toolset
func putToolset(w http.ResponseWriter, r *http.Request) {
	log.Print("put toolset")
}

// get toolset
func getToolset(w http.ResponseWriter, r *http.Request) {
	log.Print("get toolset")
}

func StartBackend(_db *sql.DB) {
	db.CreateTableIfNone(_db)
	http.HandleFunc("/put", put)
	http.HandleFunc("/get", get)
	http.HandleFunc("/getset", getToolset)
	http.HandleFunc("/putset", putToolset)
	http.ListenAndServe(":9160", nil)
}
