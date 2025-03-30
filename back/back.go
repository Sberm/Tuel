package back

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"tuel/db"
)

type Tool struct {
	Name  string `json:"name"`
	Descr string `json:"descr"` // description
}

type ToolWId struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Descr string `json:"descr"`
}

type Toolset struct {
	Name  string `json:"name"`
	Descr string `json:"descr"`
}

type ToolsetWId struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Descr string `json:"descr"`
}

type Name struct {
	Name string `json:"name"`
}

type Id struct {
	Id int64 `json:"id"`
}

type CodeMsg struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var database *sql.DB
var toolTable string
var toolsetTable string
var toolsetRelTable string

// put tool (upsert)
// {id, name, descr} | {name, dscr} -> {code, id}
// the returned id is the updated one or the newly created one
func put(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	toolWId := ToolWId{Id: -1}
	err = json.Unmarshal(body, &toolWId)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}
	var q string
	var result sql.Result
	if toolWId.Id != -1 {
		q = fmt.Sprintf(`
UPDATE %s
SET name = ?, descr = ?
WHERE id = ?
		`, toolTable)
		result, err = database.Exec(q, toolWId.Name, toolWId.Descr, toolWId.Id)
	} else {
		q = fmt.Sprintf(`
INSERT INTO %s(name, descr)
VALUES (?, ?)
		`, toolTable)
		result, err = database.Exec(q, toolWId.Name, toolWId.Descr)
	}
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("put failed in put(), id: \"%d\", name: \"%s\", descr: \"%s\". err: %s",
			toolWId.Id, toolWId.Name, toolWId.Descr, err)
	} else {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			// we need to set error because rowsAffected is needed for the logic
			code = 400
			msg = "get sql result failed"
			log.Println("failed to get rows affected, err:", err)
		}
		if rowsAffected == 0 {
			code = 400
			if toolWId.Id == -1 {
				msg = "insert failed"
			} else {
				msg = "no records found with this id"
			}
		}
		log.Println("put() rows affected", rowsAffected)
	}

	type Resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Id   int64  `json:"id"`
	}
	var _id int64
	if toolWId.Id == -1 {
		_id, err = result.LastInsertId()
		if err != nil {
			code = 400
			msg = "internal error"
			log.Println("failed to get last insert id in put(), err:", err)
		}
	} else {
		_id = toolWId.Id
	}
	resp := Resp{
		Code: code,
		Msg:  msg,
		Id:   _id,
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
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	name := Name{}
	err = json.Unmarshal(body, &name)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}
	q := fmt.Sprintf(`
SELECT id, name, descr FROM %s
WHERE name = ?
	`, toolTable)
	rows, err := database.Query(q, name.Name)
	defer rows.Close()
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("select failed in get(), name: \"%s\". err: %s",
			name.Name, err)
	}

	var toolsWId []ToolWId
	var id int64
	var _name string
	var descr string
	for rows.Next() {
		err = rows.Scan(&id, &_name, &descr)
		if err != nil {
			code = 400
			msg = "scan failed"
			log.Println("scan failed in get(), err:", err)
			break
		}
		toolsWId = append(toolsWId, ToolWId{Id: id, Name: _name, Descr: descr})
	}

	type Resp struct {
		Code    int       `json:"code"`
		Msg     string    `json:"msg"`
		ToolWId []ToolWId `json:"tools"`
	}
	resp := Resp{
		Code:    code,
		Msg:     msg,
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
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	id := Id{}
	err = json.Unmarshal(body, &id)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}
	q := fmt.Sprintf(`
SELECT name, descr FROM %s
WHERE id = ?
	`, toolTable)
	row := database.QueryRow(q, id.Id)
	tool := Tool{}
	recordNr := 1
	err = row.Scan(&tool.Name, &tool.Descr)
	if err != nil {
		code = 400
		msg = "scan failed"
		if err == sql.ErrNoRows {
			recordNr = 0
			msg = "no records found"
		} else {
			log.Println("query failed in getUsingId(), err:", err)
		}
	}

	type Resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Tool Tool   `json:"tool"`
	}
	type RespOnError struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	var data []byte
	if code == 200 {
		resp := Resp{
			Code: code,
			Msg:  msg,
			Tool: tool,
		}
		data, err = json.Marshal(resp)
	} else {
		resp := RespOnError{
			Code: code,
			Msg:  msg,
		}
		data, err = json.Marshal(resp)
	}
	if err != nil {
		log.Println("marshal failed in getUsingId(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in getUsingId(), err:", err)
	}
	log.Printf("getUsingId(): got %d record\n", recordNr)
}

