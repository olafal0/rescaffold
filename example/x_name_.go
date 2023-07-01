package main

import (
	"embed"
	"log"
	"net/http"
)

//go:embed web/index.html
var webFS embed.FS

var (
	listenPort = ":x_port_"
)

func main() {
	handler := http.FileServer(http.FS(webFS))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/web" + r.URL.Path
		handler.ServeHTTP(w, r)
	})
	log.Fatal(http.ListenAndServe(listenPort, nil))
}
