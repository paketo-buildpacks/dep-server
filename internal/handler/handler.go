package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	BucketURL string
}

func (h Handler) DependencyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.handlerError(w, http.StatusMethodNotAllowed, fmt.Sprintf("request method %s not supported", r.Method))
		return
	}

	dependencyName := r.URL.Query().Get("name")
	if dependencyName == "" {
		h.handlerError(w, http.StatusBadRequest, "must provide param 'name'")
		return
	}

	metadataFileURL := fmt.Sprintf("%s/metadata/%s.json", h.BucketURL, strings.ToLower(dependencyName))
	resp, err := http.Get(metadataFileURL)
	if err != nil {
		h.handlerError(w, http.StatusInternalServerError,
			fmt.Sprintf("error requesting dependency metadata: %s", err.Error()))
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		h.handlerError(w, http.StatusInternalServerError, "error getting dependency metadata")
		return
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		h.handlerError(w, http.StatusInternalServerError, fmt.Sprintf("error returning dependency metadata: %s", err.Error()))
		return
	}
}

func (h Handler) handlerError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, _ = fmt.Fprintf(w, `{"error": "%s"}`, message)
}
