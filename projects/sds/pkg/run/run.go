package run

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/solo-io/gloo/projects/sds/pkg/server"
	"github.com/solo-io/go-utils/contextutils"
)

const (
	sslKeyFile  = "/etc/envoy/ssl/tls.key"
	sslCertFile = "/etc/envoy/ssl/tls.crt"
	sslCaFile   = "/etc/envoy/ssl/tls.crt"
)

func Run(ctx context.Context, cancel context.CancelFunc) error {
	// Set up the gRPC server
	snapshotCache, err := server.RunSDSServer(ctx) // runs the grpc server in internal goroutines
	if err != nil {
		return err
	}

	// Initialize the SDS config
	key, cert, ca := readSecretsFromFile(ctx)
	server.UpdateSDSConfig(ctx, key, cert, ca, snapshotCache)

	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				contextutils.LoggerFrom(ctx).Info("received event: \n", event)
				key, cert, ca = readSecretsFromFile(ctx)
				server.UpdateSDSConfig(ctx, key, cert, ca, snapshotCache)

				// watch for errors
			case err := <-watcher.Errors:
				contextutils.LoggerFrom(ctx).Warn("Received error: \n", err)
			}
		}
	}()
	if err := watcher.Add(sslCertFile); err != nil {
		contextutils.LoggerFrom(ctx).Warn(fmt.Sprintf("error adding watch to file %v: %v", sslCertFile, err))
	}
	if err := watcher.Add(sslKeyFile); err != nil {
		contextutils.LoggerFrom(ctx).Warn(fmt.Sprintf("error adding watch to file %v: %v", sslKeyFile, err))
	}
	if err := watcher.Add(sslCaFile); err != nil {
		contextutils.LoggerFrom(ctx).Warn(fmt.Sprintf("error adding watch to file %v: %v", sslCaFile, err))
	}

	// Wire in signal handling
	done := make(chan bool, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	<-done
	cancel()
	return nil
}

func readSecretsFromFile(ctx context.Context) (key, cert, ca []byte) {
	var err error
	key, err = ioutil.ReadFile(sslKeyFile)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warn("err: ", err)
	}
	cert, err = ioutil.ReadFile(sslCertFile)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warn("err: ", err)
	}
	ca, err = ioutil.ReadFile(sslCaFile)
	if err != nil {
		contextutils.LoggerFrom(ctx).Warn("err: ", err)
	}
	return
}
