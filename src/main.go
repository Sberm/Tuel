package main

import (
	"net/http"
	"fmt"
)

var port = "3000"

func serve_html() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":" + port, nil)
}

func main() {
	fmt.Printf("Serving HTML at http://localhost:%s\n", port)	
	serve_html()
}