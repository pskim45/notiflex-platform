package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	newHandler(&api{podName: "test-pod"}).ServeHTTP(response, request)

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

func TestNextID(t *testing.T) {
	handler := newHandler(&api{podName: "test-pod"})

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

func TestUnsupportedMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/id", nil)
	response := httptest.NewRecorder()
	newHandler(&api{podName: "test-pod"}).ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}