// {"id": 1} -> {"code": 200}
// for tool and toolset
func del(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	path := r.URL.Path
	var tableName string
	if path == "/del" {
		tableName = "tool"
		log.Println("deleting tool")
	} else if path == "/toolset/del" {
		tableName = "toolset"
		log.Println("deleting toolset")
	} else {
		tableName = "tool"
		log.Println("don't know what to delete, so deleting tool")
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	id := Id{}
	err = json.Unmarshal(body, &id)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}
	q := fmt.Sprintf(`
DELETE FROM %s
WHERE id = ?
	`, tableName)
	result, err := database.Exec(q, id.Id)
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("Insert failed in del(), id: \"%d\". err: %s", id.Id, err)
	} else {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("failed to get rows affected, err:", err)
		}
		log.Println("del() rows affected", rowsAffected)
	}

	resp := CodeMsg{
		Code: code,
		Msg:  msg,
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

// put toolset (upsert)
// {id, name, descr} | {name, dscr} -> {code, id}
// the returned id is the updated one or the newly created one
func putToolset(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	toolsetWId := ToolsetWId{Id: -1}
	err = json.Unmarshal(body, &toolsetWId)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}
	var q string
	var result sql.Result
	if toolsetWId.Id != -1 {
		q = fmt.Sprintf(`
UPDATE %s
SET name = ?, descr = ?
WHERE id = ?
		`, toolsetTable)
		result, err = database.Exec(q, toolsetWId.Name, toolsetWId.Descr, toolsetWId.Id)
	} else {
		q = fmt.Sprintf(`
INSERT INTO %s(name, descr)
VALUES (?, ?)
		`, toolsetTable)
		result, err = database.Exec(q, toolsetWId.Name, toolsetWId.Descr)
	}
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("put failed in putToolset(), id: \"%d\", name: \"%s\", descr: \"%s\". err: %s",
			toolsetWId.Id, toolsetWId.Name, toolsetWId.Descr, err)
	} else {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("failed to get rows affected, err:", err)
		}
		log.Println("putToolset() rows affected", rowsAffected)
	}

	type Resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Id   int64  `json:"id"`
	}
	var _id int64
	if toolsetWId.Id == -1 {
		_id, err = result.LastInsertId()
		if err != nil {
			code = 400
			msg = "internal error"
			log.Println("failed to get last insert id in putToolset(), err:", err)
		}
	} else {
		_id = toolsetWId.Id
	}
	resp := Resp{
		Code: code,
		Msg:  msg,
		Id:   _id,
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
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	name := Name{}
	err = json.Unmarshal(body, &name)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}
	q := fmt.Sprintf(`
SELECT id, name, descr FROM %s
WHERE name = ?
	`, toolsetTable)
	rows, err := database.Query(q, name.Name)
	defer rows.Close()
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("select failed in getToolset(), name: \"%s\". err: %s",
			name.Name, err)
	}
	var toolsetsWId []ToolsetWId
	var id int64
	var _name string
	var descr string
	for rows.Next() {
		err = rows.Scan(&id, &_name, &descr)
		if err != nil {
			code = 400
			msg = "scan failed"
			log.Println("scan failed in getToolset(), err:", err)
			break
		}
		toolsetsWId = append(toolsetsWId, ToolsetWId{Id: id, Name: _name, Descr: descr})
	}

	type Resp struct {
		Code       int          `json:"code"`
		Msg        string       `json:"msg"`
		ToolsetWId []ToolsetWId `json:"toolsets"`
	}
	resp := Resp{
		Code:       code,
		Msg:        msg,
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
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	id := Id{}
	err = json.Unmarshal(body, &id)
	if err != nil {
		code = 400
		msg = "internal error"
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
		code = 400
		msg = "scan failed"
		if err == sql.ErrNoRows {
			recordNr = 0
			msg = "no records found"
		} else {
			log.Println("query failed in getToolsetUsingId(), err:", err)
		}
	}

	type Resp struct {
		Code    int     `json:"code"`
		Msg     string  `json:"msg"`
		Toolset Toolset `json:"toolset"`
	}
	type RespOnError struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	var data []byte
	if code == 200 {
		resp := Resp{
			Code:    code,
			Msg:     msg,
			Toolset: toolset,
		}
		data, err = json.Marshal(resp)
	} else {
		resp := RespOnError{
			Code: code,
			Msg:  msg,
		}
		data, err = json.Marshal(resp)
	}
	if err != nil {
		log.Println("marshal failed in getToolsetUsingId(), err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed in getToolsetUsingId(), err:", err)
	}
	log.Printf("getToolsetUsingId(): got %d record\n", recordNr)
}

