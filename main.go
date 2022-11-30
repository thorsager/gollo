package main

import (
	"crypto/subtle"
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
		dumpHeaders = false
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
	mux.Handle(basicPath, basicAuth(logging(pathHandler(basicPath)), credentialValidator(basicUser, basicPassword)))
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

func basicAuth(next http.Handler, isValidCredentials func(string, string) bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !isValidCredentials(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="gollo"`)
			w.WriteHeader(401)
			_, _ = w.Write([]byte("Unauthorized"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func credentialValidator(user, pass string) func(string, string) bool {
	validUser := []byte(user)
	validPassword := []byte(pass)
	return func(user, pass string) bool {
		return subtle.ConstantTimeCompare(validUser, []byte(user)) == 1 && subtle.ConstantTimeCompare(validPassword, []byte(pass)) == 1
	}

}

// getEnvOrDflt - Retrieves a name environment variable, if variable
// is not found the defaultValue is returned i stead
func getEnvOrDflt(name, defaultValue string) string {
	if value, found := os.LookupEnv(name); found {
		return value
	}
	return defaultValue
}
