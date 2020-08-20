package main

import (
	"flag"
	"github.com/pivotal/dep-server/handler"
	"net/http"
)

func main() {
	var (
		addr  string
		s3URL string
	)

	flag.StringVar(&addr, "addr", "", "Address to listen on")
	flag.StringVar(&s3URL, "s3-url", "https://s3.amazonaws.com", "URL of AWS")
	flag.Parse()

	h := handler.Handler{S3URL: s3URL}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/dependency", h.DependencyHandler)

	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
