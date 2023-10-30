package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyberark/conjur-k8s-csi-provider/pkg/provider"
)

func main() {
	var s *provider.ConjurProviderServer
	var err error

	go func() {
		s, err = provider.NewServer()
		if err != nil {
			log.Fatalf("Failed to create CSI provider server %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	s.Stop()
}
