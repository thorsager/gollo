package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPathHandler_ExactPath(t *testing.T) {
	hostname = "test-host"
	message = "test-message"
	dumpHeaders = false
	dumpEnvironment = false
	dumpPublicIp = false

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	pathHandler("/").ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	body := w.Body.String()
	for _, want := range []string{"I'm Gollo", "test-host", "test-message"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestPathHandler_WrongPath(t *testing.T) {
	r := httptest.NewRequest("GET", "/other", nil)
	w := httptest.NewRecorder()
	pathHandler("/").ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestPathHandler_DumpHeaders(t *testing.T) {
	saved := dumpHeaders
	dumpHeaders = true
	defer func() { dumpHeaders = saved }()

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Test", "value")
	w := httptest.NewRecorder()
	pathHandler("/").ServeHTTP(w, r)

	body := w.Body.String()
	if !strings.Contains(body, "--- headers ---") {
		t.Error("body missing '--- headers ---'")
	}
	if !strings.Contains(body, "X-Test") {
		t.Error("body missing request header")
	}
}

func TestPathHandler_DumpEnvironment(t *testing.T) {
	saved := dumpEnvironment
	dumpEnvironment = true
	defer func() { dumpEnvironment = saved }()
	t.Setenv("GOLLO_TEST_SENTINEL", "sentinel-value")

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	pathHandler("/").ServeHTTP(w, r)

	body := w.Body.String()
	if !strings.Contains(body, "--- environment ---") {
		t.Error("body missing '--- environment ---'")
	}
	if !strings.Contains(body, "GOLLO_TEST_SENTINEL=sentinel-value") {
		t.Error("body missing sentinel env var")
	}
}

func TestFormatRequest_BasicGET(t *testing.T) {
	r := httptest.NewRequest("GET", "/foo", nil)
	r.Host = "example.com"
	r.Header.Set("X-Foo", "bar")

	out := formatRequest(r)

	for _, want := range []string{"GET /foo HTTP/1.1", "Host: example.com", "X-Foo: bar"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestFormatRequest_HeadersSorted(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Z-Header", "z")
	r.Header.Set("A-Header", "a")
	r.Header.Set("M-Header", "m")

	out := formatRequest(r)

	posA := strings.Index(out, "A-Header")
	posM := strings.Index(out, "M-Header")
	posZ := strings.Index(out, "Z-Header")

	if posA < 0 || posM < 0 || posZ < 0 {
		t.Fatal("not all headers present in output")
	}
	if !(posA < posM && posM < posZ) {
		t.Error("headers are not sorted alphabetically")
	}
}

func TestFormatRequest_POST_UrlEncoded(t *testing.T) {
	body := strings.NewReader("foo=bar&baz=qux")
	r := httptest.NewRequest("POST", "/", body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	out := formatRequest(r)

	if !strings.Contains(out, "foo=bar") {
		t.Error("output missing form values")
	}
}

func TestFormatRequest_POST_OtherBody(t *testing.T) {
	body := strings.NewReader(`{"key":"value"}`)
	r := httptest.NewRequest("POST", "/", body)
	r.Header.Set("Content-Type", "application/json")

	out := formatRequest(r)

	if !strings.Contains(out, `{"key":"value"}`) {
		t.Error("output missing raw body")
	}
}
