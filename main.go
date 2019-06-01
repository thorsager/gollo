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

	mux := http.NewServeMux()

	mux.Handle("/actuator/prometheus", promhttp.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		defer log.Printf("Request processed in %v", time.Now().Sub(startTime))
		defer totalRequests.Inc()

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(
			fmt.Sprintf("(%s) [%v] Gollo, I'm %s and running a new version\n",
				version,
				startTime.Format(time.RFC1123), hostname)))
	})

	mux.HandleFunc("/actuator/health", func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		defer log.Printf("Health-check processed in %v", time.Now().Sub(startTime))
		defer healthRequests.Inc()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{ \"health\": \"100%\" }\n"))
	})

	log.Printf("Starting Server (%s)", hostname)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}
