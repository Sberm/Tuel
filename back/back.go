package back;

import (
	"net/http"
	"database/sql"
	"log"
	"tuel/db"
	"io"
	"encoding/json"
	"fmt"
)

type Tool struct {
	Name string `json:"name"`
	Descr string `json:"descr"` // description
}

type ToolWId struct {
	Id uint64 `json:"id"`
	Name string `json:"name"`
	Descr string `json:"descr"` // description
}

type Toolset struct {
	Name string `json:"name"`
	Descr string `json:"descr"` // description
}

type ToolsetWId struct {
	Id uint64 `json:"id"`
	Name string `json:"name"`
	Descr string `json:"descr"` // description
}

type Name struct {
	Name string `json:"name"`
}

type Id struct {
	Id uint64 `json:"id"`
}

var database *sql.DB
var toolTable string
var toolsetTable string

// write tool item, return an id
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
	q := fmt.Sprintf(`
INSERT INTO %s(name, descr)
VALUES (?, ?)
	`, toolTable)
	result, err := database.Exec(q, tool.Name, tool.Descr)
	if err != nil {
		log.Printf("Insert failed in put(), name: \"%s\", descr: \"%s\". err: %s",
		tool.Name, tool.Descr, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("failed to get rows affected, err:", err)
	}
	log.Println("put() rows affected", rowsAffected)

	type Resp struct {
		Code int `json:"code"`
		Id uint64 `json:"id"`
	}
	_id, err := result.LastInsertId()
	if err != nil {
		log.Println("failed to get last insert id in put(), err:", err)
	}
	resp := Resp{
		Code: 200,
		Id: uint64(_id),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed in put(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in put(), err:", err)
	}
}

// get tool info using name, return [{id, name, descr}]
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
	q := fmt.Sprintf(`
SELECT id, name, descr FROM %s
WHERE name = ?
	`, toolTable)
	rows, err := database.Query(q, name.Name)
	defer rows.Close()
	if err != nil {
		log.Printf("select failed in get(), name: \"%s\". err: %s",
		name.Name, err)
	}
	var toolsWId []ToolWId
	var id uint64
	var _name string
	var descr string
	for rows.Next() {
		err = rows.Scan(&id, &_name, &descr)
		if err != nil {
			log.Println("scan failed in get(), err:", err)
		}
		toolsWId = append(toolsWId, ToolWId{Id: id, Name: _name, Descr: descr})
	}

	type Resp struct {
		Code int `json:"code"`
		ToolWId []ToolWId `json:"tools"`
	}
	resp := Resp {
		Code: 200,
		ToolWId: toolsWId,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed in get(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in get(), err:", err)
	}
	log.Printf("get(): got %d records\n", len(toolsWId))
}

// get tool info using id, return {name, descr}
func getUsingId(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read the request body")
	}
	id := Id{}
	err = json.Unmarshal(body, &id)
	if err != nil {
		log.Println("Unmarshal failed, err:", err)
	}
	q := fmt.Sprintf(`
SELECT name, descr FROM %s
WHERE id = ?
	`, toolTable)
	row := database.QueryRow(q, id.Id)
	tool := Tool{}
	recordNr := 1
	row.Scan(&tool.Name, &tool.Descr)
	if err != nil {
		if err == sql.ErrNoRows {
			recordNr = 0
		}
		log.Println("query failed in getUsingId(), err:", err)
	}

	type Resp struct {
		Code int `json:"code"`
		Tool Tool `json:"tool"`
	}
	resp := Resp {
		Code: 200,
		Tool: tool,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed in getUsingId(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in getUsingId(), err:", err)
	}
	log.Printf("getUsingId(): got %d record\n", recordNr)
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
	q := fmt.Sprintf(`
INSERT INTO %s(name, descr)
VALUES (?, ?)
	`, toolsetTable)
	result, err := database.Exec(q, toolset.Name, toolset.Descr)
	if err != nil {
		log.Printf("Insert failed in putToolset(), name: \"%s\", descr: \"%s\". err: %s",
		toolset.Name, toolset.Descr, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("failed to get rows affected, err:", err)
	}
	log.Println("putToolset() rows affected", rowsAffected)

	type Resp struct {
		Code int `json:"code"`
		Id uint64 `json:"id"`
	}
	_id, err := result.LastInsertId()
	if err != nil {
		log.Println("failed to get last insert id in putToolset(), err:", err)
	}
	resp := Resp{
		Code: 200,
		Id: uint64(_id),
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed in putToolset(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in putToolset(), err:", err)
	}
}

// get toolset info using name, return [{id, name, descr}]
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
	q := fmt.Sprintf(`
SELECT id, name, descr FROM %s
WHERE name = ?
	`, toolsetTable)
	rows, err := database.Query(q, name.Name)
	defer rows.Close()
	if err != nil {
		log.Printf("select failed in getToolset(), name: \"%s\". err: %s",
		name.Name, err)
	}
	var toolsetsWId []ToolsetWId
	var id uint64
	var _name string
	var descr string
	for rows.Next() {
		err = rows.Scan(&id, &_name, &descr)
		if err != nil {
			log.Println("scan failed in getToolset(), err:", err)
		}
		toolsetsWId = append(toolsetsWId, ToolsetWId{Id: id, Name: _name, Descr: descr})
	}

	type Resp struct {
		Code int `json:"code"`
		ToolsetWId []ToolsetWId `json:"toolsets"`
	}
	resp := Resp {
		Code: 200,
		ToolsetWId: toolsetsWId,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed in getToolset(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in getToolset(), err:", err)
	}
	log.Printf("getToolset(): got %d records\n", len(toolsetsWId))
}

// get toolset info using id, return {name, descr}
func getToolsetUsingId(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read the request body")
	}
	id := Id{}
	err = json.Unmarshal(body, &id)
	if err != nil {
		log.Println("Unmarshal failed, err:", err)
	}
	q := fmt.Sprintf(`
SELECT name, descr FROM %s
WHERE id = ?
	`, toolsetTable)
	row := database.QueryRow(q, id.Id)
	toolset := Toolset{}
	recordNr := 1
	err = row.Scan(&toolset.Name, &toolset.Descr)
	if err != nil {
		if err == sql.ErrNoRows {
			recordNr = 0
		}
		log.Println("query failed in getToolsetUsingId(), err:", err)
	}

	type Resp struct {
		Code int `json:"code"`
		Toolset Toolset `json:"toolset"`
	}
	resp := Resp {
		Code: 200,
		Toolset: toolset,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed in getToolsetUsingId(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in getToolsetUsingId(), err:", err)
	}
	log.Printf("getToolsetUsingId(): got %d record\n", recordNr)
}

func StartBackend(_db *sql.DB) {
	database = _db
	toolTable = "tool"
	toolsetTable = "toolset"
	db.CreateTablesIfNone(database, toolTable, toolsetTable)
	http.HandleFunc("/put", put)
	http.HandleFunc("/get", get)
	http.HandleFunc("/getid", getUsingId)
	http.HandleFunc("/putset", putToolset)
	http.HandleFunc("/getset", getToolset)
	http.HandleFunc("/getsetid", getToolsetUsingId)
	port := ":9160"
	fmt.Println("Serving backend on http://localhost" + port)
	http.ListenAndServe(port, nil)
}
