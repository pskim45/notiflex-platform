package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

const version = "v0.1.1"

type api struct {
	podName string
	counter atomic.Uint64
}

type healthResponse struct {
	Status string `json:"status"`
}

type versionResponse struct {
	Version string `json:"version"`
}

type idResponse struct {
	ID          string `json:"id"`
	GeneratedBy string `json:"generated_by"`
}

func main() {
	podName, err := os.Hostname()
	if err != nil {
		podName = "unknown"
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           newHandler(&api{podName: podName}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("notiflex API listening on %s (pod: %s)", server.Addr, podName)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func newHandler(service *api) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", service.health)
	mux.HandleFunc("GET /version", service.version)
	mux.HandleFunc("GET /id", service.nextID)
	return mux
}

func (a *api) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, versionResponse{Version: version})
}

func (a *api) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func (a *api) nextID(w http.ResponseWriter, _ *http.Request) {
	id := a.counter.Add(1)
	writeJSON(w, http.StatusOK, idResponse{
		ID:          strconv.FormatUint(id, 10),
		GeneratedBy: a.podName,
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
