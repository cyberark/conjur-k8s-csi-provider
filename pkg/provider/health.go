package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/logmessages"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

const DefaultPort int = 8080

// HealthServer is responsible for serving a /healthz endpoint over HTTP on a
// given port in order to report on the health of a ConjurProviderServer instance.
type HealthServer struct {
	port     int
	provider *ConjurProviderServer
	server   *http.Server
}

// NewHealthServer creates a new instance given a ConjurProviderServer instance
// with the default health check behavior.
func NewHealthServer(provider *ConjurProviderServer, port int) *HealthServer {
	return newHealthServerWithDeps(
		provider,
		port,
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
	log.Info(logmessages.CKCP024, h.port)
	return h.server.ListenAndServe()
}

// Stop gracefully shuts down the HeathServer's HTTP server.
func (h *HealthServer) Stop() error {
	log.Info(logmessages.CKCP025)

	err := h.server.Shutdown(context.TODO())
	if err == nil {
		log.Info(logmessages.CKCP026)
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
