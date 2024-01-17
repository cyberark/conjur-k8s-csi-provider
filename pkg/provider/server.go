package provider

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"google.golang.org/grpc"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

const DefaultSocketPath string = "/var/run/secrets-store-csi-providers/conjur.sock"

type grpcServer interface {
	RegisterService(*grpc.ServiceDesc, any)
	Serve(net.Listener) error
	GracefulStop()
}

// ConjurProviderServer is an implementation of the v1alpha1.CSIDriverProviderServer
// interface.
type ConjurProviderServer struct {
	socketPath  string
	grpcServer  grpcServer
	listener    net.Listener
	mountFunc   func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error)
	versionFunc func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error)
}

// NewServer returns the default ConjurProviderServer struct.
func NewServer(socketPath string) *ConjurProviderServer {
	return newServerWithDeps(
		socketPath,
		func(opt ...grpc.ServerOption) grpcServer { return grpc.NewServer(opt...) },
		Mount,
		Version,
	)
}

func newServerWithDeps(
	socketPath string,
	grpcFactory func(...grpc.ServerOption) grpcServer,
	mountFunc func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error),
	versionFunc func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error),
) *ConjurProviderServer {
	log.Info("Creating and registering gRPC server...")
	validateSocket(socketPath)

	grpcServer := grpcFactory()
	providerServer := &ConjurProviderServer{
		socketPath:  socketPath,
		grpcServer:  grpcServer,
		mountFunc:   mountFunc,
		versionFunc: versionFunc,
	}
	v1alpha1.RegisterCSIDriverProviderServer(grpcServer, providerServer)
	return providerServer
}

// Start serves the gRPC server on the default socket.
func (c *ConjurProviderServer) Start() error {
	return c.startWithDeps(net.Listen, c.socketPath)
}

func (c *ConjurProviderServer) startWithDeps(
	listenerFactory func(string, string) (net.Listener, error),
	socketPath string,
) error {
	var err error
	c.listener, err = listenerFactory("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to start socket listener: %w", err)
	}

	log.Info("Serving gRPC server on socket %s...", socketPath)
	return c.grpcServer.Serve(c.listener)
}

// Stop halts the gRPC server and closes the socket listener.
func (c *ConjurProviderServer) Stop() {
	log.Info("Stopping gRPC server...")

	c.grpcServer.GracefulStop()

	log.Info("gRPC server stopped.")
}

func (c *ConjurProviderServer) Mount(ctx context.Context, req *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
	return c.mountFunc(ctx, req)
}

func (c *ConjurProviderServer) Version(ctx context.Context, req *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
	return c.versionFunc(ctx, req)
}

func validateSocket(path string) {
	dir := filepath.Dir(path)
	if !strings.HasPrefix(dir, "/var/run/secrets-store-csi-providers") &&
		!strings.HasPrefix(dir, "/etc/kubernetes/secrets-store-csi-providers") {

		log.Warn("Using non-standard providers directory %s: "+
			"Ensure this directory has been configured on your CSI Driver before proceeding",
			dir,
		)
	}
}
