package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

func pathHandler(path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintf(w, "(%s) [%v] I'm Gollo, running on \"%s\", I say to you \"%s\"\n",
			version,
			time.Now().Format(time.RFC1123), hostname, message)

		if dumpHeaders {
			_, _ = fmt.Fprintf(w, "\n--- headers ---\n%s\n", formatRequest(r))
		}
		if dumpEnvironment {
			_, _ = fmt.Fprintf(w, "\n--- environment ---\n%s\n", strings.Join(os.Environ(), "\n"))
		}
		if dumpPublicIp {
			_, _ = fmt.Fprint(w, "\n--- public ip ---\n")
			if ipv4, err := detectIPv4(); err != nil {
				_, _ = fmt.Fprintf(w, "%s\n", err)
			} else {
				_, _ = fmt.Fprintf(w, "public-facing IPv4: %s\n", ipv4)
			}
			if ipv6, err := detectIPv64(); err != nil {
				_, _ = fmt.Fprintf(w, "%s\n", err)
			} else {
				_, _ = fmt.Fprintf(w, "public-facing IPv6: %s\n", ipv6)
			}
		}
	})
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))

	var headerNames []string
	for name := range r.Header {
		headerNames = append(headerNames, name)
	}
	sort.Strings(headerNames)

	// Loop through headers
	for _, name := range headerNames {
		for _, h := range r.Header[name] {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" && r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
		_ = r.ParseForm()
		request = append(request, r.Form.Encode())
	} else {
		bodyBytes, _ := io.ReadAll(r.Body)
		request = append(request, string(bodyBytes))
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}
