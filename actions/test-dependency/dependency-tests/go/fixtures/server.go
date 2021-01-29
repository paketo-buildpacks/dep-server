package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := ":" + os.Args[1]

	http.HandleFunc("/", HelloServer)
	_ = http.ListenAndServe(port, nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello, world!")
}
