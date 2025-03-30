package back

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// put tool (upsert)
// [{id, name, descr}] | [{name, dscr}] -> {code, msg, [ids]}
// the returned id is the updated one or the newly created one
func put_batch(w http.ResponseWriter, r *http.Request) {
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
	insertParam := make([]any, 0, len(toolsWId) / 2) // same as []interface{}
	updateToolsWId := make([]ToolWId, 0, len(toolsWId) / 2)
	for _, toolWId := range(toolsWId) {
		if toolWId.Id == 0 {
			if doInsert == false {
				doInsert = true
			} else {
				insert += ","
			}
			insert += "(?, ?)"
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
		result, err := database.Exec(insert, insertParam...)
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
		var rowsUpdated int64 := 0
		for _, toolWId := range(updateToolsWId) {
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
		Num:   rowsAffected,
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

func AddBatchHandlers(_db *sql.DB) {
	database = _db
	toolTable = "tool"
	toolsetTable = "toolset"
	toolsetRelTable = "toolset_rel"

	http.HandleFunc("/put-batch", put_batch)
	// http.HandleFunc("/getid-batch", getUsingId_batch)
	// http.HandleFunc("/del-batch", del_batch)
	//
	// http.HandleFunc("/toolset/put-batch", putToolset_batch)
	// http.HandleFunc("/toolset/getid-batch", getToolsetUsingId_batch)
	// http.HandleFunc("/toolset/del", del_batch)
	// http.HandleFunc("/toolset/gettool", getRelByToolset_batch)
	// http.HandleFunc("/toolset/addtool", addToolToSet_batch)
	// http.HandleFunc("/toolset/deltool", delToolFmSet_batch)
}
