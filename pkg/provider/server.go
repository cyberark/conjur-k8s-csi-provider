package provider

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/logmessages"
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
	log.Info(logmessages.CKCP018)
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
		log.Error(logmessages.CKCP020, err)
		return fmt.Errorf(logmessages.CKCP020, err)
	}

	log.Info(logmessages.CKCP021, socketPath)
	return c.grpcServer.Serve(c.listener)
}

// Stop halts the gRPC server and closes the socket listener.
func (c *ConjurProviderServer) Stop() {
	log.Info(logmessages.CKCP022)

	c.grpcServer.GracefulStop()

	log.Info(logmessages.CKCP023)
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

		log.Warn(logmessages.CKCP019, dir)
	}
}
