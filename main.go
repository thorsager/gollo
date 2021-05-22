package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	version = "*unset*"

	healthRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_total_health_requests",
		Help: "The total number of received health requests",
	})

	totalRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_total_requests",
		Help: "The total number of received requests",
	})

	hostname, message, bindAddr, port, prometheusPath, healthPath string
	dumpEnvironment, dumpHeaders                                  bool
)

func init() {
	var err error
	prometheus.MustRegister(healthRequests, totalRequests)
	prometheusPath = getEnvOrDflt("PROMETHEUS_PATH", "/actuator/prometheus")
	healthPath = getEnvOrDflt("HEALTH_PATH", "/actuator/health")
	hostname = getEnvOrDflt("HOSTNAME", "anonymous")
	message = getEnvOrDflt("GOLLO_MESSAGE", "Good day Sir.")
	bindAddr = getEnvOrDflt("SERVER_IP", "")
	port = getEnvOrDflt("SERVER_PORT", "8080")
	dumpHeaders, err = strconv.ParseBool(getEnvOrDflt("DUMP_HEADERS", "false"))
	if err != nil {
		log.Printf("WARNING: %s", err)
		dumpHeaders = false
	}
	dumpEnvironment, err = strconv.ParseBool(getEnvOrDflt("DUMP_ENVIRONMENT", "false"))
	if err != nil {
		log.Printf("WARNING: %s", err)
		dumpHeaders = false
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", logging(rootHandler()))
	mux.Handle(prometheusPath, logging(promhttp.Handler()))
	mux.Handle(healthPath, logging(actuatorHandler()))
	log.Printf("Starting Gollo Server v%s (%s) on port %s, header=%t, env=%t, metrics='%s', health='%s'",
		version, hostname, port, dumpHeaders, dumpEnvironment, prometheusPath, healthPath)
	err := http.ListenAndServe(bindAddr+":"+port, mux)
	if err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer totalRequests.Inc()
		startTime := time.Now()
		o := responseWriterObserver{ResponseWriter: w}
		next.ServeHTTP(&o, r)
		log.Printf("(%s) \"%s\" [%d] (%d) served to in %v", r.RemoteAddr, r.URL.Path, o.statusCode, o.size, time.Since(startTime))
	})
}

// getEnvOrDflt - Retrieves a name environment variable, if variable
// is not found the defaultValue is returned i stead
func getEnvOrDflt(name, defaultValue string) string {
	if value, found := os.LookupEnv(name); found {
		return value
	}
	return defaultValue
}
