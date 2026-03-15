package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeIPTestClient(server *httptest.Server) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: server.Client().Transport.(*http.Transport).DialContext,
		},
	}
}

func newIPServer(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
}

// patchIPClient replaces ipClient with one that always routes to server,
// and returns a restore function.
func patchIPClient(server *httptest.Server) func() {
	saved := ipClient
	ipClient = &http.Client{
		Transport: &singleHostTransport{
			base:    http.DefaultTransport,
			target:  server.URL,
			testCli: server.Client(),
		},
	}
	return func() { ipClient = saved }
}

// singleHostTransport rewrites every request to target so tests don't
// need real network access.
type singleHostTransport struct {
	base    http.RoundTripper
	target  string
	testCli *http.Client
}

func (t *singleHostTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.URL.Scheme = "http"
	r2.URL.Host = r.URL.Host // will be overridden below
	parsed, _ := http.NewRequest(r.Method, t.target+r.URL.Path, r.Body)
	return t.testCli.Do(parsed)
}

func TestDetectIPv4_ValidIP(t *testing.T) {
	srv := newIPServer(http.StatusOK, "1.2.3.4")
	defer srv.Close()
	defer patchIPClient(srv)()

	ip, err := detectIPv4()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ip.Equal(net.ParseIP("1.2.3.4")) {
		t.Errorf("ip = %v, want 1.2.3.4", ip)
	}
}

func TestDetectIPv4_Non200Status(t *testing.T) {
	srv := newIPServer(http.StatusServiceUnavailable, "service unavailable")
	defer srv.Close()
	defer patchIPClient(srv)()

	_, err := detectIPv4()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := err.Error(); len(got) == 0 {
		t.Error("error message is empty")
	}
}

func TestDetectIPv4_InvalidBody(t *testing.T) {
	srv := newIPServer(http.StatusOK, "not-an-ip")
	defer srv.Close()
	defer patchIPClient(srv)()

	_, err := detectIPv4()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDetectIPv64_ValidIPv6(t *testing.T) {
	srv := newIPServer(http.StatusOK, "2001:db8::1")
	defer srv.Close()
	defer patchIPClient(srv)()

	ip, err := detectIPv64()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip.To4() != nil {
		t.Errorf("expected IPv6, got IPv4-mappable address: %v", ip)
	}
}

func TestDetectIPv64_IPv4AddressRejected(t *testing.T) {
	srv := newIPServer(http.StatusOK, "1.2.3.4")
	defer srv.Close()
	defer patchIPClient(srv)()

	_, err := detectIPv64()
	if err == nil {
		t.Fatal("expected error for IPv4 response, got nil")
	}
	if err.Error() != "no IPv6 detected" {
		t.Errorf("error = %q, want %q", err.Error(), "no IPv6 detected")
	}
}

func TestDetectIPv64_Non200Status(t *testing.T) {
	srv := newIPServer(http.StatusServiceUnavailable, "service unavailable")
	defer srv.Close()
	defer patchIPClient(srv)()

	_, err := detectIPv64()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
