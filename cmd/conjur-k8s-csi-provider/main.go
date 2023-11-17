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

	s = provider.NewServer()
	go func() {
		err = s.Start()
		if err != nil {
			log.Fatalf("Failed to start CSI provider server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	err = s.Stop()
	if err != nil {
		log.Fatalf("Failed to stop the CSI provider server: %v", err)
	}
}
