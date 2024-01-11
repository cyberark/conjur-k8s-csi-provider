package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

const defaultPort int = 8080

// HealthServer is responsible for serving a /healthz endpoint over HTTP on a
// given port in order to report on the health of a ConjurProviderServer instance.
type HealthServer struct {
	port     int
	provider *ConjurProviderServer
	server   *http.Server
}

// NewHealthServer creates a new instance given a ConjurProviderServer instance
// with the default port and health check behavior.
func NewHealthServer(provider *ConjurProviderServer) *HealthServer {
	return newHealthServerWithDeps(
		provider,
		defaultPort,
		defaultHealthCheckFactory,
	)
}

func newHealthServerWithDeps(
	provider *ConjurProviderServer,
	port int,
	healthCheckFactory func(*ConjurProviderServer) func(http.ResponseWriter, *http.Request),
) *HealthServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthCheckFactory(provider))

	return &HealthServer{
		port:     port,
		provider: provider,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: mux,
		},
	}
}

// Start serves the HealthServer's HTTP server on the given port.
func (h *HealthServer) Start() error {
	log.Printf("Serving health server on port %d...\n", h.port)
	return h.server.ListenAndServe()
}

// Stop gracefully shuts down the HeathServer's HTTP server.
func (h *HealthServer) Stop() error {
	log.Println("Stopping health server...")

	err := h.server.Shutdown(context.TODO())
	if err == nil {
		log.Println("Health server stopped.")
	}

	return err
}

func defaultHealthCheckFactory(provider *ConjurProviderServer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		_, err := provider.versionFunc(context.TODO(), &v1alpha1.VersionRequest{
			Version: "health",
		})
		if err == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
