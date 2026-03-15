package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	hostname, message, bindAddr, port, prometheusPath, healthPath, basicUser, basicPassword, basicPath string
	dumpEnvironment, dumpHeaders, dumpPublicIp                                                         bool
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
	basicPath = getEnvOrDflt("BASIC_PATH", "/basic")
	basicUser = getEnvOrDflt("BASIC_USER", "gollo")
	basicPassword = getEnvOrDflt("BASIC_PASSWORD", "gollo")
	dumpHeaders, err = strconv.ParseBool(getEnvOrDflt("DUMP_HEADERS", "false"))
	if err != nil {
		log.Printf("WARNING: %s", err)
		dumpHeaders = false
	}
	dumpEnvironment, err = strconv.ParseBool(getEnvOrDflt("DUMP_ENVIRONMENT", "false"))
	if err != nil {
		log.Printf("WARNING: %s", err)
		dumpEnvironment = false
	}
	dumpPublicIp, err = strconv.ParseBool(getEnvOrDflt("DUMP_PUBLIC_IP", "false"))
	if err != nil {
		log.Printf("WARNING: %s", err)
		dumpPublicIp = false
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", logging(pathHandler("/")))
	mux.Handle(basicPath, basicAuth(logging(pathHandler(basicPath)), basicUser, basicPassword))
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

func basicAuth(next http.Handler, user, pass string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(u), []byte(user)) != 1 || subtle.ConstantTimeCompare([]byte(p), []byte(pass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="gollo"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
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
