package provider

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

const defaultSocketPath string = "/etc/kubernetes/secrets-store-csi-providers/conjur.sock"

type grpcServer interface {
	RegisterService(*grpc.ServiceDesc, any)
	Serve(net.Listener) error
	GracefulStop()
}

// ConjurProviderServer is an implementation of the v1alpha1.CSIDriverProviderServer
// interface.
type ConjurProviderServer struct {
	grpcServer  grpcServer
	listener    net.Listener
	mountFunc   func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error)
	versionFunc func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error)
}

// NewServer returns the default ConjurProviderServer struct.
func NewServer() *ConjurProviderServer {
	return newServerWithDeps(
		func(opt ...grpc.ServerOption) grpcServer { return grpc.NewServer(opt...) },
		Mount,
		Version,
	)
}

func newServerWithDeps(
	grpcFactory func(...grpc.ServerOption) grpcServer,
	mountFunc func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error),
	versionFunc func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error),
) *ConjurProviderServer {
	grpcServer := grpcFactory()
	providerServer := &ConjurProviderServer{
		grpcServer:  grpcServer,
		mountFunc:   mountFunc,
		versionFunc: versionFunc,
	}
	v1alpha1.RegisterCSIDriverProviderServer(grpcServer, providerServer)
	return providerServer
}

// Start serves the gRPC server on the default socket.
func (c *ConjurProviderServer) Start() error {
	return c.startWithDeps(net.Listen, defaultSocketPath)
}

func (c *ConjurProviderServer) startWithDeps(
	listenerFactory func(string, string) (net.Listener, error),
	socketPath string,
) error {
	var err error
	c.listener, err = listenerFactory("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	log.Println("Starting Conjur CSI Provider server...")
	return c.grpcServer.Serve(c.listener)
}

// Stop halts the gRPC server and closes the socket listener.
func (c *ConjurProviderServer) Stop() error {
	log.Println("Cleaning up Conjur CSI Provider server...")
	c.grpcServer.GracefulStop()

	err := c.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

func (c *ConjurProviderServer) Mount(ctx context.Context, req *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
	return c.mountFunc(ctx, req)
}

func (c *ConjurProviderServer) Version(ctx context.Context, req *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
	return c.versionFunc(ctx, req)
}
