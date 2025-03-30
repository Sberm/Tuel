package back

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// put tool (upsert)
// [{id, name, descr}] | [{name, dscr}] -> {code, msg, num}
func putBatch(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body")
	}
	var toolsWId []ToolWId
	err = json.Unmarshal(body, &toolsWId)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	doInsert := false
	doUpdate := false
	insert := fmt.Sprintf("INSERT INTO %s(name, descr) VALUES ", toolTable)
	update := fmt.Sprintf("UPDATE %s SET name = ?, descr = ? WHERE id = ?", toolTable)
	insertParam := make([]any, 0, len(toolsWId)/2) // same as []interface{}
	updateToolsWId := make([]ToolWId, 0, len(toolsWId)/2)
	insertClause := make([]string, 0, cap(insertParam))
	for _, toolWId := range toolsWId {
		if toolWId.Id == 0 {
			if doInsert == false {
				doInsert = true
			}
			insertClause = append(insertClause, "(?, ?)")
			insertParam = append(insertParam, toolWId.Name)
			insertParam = append(insertParam, toolWId.Descr)
		} else {
			if doUpdate == false {
				doUpdate = true
			}
			updateToolsWId = append(updateToolsWId, toolWId)
		}
	}

	var rowsAffected int64 = 0
	if doInsert {
		result, err := database.Exec(insert+strings.Join(insertClause, ","), insertParam...)
		if err != nil {
			log.Println("batch insert failed, err:", err)
		} else {
			_rowsAffected, err := result.RowsAffected()
			rowsAffected += _rowsAffected
			if err != nil {
				log.Println("failed to get rows affected, err:", err)
			} else {
				log.Println("inserted", _rowsAffected)
			}
		}
	}
	if doUpdate {
		var rowsUpdated int64 = 0
		for _, toolWId := range updateToolsWId {
			result, err := database.Exec(update, toolWId.Name, toolWId.Descr, toolWId.Id)
			if err != nil {
				log.Println("batch update failed, err:", err)
			} else {
				_rowsAffected, err := result.RowsAffected()
				rowsUpdated += _rowsAffected
				if err != nil {
					log.Println("failed to get rows affected, err:", err)
				}
			}
		}
		log.Println("updated", rowsUpdated)
		rowsAffected += rowsUpdated
	}

	type Resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Num  int64  `json:"num"`
	}
	resp := Resp{
		Code: code,
		Msg:  msg,
		Num:  rowsAffected,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed, err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response failed, err:", err)
	}
}

// [
//
//	{"id": 4}
//	{"id": 5}
//	{"id": 6}
//
// ]
//
//	or
//
// [4, 5, 6]
//
// [ids] -> [{name, descr}]
func getUsingIdBatch(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	var ids []Id
	err = json.Unmarshal(body, &ids)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	if len(ids) == 0 {
		code = 400
		msg = "null ids"
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
			log.Println("write response failed, err:", err)
		}
		return
	}

	q := fmt.Sprintf("SELECT id, name, descr FROM %s WHERE id = ?", toolTable)
	args := []any{ids[0].Id} // cannot be []int64
	for _, id := range ids[1:] {
		q += " or id = ?"
		args = append(args, id.Id)
	}

	rows, err := database.Query(q, args...)
	if err != nil {
		code = 400
		msg = "query failed"
		log.Println("batch select failed, err:", err, "ids", ids)
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
			log.Println("write response failed, err:", err)
		}
		return
	}

	tools := make([]ToolWId, 0, len(ids))
	var tool ToolWId
	for rows.Next() {
		rows.Scan(&tool.Id, &tool.Name, &tool.Descr)
		tools = append(tools, tool)
	}
	log.Printf("selected %d records\n", len(tools))

	type Resp struct {
		Code  int       `json:"code"`
		Msg   string    `json:"msg"`
		Tools []ToolWId `json:"tools"`
	}
	resp := Resp{
		Code:  code,
		Msg:   msg,
		Tools: tools,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed, err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response failed, err:", err)
	}
}

