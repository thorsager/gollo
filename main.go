package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	version string

	healthRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_total_health_requests",
		Help: "The total number of received health requests",
	})

	totalRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_total_requests",
		Help: "The total number of received requests",
	})
)

func init() {
	prometheus.MustRegister(healthRequests, totalRequests)
}

func main() {
	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "anonymous"
	}
	message := os.Getenv("GOLLO_MESSAGE")
	if message == "" {
		message = "Good day sir."
	}
	bindAddr := os.Getenv("SERVER_IP")
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	dumpHeaders, err := strconv.ParseBool(os.Getenv("DUMP_HEADERS"))
	if err != nil {
		log.Printf("WARNING: %s", err)
		dumpHeaders = false
	}

	dumpEnvironment, err := strconv.ParseBool(os.Getenv("DUMP_ENVIRONMENT"))
	if err != nil {
		log.Printf("WARNING: %s", err)
		dumpHeaders = false
	}

	mux := http.NewServeMux()

	mux.Handle("/actuator/prometheus", promhttp.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		defer log.Printf("Request processed in %v", time.Now().Sub(startTime))
		defer totalRequests.Inc()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "(%s) [%v] I'm Gollo, running on \"%s\", I say to you \"%s\"\n",
			version,
			startTime.Format(time.RFC1123), hostname, message)

		if dumpHeaders {
			_, _ = fmt.Fprintf(w, "\n--- headers ---\n%s\n", formatRequest(r))
		}
		if dumpEnvironment {
			_, _ = fmt.Fprintf(w, "\n--- environment ---\n%s\n", strings.Join(os.Environ(), "\n"))
		}

	})

	mux.HandleFunc("/actuator/health", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		defer log.Printf("Health-check processed in %v", time.Now().Sub(startTime))
		defer healthRequests.Inc()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{ \"health\": \"100%\" }\n"))
	})

	log.Printf("Starting Server (%s) on port %s", hostname, port)
	err = http.ListenAndServe(bindAddr+":"+port, mux)
	if err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
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
	for name, _ := range r.Header {
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
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}