type IdPair struct {
	ToolId    int64 `json:"tool_id"`
	ToolsetId int64 `json:"toolset_id"`
}

// add tool(s) to toolset {id} and return the status code
func addToolToSet(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	idPair := IdPair{}
	err = json.Unmarshal(body, &idPair)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	q := fmt.Sprintf(`
INSERT INTO %s(tool_id, toolset_id)
VALUES (?, ?)
	`, toolsetRelTable)
	result, err := database.Exec(q, idPair.ToolId, idPair.ToolsetId)
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("add tool to toolset failed tool_id: %d, toolset_id: %d, err: %s\n", idPair.ToolId, idPair.ToolsetId, err)
	} else {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("failed to get rows affected, err:", err)
		}
		log.Println("addToolToSet() rows affected", rowsAffected)
	}

	resp := CodeMsg{
		Code: code,
		Msg:  msg,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed, err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed, err:", err)
	}
}

// delete tool from toolset and return the status code
func delToolFmSet(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	idPair := IdPair{}
	err = json.Unmarshal(body, &idPair)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	q := fmt.Sprintf(`
DELETE FROM %s
WHERE tool_id = ? AND toolset_id = ?
	`, toolsetRelTable)
	result, err := database.Exec(q, idPair.ToolId, idPair.ToolsetId)
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("delete tool from toolset failed tool_id: %d, toolset_id: %d err: %s\n", idPair.ToolId, idPair.ToolsetId, err)
	} else {
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			log.Println("failed to get rows affected, err:", err)
		}
		log.Println("delToolFmSet() rows affected", rowsAffected)
	}

	resp := CodeMsg{
		Code: code,
		Msg:  msg,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed, err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed, err:", err)
	}
}

type IdsResp struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Ids  []int64 `json:"ids"`
}

// id -> [toolsets]
func getRelByTool(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	id := Id{}
	err = json.Unmarshal(body, &id)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	q := fmt.Sprintf(`
SELECT toolset_id FROM %s
WHERE tool_id = ?
	`, toolsetRelTable)
	rows, err := database.Query(q, id.Id)
	defer rows.Close()
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("sql query failed in getRelByTool(), err:", err)
	}

	var ids []int64
	var tmp int64
	for rows.Next() {
		err = rows.Scan(&tmp)
		if err != nil {
			code = 400
			msg = "scan failed"
			log.Println("scan failed, err:", err)
		}
		ids = append(ids, tmp)
	}

	resp := IdsResp{
		Code: code,
		Msg:  msg,
		Ids:  ids,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed, err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed, err:", err)
	}
	log.Printf("got %d records\n", len(ids))
}

// id -> [tools]
func getRelByToolset(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	id := Id{}
	err = json.Unmarshal(body, &id)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	q := fmt.Sprintf(`
SELECT toolset_id FROM %s
WHERE toolset_id = ?
	`, toolsetRelTable)
	rows, err := database.Query(q, id.Id)
	defer rows.Close()
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("sql query failed in getRelByToolset(), err:", err)
	}

	var ids []int64
	var tmp int64
	for rows.Next() {
		err = rows.Scan(&tmp)
		if err != nil {
			code = 400
			msg = "scan failed"
			log.Println("scan failed, err:", err)
		}
		ids = append(ids, tmp)
	}

	resp := IdsResp{
		Code: code,
		Msg:  msg,
		Ids:  ids,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed, err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response data failed, err:", err)
	}
	log.Printf("got %d records\n", len(ids))
}

func StartBackend(_db *sql.DB) {
	database = _db
	toolTable = "tool"
	toolsetTable = "toolset"
	toolsetRelTable = "toolset_rel"
	db.CreateTablesIfNone(database, toolTable, toolsetTable, toolsetRelTable)
	db.TurnOnForeignKey(database)
	http.HandleFunc("/put", put)
	http.HandleFunc("/get", get) // using name, basically useless
	http.HandleFunc("/getid", getUsingId)
	http.HandleFunc("/del", del)
	http.HandleFunc("/getset", getRelByTool) // array

	http.HandleFunc("/toolset/put", putToolset)
	http.HandleFunc("/toolset/get", getToolset) // using toolset name, basically useless
	http.HandleFunc("/toolset/getid", getToolsetUsingId)
	http.HandleFunc("/toolset/del", del)
	http.HandleFunc("/toolset/gettool", getRelByToolset) // returns an array
	http.HandleFunc("/toolset/addtool", addToolToSet)
	http.HandleFunc("/toolset/deltool", delToolFmSet)

	AddBatchHandlers(database)

	port := ":9160"
	fmt.Println("Serving backend on http://localhost" + port)
	http.ListenAndServe(port, nil)
}
