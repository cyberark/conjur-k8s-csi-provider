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

type mockListener struct {
	close func() error
}

func (l mockListener) Accept() (net.Conn, error) {
	return nil, nil
}

func (l mockListener) Addr() net.Addr {
	return nil
}

func (l mockListener) Close() error {
	return l.close()
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
		description string
		grpcFactory func(opt ...grpc.ServerOption) grpcServer
		mountFunc   func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error)
		versionFunc func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error)
		assertions  func(*testing.T, *ConjurProviderServer)
	}{
		{
			description: "provider server calls custom mount and version functions",
			grpcFactory: func(opt ...grpc.ServerOption) grpcServer {
				return mockGrpc{
					stop:            func() {},
					registerService: func(sd *grpc.ServiceDesc, ss any) {},
					serve:           func(lis net.Listener) error { return nil },
				}
			},
			mountFunc: func(context.Context, *v1alpha1.MountRequest) (*v1alpha1.MountResponse, error) {
				return nil, fmt.Errorf("custom mount error")
			},
			versionFunc: func(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.VersionResponse, error) {
				return nil, fmt.Errorf("custom version error")
			},
			assertions: func(t *testing.T, c *ConjurProviderServer) {
				var err error

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
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			p := newServerWithDeps(
				tc.grpcFactory,
				tc.mountFunc,
				tc.versionFunc,
			)
			tc.assertions(t, p)
		})
	}
}

func TestStart(t *testing.T) {
	testCases := []struct {
		description     string
		serveError      error
		listenerFactory func(string, string) (net.Listener, error)
		assertions      func(*testing.T, error)
	}{
		{
			description: "serving gRPC fails",
			serveError:  errors.New("serve msg"),
			listenerFactory: func(string, string) (net.Listener, error) {
				return mockListener{}, nil
			},
			assertions: func(t *testing.T, err error) {
				assert.Equal(t, "serve msg", err.Error())
			},
		},
		{
			description: "listener factory fails",
			serveError:  nil,
			listenerFactory: func(string, string) (net.Listener, error) {
				return nil, errors.New("listener msg")
			},
			assertions: func(t *testing.T, err error) {
				assert.Equal(t, "failed to start socket listener: listener msg", err.Error())
			},
		},
		{
			description: "happy path",
			serveError:  nil,
			listenerFactory: func(string, string) (net.Listener, error) {
				return mockListener{}, nil
			},
			assertions: func(t *testing.T, err error) {
				assert.Nil(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			grpcFactory := func(opt ...grpc.ServerOption) grpcServer {
				return mockGrpc{
					stop:            func() {},
					registerService: func(sd *grpc.ServiceDesc, ss any) {},
					serve: func(lis net.Listener) error {
						return tc.serveError
					},
				}
			}

			p := newServerWithDeps(grpcFactory, nil, nil)
			err := p.startWithDeps(tc.listenerFactory, "")
			tc.assertions(t, err)
		})
	}
}

func TestStop(t *testing.T) {
	testCases := []struct {
		description string
		assertions  func(*testing.T)
	}{
		{
			description: "happy path",
			assertions: func(t *testing.T) {
				assert.True(t, stopped)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			grpcFactory := func(opt ...grpc.ServerOption) grpcServer {
				return mockGrpc{
					stop:            func() { stopped = true },
					registerService: func(sd *grpc.ServiceDesc, ss any) {},
					serve:           func(lis net.Listener) error { return nil },
				}
			}
			listenerFactory := func(string, string) (net.Listener, error) {
				return mockListener{}, nil
			}

			p := newServerWithDeps(grpcFactory, nil, nil)
			err := p.startWithDeps(listenerFactory, "")
			assert.Nil(t, err)
			stopped = false

			p.Stop()
			tc.assertions(t)
		})
	}
}
