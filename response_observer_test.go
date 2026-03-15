package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewResponseWriterObserver_Defaults(t *testing.T) {
	o := newResponseWriterObserver(httptest.NewRecorder())
	if o.statusCode != http.StatusOK {
		t.Errorf("statusCode = %d, want 200", o.statusCode)
	}
	if o.wroteHeader {
		t.Error("wroteHeader should be false")
	}
	if o.size != 0 {
		t.Errorf("size = %d, want 0", o.size)
	}
}

func TestResponseWriterObserver_Write(t *testing.T) {
	o := newResponseWriterObserver(httptest.NewRecorder())

	n, err := o.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Errorf("n = %d, want 5", n)
	}
	if !o.wroteHeader {
		t.Error("wroteHeader should be true after Write")
	}
	if o.statusCode != http.StatusOK {
		t.Errorf("statusCode = %d, want 200", o.statusCode)
	}
	if o.size != 5 {
		t.Errorf("size = %d, want 5", o.size)
	}

	_, _ = o.Write([]byte(" world"))
	if o.size != 11 {
		t.Errorf("size = %d, want 11 after second write", o.size)
	}
}

func TestResponseWriterObserver_WriteHeader(t *testing.T) {
	o := newResponseWriterObserver(httptest.NewRecorder())

	o.WriteHeader(http.StatusCreated)
	if !o.wroteHeader {
		t.Error("wroteHeader should be true")
	}
	if o.statusCode != http.StatusCreated {
		t.Errorf("statusCode = %d, want 201", o.statusCode)
	}
}

func TestResponseWriterObserver_WriteImpliesStatusOK(t *testing.T) {
	rec := httptest.NewRecorder()
	o := newResponseWriterObserver(rec)

	_, _ = o.Write([]byte("data"))

	if rec.Code != http.StatusOK {
		t.Errorf("underlying recorder code = %d, want 200", rec.Code)
	}
}
