package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/pivotal/dep-server/handler"
)

func main() {
	var bucketURL string

	flag.StringVar(&bucketURL, "bucket-url", "https://deps.paketo.io", "URL of Metadata Bucket")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	h := handler.Handler{BucketURL: bucketURL}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/dependency", h.DependencyHandler)

	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		panic(err)
	}
}
