package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"
)

const version = "v0.3.0"

const idKey = "notiflex:id"

type idStore interface {
	NextID(context.Context) (uint64, error)
}

type valkeyIDStore struct {
	client valkey.Client
}

type api struct {
	podName string
	ids     idStore
}

type healthResponse struct {
	Status string `json:"status"`
}

type versionResponse struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	Hostname  string `json:"hostname"`
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

	ids, err := newValkeyIDStore(os.Getenv("VALKEY_ADDR"), os.Getenv("VALKEY_PASSWORD"))
	if err != nil {
		log.Fatal(err)
	}
	defer ids.Close()

	server := &http.Server{
		Addr:              ":8080",
		Handler:           newHandler(&api{podName: podName, ids: ids}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("notiflex API listening on %s (pod: %s)", server.Addr, podName)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func newValkeyIDStore(addr, password string) (*valkeyIDStore, error) {
	if addr == "" {
		return nil, fmt.Errorf("VALKEY_ADDR is required")
	}

	var lastErr error
	for attempt := 1; attempt <= 10; attempt++ {
		client, err := valkey.NewClient(valkey.ClientOption{
			InitAddress: []string{addr},
			Password:    password,
		})
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err = client.Do(ctx, client.B().Ping().Build()).Error()
			cancel()
			if err == nil {
				return &valkeyIDStore{client: client}, nil
			}
			client.Close()
		}

		lastErr = err
		log.Printf("Valkey 연결 재시도 %d/10: %v", attempt, err)
		if attempt < 10 {
			time.Sleep(3 * time.Second)
		}
	}

	return nil, fmt.Errorf("Valkey 연결 실패: %w", lastErr)
}

func (s *valkeyIDStore) NextID(ctx context.Context) (uint64, error) {
	id, err := s.client.Do(ctx, s.client.B().Incr().Key(idKey).Build()).AsInt64()
	if err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (s *valkeyIDStore) Close() {
	s.client.Close()
}

func newHandler(service *api) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", service.health)
	mux.HandleFunc("GET /version", service.version)
	mux.HandleFunc("GET /id", service.nextID)
	return mux
}

func (a *api) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, versionResponse{
		Version:   version,
		GoVersion: runtime.Version(),
		Hostname:  a.podName,
	})
}

func (a *api) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func (a *api) nextID(w http.ResponseWriter, r *http.Request) {
	id, err := a.ids.NextID(r.Context())
	if err != nil {
		log.Printf("generate ID: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "Valkey unavailable"})
		return
	}
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
