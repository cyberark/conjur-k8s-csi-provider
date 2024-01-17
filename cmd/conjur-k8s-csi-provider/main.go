package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyberark/conjur-authn-k8s-client/pkg/log"
	"github.com/cyberark/conjur-k8s-csi-provider/pkg/provider"
)

func main() {
	exitCode := 0

	healthPort := flag.Int("healthPort", provider.DefaultPort, "Port to expose Conjur Provider health server")
	socketPath := flag.String("socketPath", provider.DefaultSocketPath, "Socket to expose Conjur Provider gRPC server")
	flag.Parse()

	var providerServer *provider.ConjurProviderServer
	providerErr := make(chan error)
	var healthServer *provider.HealthServer
	healthErr := make(chan error)

	providerServer = provider.NewServer(*socketPath)
	go func() {
		err := providerServer.Start()
		if err != nil {
			providerErr <- err
		}
	}()

	healthServer = provider.NewHealthServer(providerServer, *healthPort)
	go func() {
		err := healthServer.Start()
		if err != nil {
			healthErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-providerErr:
		log.Error("CSI provider server failed: %v", err)
		exitCode = 1
	case err := <-healthErr:
		log.Error("CSI provider health server failed: %v", err)
		exitCode = 1
	case <-stop:
	}

	err := healthServer.Stop()
	if err != nil {
		log.Error("Failed to stop the CSI provider health server: %v", err)
		exitCode = 1
	}

	providerServer.Stop()
	os.Exit(exitCode)
}
