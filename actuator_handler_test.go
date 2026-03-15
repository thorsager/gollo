package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestActuatorHandler_Response(t *testing.T) {
	r := httptest.NewRequest("GET", "/actuator/health", nil)
	w := httptest.NewRecorder()
	actuatorHandler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	if got := w.Body.String(); got != "{ \"health\": \"100%\" }\n" {
		t.Errorf("body = %q", got)
	}
}
