package main

import (
	"tuel/db"
	"tuel/serve"
	"tuel/back"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go serve.ServeHtml()
	db := db.ConnectDB()
	defer db.Close()
	back.StartBackend(db)
	wg.Wait()
}
