package main

import (
	"sync"
	"tuel/back"
	"tuel/db"
	"tuel/serve"
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
