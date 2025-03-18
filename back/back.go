package back;

import (
	"net/http"
	"database/sql"
	"log"
	"tuel/db"
	"io"
	"encoding/json"
	"fmt"
	"strings"
)

type Tool struct {
	Name string `json:"name"`
	Descr string `json:"descr"` // description
}

type Toolset struct {
	Name string `json:"name"`
	Descr string `json:"descr"` // description
}

type Name struct {
	Name string `json: name`
}

var database *sql.DB
var toolTable string
var toolsetTable string

// write tool item
func put(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read the request body")
	}
	tool := Tool{}
	err = json.Unmarshal(body, &tool)
	if err != nil {
		log.Println("Unmarshal failed, err:", err)
	}
	// sql single quote escape
	tool.Name = strings.Replace(tool.Name, "'", "''", -1)
	tool.Descr = strings.Replace(tool.Descr, "'", "''", -1)
	q := fmt.Sprintf(`
INSERT INTO %s (name, descr)
VALUES ('%s', '%s')
	`, toolTable, tool.Name, tool.Descr)
	result, err := database.Exec(q)
	if err != nil {
		log.Printf("Insert failed in put(), name: \"%s\", descr: \"%s\". err: %s",
		tool.Name, tool.Descr, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("failed to get rows affected, err:", err)
	}
	log.Println("put() rows affected", rowsAffected)
}

// get tool info
func get(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read the request body")
	}
	name := Name{}
	err = json.Unmarshal(body, &name)
	if err != nil {
		log.Println("Unmarshal failed, err:", err)
	}
	// sql single quote escape
	name.Name = strings.Replace(name.Name, "'", "''", -1)
	q := fmt.Sprintf(`
SELECT name, descr FROM %s
WHERE name = '%s'
	`, toolTable, name.Name)
	rows, err := database.Query(q)
	defer rows.Close()
	if err != nil {
		log.Printf("select failed in get(), name: \"%s\". err: %s",
		name.Name, err)
	}
	var tools []Tool
	var _name string
	var descr string
	for rows.Next() {
		rows.Scan(&_name, &descr)
		tools = append(tools, Tool{Name: _name, Descr: descr})
	}

	type Resp struct {
		Code int `json:"code"`
		Tool []Tool `json:"tool"` // let's use singular for simplicity
	}
	resp := Resp {
		Code: 200,
		Tool: tools,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed in get(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in get(), err:", err)
	}
	log.Printf("get(): got %d records\n", len(tools))
}

// put toolset
func putToolset(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read the request body")
	}
	toolset := Toolset{}
	err = json.Unmarshal(body, &toolset)
	if err != nil {
		log.Println("Unmarshal failed, err:", err)
	}
	// sql single quote escape
	toolset.Name = strings.Replace(toolset.Name, "'", "''", -1)
	toolset.Descr = strings.Replace(toolset.Descr, "'", "''", -1)
	q := fmt.Sprintf(`
INSERT INTO %s(name, descr)
VALUES ('%s', '%s')
	`, toolsetTable, toolset.Name, toolset.Descr)
	result, err := database.Exec(q)
	if err != nil {
		log.Printf("Insert failed in putToolset(), name: \"%s\", descr: \"%s\". err: %s",
		toolset.Name, toolset.Descr, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("failed to get rows affected, err:", err)
	}
	log.Println("putToolset() rows affected", rowsAffected)
}

// get toolset
func getToolset(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read the request body")
	}
	name := Name{}
	err = json.Unmarshal(body, &name)
	if err != nil {
		log.Println("Unmarshal failed, err:", err)
	}
	// sql single quote escape
	name.Name = strings.Replace(name.Name, "'", "''", -1)
	q := fmt.Sprintf(`
SELECT name, descr FROM %s
WHERE name = '%s'
	`, toolsetTable, name.Name)
	rows, err := database.Query(q)
	defer rows.Close()
	if err != nil {
		log.Printf("select failed in getToolset(), name: \"%s\". err: %s",
		name.Name, err)
	}
	var toolsets []Toolset
	var _name string
	var descr string
	for rows.Next() {
		rows.Scan(&_name, &descr)
		toolsets = append(toolsets, Toolset{Name: _name, Descr: descr})
	}

	type Resp struct {
		Code int `json:"code"`
		Toolset []Toolset `json:"toolset"` // let's use singular for simplicity
	}
	resp := Resp {
		Code: 200,
		Toolset: toolsets,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed in getToolset(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in getToolset(), err:", err)
	}
	log.Printf("getToolset(): got %d records\n", len(toolsets))
}

func StartBackend(_db *sql.DB) {
	database = _db
	toolTable = "tool"
	toolsetTable = "toolset"
	db.CreateTablesIfNone(database, toolTable, toolsetTable)
	http.HandleFunc("/put", put)
	http.HandleFunc("/get", get)
	http.HandleFunc("/getset", getToolset)
	http.HandleFunc("/putset", putToolset)
	port := ":9160"
	fmt.Println("Serving backend on http://localhost" + port)
	http.ListenAndServe(port, nil)
}
