package provider

import (
	"context"
	"fmt"
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

type ConjurProviderServer struct {
	grpcServer  grpcServer
	mountFunc   func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error)
	versionFunc func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error)
}

func NewServer() (*ConjurProviderServer, error) {
	return NewServerWithDeps(
		func(opt ...grpc.ServerOption) grpcServer { return grpc.NewServer(opt...) },
		net.Listen,
		Mount,
		Version,
	)
}

func NewServerWithDeps(
	grpcFactory func(...grpc.ServerOption) grpcServer,
	listenerFactory func(string, string) (net.Listener, error),
	mountFunc func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error),
	versionFunc func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error),
) (*ConjurProviderServer, error) {
	grpcServer := grpcFactory()
	listener, err := listenerFactory("unix", defaultSocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to start listener: %w", err)
	}

	providerServer := &ConjurProviderServer{
		grpcServer:  grpcServer,
		mountFunc:   mountFunc,
		versionFunc: versionFunc,
	}
	v1alpha1.RegisterCSIDriverProviderServer(grpcServer, providerServer)

	err = grpcServer.Serve(listener)
	if err != nil {
		return nil, fmt.Errorf("failed to serve gRPC on listener: %w", err)
	}

	return providerServer, nil
}

func (c *ConjurProviderServer) Stop() {
	c.grpcServer.GracefulStop()
}

func (c *ConjurProviderServer) Mount(ctx context.Context, req *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
	return c.mountFunc(ctx, req)
}

func (c *ConjurProviderServer) Version(ctx context.Context, req *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
	return c.versionFunc(ctx, req)
}