// [ids] -> {code, msg, num}
func delBatch(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body")
	}

	path := r.URL.Path
	var tableName string
	if path == "/del-batch" {
		tableName = "tool"
		log.Println("deleting tool")
	} else if path == "/toolset/del-batch" {
		tableName = "toolset"
		log.Println("deleting toolset")
	} else {
		tableName = "tool"
		log.Println("don't know what to delete, so deleting tool")
	}

	var ids []Id
	err = json.Unmarshal(body, &ids)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	if len(ids) == 0 {
		code = 400
		msg = "null ids"
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
			log.Println("write response failed, err:", err)
		}
		return
	}

	del := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	args := []any{ids[0].Id}
	for _, id := range ids[1:] {
		del += " or id = ?"
		args = append(args, id.Id)
	}
	result, err := database.Exec(del, args...)
	if err != nil {
		code = 400
		msg = "sql failed"
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
			log.Println("write response failed, err:", err)
		}
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("failed to get rows affected, err", err)
		rowsAffected = 0
	}

	type Resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Num  int64  `json:"num"`
	}
	resp := Resp{
		Code: code,
		Msg:  msg,
		Num:  rowsAffected,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed, err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response failed, err:", err)
	}
}

// put toolset (upsert)
// [{id, name, descr}] | [{name, dscr}] -> {code, msg, num}
func putToolsetBatch(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	var toolsetsWId []ToolsetWId
	err = json.Unmarshal(body, &toolsetsWId)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	doInsert := false
	doUpdate := false
	insert := fmt.Sprintf("INSERT INTO %s(name, descr) VALUES ", toolsetTable)
	update := fmt.Sprintf("UPDATE %s SET name = ?, descr = ? WHERE id = ?", toolsetTable)
	insertParam := make([]any, 0, len(toolsetsWId)/2) // same as []interface{}
	updateToolsetsWId := make([]ToolsetWId, 0, len(toolsetsWId)/2)
	insertClause := make([]string, 0, cap(insertParam))
	for _, toolsetWId := range toolsetsWId {
		if toolsetWId.Id == 0 {
			if doInsert == false {
				doInsert = true
			}
			insertClause = append(insertClause, "(?, ?)")
			insertParam = append(insertParam, toolsetWId.Name)
			insertParam = append(insertParam, toolsetWId.Descr)
		} else {
			if doUpdate == false {
				doUpdate = true
			}
			updateToolsetsWId = append(updateToolsetsWId, toolsetWId)
		}
	}

	var rowsAffected int64 = 0
	if doInsert {
		result, err := database.Exec(insert+strings.Join(insertClause, ","), insertParam...)
		if err != nil {
			log.Println("batch insert failed, err:", err)
		} else {
			_rowsAffected, err := result.RowsAffected()
			rowsAffected += _rowsAffected
			if err != nil {
				log.Println("failed to get rows affected, err:", err)
			} else {
				log.Println("inserted", _rowsAffected)
			}
		}
	}
	if doUpdate {
		var rowsUpdated int64 = 0
		for _, toolsetWId := range updateToolsetsWId {
			result, err := database.Exec(update, toolsetWId.Name, toolsetWId.Descr, toolsetWId.Id)
			if err != nil {
				log.Println("batch update failed, err:", err)
			} else {
				_rowsAffected, err := result.RowsAffected()
				rowsUpdated += _rowsAffected
				if err != nil {
					log.Println("failed to get rows affected, err:", err)
				}
			}
		}
		log.Println("updated", rowsUpdated)
		rowsAffected += rowsUpdated
	}

	type Resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Num  int64  `json:"num"`
	}
	resp := Resp{
		Code: code,
		Msg:  msg,
		Num:  rowsAffected,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed, err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response failed, err:", err)
	}
}

