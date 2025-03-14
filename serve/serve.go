package serve;

import (
	"net/http"
	"fmt"
)

func ServeHtml() {
	port := "3000"
	serveDir := "serve"
	fmt.Printf("Serving HTML at http://localhost:%s\n", port)
	http.Handle("/", http.FileServer(http.Dir(serveDir)))
	http.ListenAndServe(":" + port, nil)
}
