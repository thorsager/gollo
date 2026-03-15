package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEnvOrDflt(t *testing.T) {
	t.Run("unset returns default", func(t *testing.T) {
		t.Setenv("GOLLO_TEST_UNSET", "")
		// ensure it's truly unset by not setting it
		if got := getEnvOrDflt("GOLLO_TEST_DEFINITELY_NOT_SET_XYZ", "default"); got != "default" {
			t.Errorf("got %q, want %q", got, "default")
		}
	})

	t.Run("set returns value", func(t *testing.T) {
		t.Setenv("GOLLO_TEST_VAR", "myvalue")
		if got := getEnvOrDflt("GOLLO_TEST_VAR", "default"); got != "myvalue" {
			t.Errorf("got %q, want %q", got, "myvalue")
		}
	})

	t.Run("set to empty returns empty not default", func(t *testing.T) {
		t.Setenv("GOLLO_TEST_EMPTY", "")
		if got := getEnvOrDflt("GOLLO_TEST_EMPTY", "default"); got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})
}

func TestCredentialValidator(t *testing.T) {
	validate := credentialValidator("alice", "secret")

	cases := []struct {
		user, pass string
		want       bool
	}{
		{"alice", "secret", true},
		{"alice", "wrong", false},
		{"wrong", "secret", false},
		{"wrong", "wrong", false},
		{"", "", false},
	}
	for _, c := range cases {
		if got := validate(c.user, c.pass); got != c.want {
			t.Errorf("validate(%q, %q) = %v, want %v", c.user, c.pass, got, c.want)
		}
	}
}

func TestCredentialValidator_EmptyCredentials(t *testing.T) {
	validate := credentialValidator("", "")
	if !validate("", "") {
		t.Error("empty credentials should match empty validator")
	}
}

func TestBasicAuth_NoCredentials(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	handler := basicAuth(inner, credentialValidator("user", "pass"))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if got := w.Header().Get("WWW-Authenticate"); got != `Basic realm="gollo"` {
		t.Errorf("WWW-Authenticate = %q", got)
	}
	if w.Body.String() != "Unauthorized" {
		t.Errorf("body = %q, want %q", w.Body.String(), "Unauthorized")
	}
	if called {
		t.Error("inner handler should not have been called")
	}
}

func TestBasicAuth_WrongCredentials(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
	handler := basicAuth(inner, credentialValidator("user", "pass"))

	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("wrong", "wrong")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	if called {
		t.Error("inner handler should not have been called")
	}
}

func TestBasicAuth_CorrectCredentials(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := basicAuth(inner, credentialValidator("user", "pass"))

	r := httptest.NewRequest("GET", "/", nil)
	r.SetBasicAuth("user", "pass")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !called {
		t.Error("inner handler should have been called")
	}
}

func TestLogging_PassesThrough(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	})
	handler := logging(inner)

	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if w.Body.String() != "hello" {
		t.Errorf("body = %q, want %q", w.Body.String(), "hello")
	}
}