// [ids] -> [{name, descr}]
func getToolsetUsingIdBatch(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	var ids []Id
	err = json.Unmarshal(body, &ids)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	if len(ids) == 0 {
		code = 400
		msg = "null ids"
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
			log.Println("write response failed, err:", err)
		}
		return
	}

	q := fmt.Sprintf("SELECT id, name, descr FROM %s WHERE id = ?", toolsetTable)
	args := []any{ids[0].Id} // cannot be []int64
	for _, id := range ids[1:] {
		q += " or id = ?"
		args = append(args, id.Id)
	}

	rows, err := database.Query(q, args...)
	if err != nil {
		code = 400
		msg = "query failed"
		log.Println("batch select failed, err:", err, "ids", ids)
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
			log.Println("write response failed, err:", err)
		}
		return
	}

	toolsets := make([]ToolsetWId, 0, len(ids))
	var toolset ToolsetWId
	for rows.Next() {
		rows.Scan(&toolset.Id, &toolset.Name, &toolset.Descr)
		toolsets = append(toolsets, toolset)
	}
	log.Printf("selected %d records\n", len(toolsets))

	type Resp struct {
		Code     int          `json:"code"`
		Msg      string       `json:"msg"`
		Toolsets []ToolsetWId `json:"toolsets"`
	}
	resp := Resp{
		Code:     code,
		Msg:      msg,
		Toolsets: toolsets,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Println("marshal failed, err:", err)
	}
	_, err = w.Write(data)
	if err != nil {
		log.Println("write response failed, err:", err)
	}
}

// [{tool_id, toolset_id}] -> {code, msg, num}
func addToolToSetBatch(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	var idPairs []IdPair
	err = json.Unmarshal(body, &idPairs)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	q := fmt.Sprintf("INSERT INTO %s(tool_id, toolset_id) VALUES", toolsetRelTable)
	var args []any
	var values []string
	for _, idPair := range idPairs {
		args = append(args, idPair.ToolId)
		args = append(args, idPair.ToolsetId)
		values = append(values, "(?, ?)")
	}
	q += strings.Join(values, ",")
	var rowsAffected int64 = 0
	result, err := database.Exec(q, args...)
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("add tool to toolset failed, err:", err)
	} else {
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			log.Println("failed to get rows affected, err:", err)
		}
		log.Println("rows affected", rowsAffected)
	}

	type Resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Num  int64  `json:"num"`
	}
	resp := Resp{
		code, msg, rowsAffected,
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

// [ids] -> {code, msg, num}
func delToolFmSetBatch(w http.ResponseWriter, r *http.Request) {
	code := 200
	msg := "success"
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Failed to read the request body, err:", err)
	}
	var idPairs []IdPair
	err = json.Unmarshal(body, &idPairs)
	if err != nil {
		code = 400
		msg = "internal error"
		log.Println("Unmarshal failed, err:", err)
	}

	q := fmt.Sprintf("DELETE FROM %s WHERE", toolsetRelTable)
	var cond []string
	var args []any
	for _, idPair := range idPairs {
		cond = append(cond, "(tool_id = ? AND toolset_id = ?)")
		args = append(args, idPair.ToolId)
		args = append(args, idPair.ToolsetId)
	}
	q += strings.Join(cond, "OR")
	result, err := database.Exec(q, args...)
	var rowsAffected int64 = 0
	if err != nil {
		code = 400
		msg = "sql query failed"
		log.Printf("delete tool from toolset failed, err:", err)
	} else {
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			log.Println("failed to get rows affected, err:", err)
		}
		log.Println("delToolFmSet() rows affected", rowsAffected)
	}

	type Resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Num  int64  `json:"num"`
	}
	resp := Resp{
		code, msg, rowsAffected,
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

func AddBatchHandlers(_db *sql.DB) {
	database = _db
	toolTable = "tool"
	toolsetTable = "toolset"
	toolsetRelTable = "toolset_rel"

	http.HandleFunc("/put-batch", putBatch)
	http.HandleFunc("/getid-batch", getUsingIdBatch)
	http.HandleFunc("/del-batch", delBatch)

	http.HandleFunc("/toolset/put-batch", putToolsetBatch)
	http.HandleFunc("/toolset/getid-batch", getToolsetUsingIdBatch)
	http.HandleFunc("/toolset/del-batch", delBatch)
	http.HandleFunc("/toolset/addtool-batch", addToolToSetBatch)
	http.HandleFunc("/toolset/deltool-batch", delToolFmSetBatch)
}
