package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1"
)

type mockListener struct{}

func (l mockListener) Accept() (net.Conn, error) {
	return nil, nil
}

func (l mockListener) Addr() net.Addr {
	return nil
}

func (l mockListener) Close() error {
	return nil
}

type mockGrpc struct {
	stop            func()
	registerService func(*grpc.ServiceDesc, any)
	serve           func(net.Listener) error
}

func (g mockGrpc) GracefulStop() {
	g.stop()
}

func (g mockGrpc) RegisterService(sd *grpc.ServiceDesc, ss any) {
	g.registerService(sd, ss)
}

func (g mockGrpc) Serve(lis net.Listener) error {
	return g.serve(lis)
}

var stopped bool

func TestNewServerWithDeps(t *testing.T) {
	testCases := []struct {
		description     string
		grpcFactory     func(opt ...grpc.ServerOption) grpcServer
		listenerFactory func(string, string) (net.Listener, error)
		mountFunc       func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error)
		versionFunc     func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error)
		assertions      func(*testing.T, *ConjurProviderServer, error)
	}{
		{
			description: "listener factory fails",
			grpcFactory: func(opt ...grpc.ServerOption) grpcServer {
				return mockGrpc{
					stop:            func() {},
					registerService: func(sd *grpc.ServiceDesc, ss any) {},
					serve:           func(lis net.Listener) error { return nil },
				}
			},
			listenerFactory: func(string, string) (net.Listener, error) {
				return nil, errors.New("listener msg")
			},
			assertions: func(t *testing.T, c *ConjurProviderServer, err error) {
				assert.Equal(t, "failed to start listener: listener msg", err.Error())
			},
		},
		{
			description: "gRPC server fails on listener",
			grpcFactory: func(opt ...grpc.ServerOption) grpcServer {
				return mockGrpc{
					stop:            func() {},
					registerService: func(sd *grpc.ServiceDesc, ss any) {},
					serve: func(lis net.Listener) error {
						return errors.New("serve msg")
					},
				}
			},
			listenerFactory: func(string, string) (net.Listener, error) {
				return mockListener{}, nil
			},
			assertions: func(t *testing.T, c *ConjurProviderServer, err error) {
				assert.Equal(t, "failed to serve gRPC on listener: serve msg", err.Error())
			},
		},
		{
			description: "provider server calls custom mount and version functions",
			grpcFactory: func(opt ...grpc.ServerOption) grpcServer {
				return mockGrpc{
					stop:            func() {},
					registerService: func(sd *grpc.ServiceDesc, ss any) {},
					serve:           func(lis net.Listener) error { return nil },
				}
			},
			listenerFactory: func(string, string) (net.Listener, error) {
				return mockListener{}, nil
			},
			mountFunc: func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
				return nil, fmt.Errorf("custom mount error")
			},
			versionFunc: func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
				return nil, fmt.Errorf("custom version error")
			},
			assertions: func(t *testing.T, c *ConjurProviderServer, err error) {
				assert.Nil(t, err)

				_, err = c.Mount(context.TODO(), &v1alpha1.MountRequest{
					Attributes: "{}",
					Secrets:    "{}",
					Permission: "777",
					TargetPath: "/some/path",
				})
				assert.Contains(t, err.Error(), "custom mount error")

				_, err = c.Version(context.TODO(), &v1alpha1.VersionRequest{
					Version: "0.0.test",
				})
				assert.Contains(t, err.Error(), "custom version error")
			},
		},
		{
			description: "stopping the gRPC server",
			grpcFactory: func(opt ...grpc.ServerOption) grpcServer {
				return mockGrpc{
					stop:            func() { stopped = true },
					registerService: func(sd *grpc.ServiceDesc, ss any) {},
					serve:           func(lis net.Listener) error { return nil },
				}
			},
			listenerFactory: func(string, string) (net.Listener, error) {
				return mockListener{}, nil
			},
			assertions: func(t *testing.T, c *ConjurProviderServer, err error) {
				assert.Nil(t, err)

				stopped = false
				c.Stop()
				assert.True(t, stopped)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			p, err := NewServerWithDeps(
				tc.grpcFactory,
				tc.listenerFactory,
				tc.mountFunc,
				tc.versionFunc,
			)
			tc.assertions(t, p, err)
		})
	}
}
