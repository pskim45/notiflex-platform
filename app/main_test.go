package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type fakeIDStore struct {
	value uint64
	err   error
}

func (s *fakeIDStore) NextID(context.Context) (uint64, error) {
	if s.err != nil {
		return 0, s.err
	}
	s.value++
	return s.value, nil
}

func testAPI() *api {
	return &api{podName: "test-pod", ids: &fakeIDStore{}}
}

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	newHandler(testAPI()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body healthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status body = %q, want ok", body.Status)
	}
}

func TestVersion(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/version", nil)
	response := httptest.NewRecorder()
	newHandler(testAPI()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body versionResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Version != version {
		t.Fatalf("version = %q, want %q", body.Version, version)
	}
	if body.GoVersion == "" {
		t.Fatal("go_version is empty")
	}
	if body.Hostname != "test-pod" {
		t.Fatalf("hostname = %q, want test-pod", body.Hostname)
	}
}

func TestNextID(t *testing.T) {
	handler := newHandler(testAPI())

	for want := range 2 {
		request := httptest.NewRequest(http.MethodGet, "/id", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		var body idResponse
		if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		wantID := string(rune('1' + want))
		if body.ID != wantID || body.GeneratedBy != "test-pod" {
			t.Fatalf("response = %#v, want id=%s generated_by=test-pod", body, wantID)
		}
	}
}

func TestNextIDStoreFailure(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/id", nil)
	response := httptest.NewRecorder()
	service := &api{podName: "test-pod", ids: &fakeIDStore{err: errors.New("unavailable")}}
	newHandler(service).ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestValkeyPasswordFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "valkey-password")
	if err := os.WriteFile(path, []byte("file-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("VALKEY_PASSWORD_FILE", path)
	t.Setenv("VALKEY_PASSWORD", "environment-secret")

	password, err := valkeyPassword()
	if err != nil {
		t.Fatal(err)
	}
	if password != "file-secret" {
		t.Fatalf("password = %q, want file-secret", password)
	}
}

func TestValkeyPasswordEnvironmentFallback(t *testing.T) {
	t.Setenv("VALKEY_PASSWORD_FILE", "")
	t.Setenv("VALKEY_PASSWORD", "environment-secret")

	password, err := valkeyPassword()
	if err != nil {
		t.Fatal(err)
	}
	if password != "environment-secret" {
		t.Fatalf("password = %q, want environment-secret", password)
	}
}

func TestUnsupportedMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/id", nil)
	response := httptest.NewRecorder()
	newHandler(testAPI()).ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}
