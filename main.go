package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
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

	mux := http.NewServeMux()

	mux.Handle("/actuator/prometheus", promhttp.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		defer log.Printf("Request processed in %v", time.Now().Sub(startTime))
		defer totalRequests.Inc()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(
			fmt.Sprintf("(%s) [%v] I'm Gollo, running on \"%s\", I say to you \"%s\"\n",
				version,
				startTime.Format(time.RFC1123), hostname, message)))
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
	err := http.ListenAndServe(bindAddr+":"+port, mux)
	if err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}
