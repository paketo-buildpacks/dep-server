package main

import (
	"flag"
	"github.com/pivotal/dep-server/handler"
	"net/http"
	"os"
)

func main() {
	var s3URL string

	flag.StringVar(&s3URL, "s3-url", "https://s3.amazonaws.com", "URL of AWS")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	h := handler.Handler{S3URL: s3URL}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/dependency", h.DependencyHandler)

	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
