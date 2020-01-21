package main

import (
	"context"
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net"
	"os"

	"github.com/solo-io/gloo/pkg/version"
	"github.com/solo-io/go-utils/contextutils"

	"github.com/fsnotify/fsnotify"
	"google.golang.org/grpc"

	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	sds "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
)

const (
	sslKeyFile  = "/etc/envoy/ssl/tls.key"
	sslCertFile = "/etc/envoy/ssl/tls.crt"
	sslCaFile   = "/etc/envoy/ssl/tls.crt"
	sdsClient   = "sds_client"
)

var (
	sdsServerAddress = flag.String("sdsServerAddress", "127.0.0.1:8234", "The SDS server address.")
	key, cert, ca    []byte
	grpcOptions      = []grpc.ServerOption{grpc.MaxConcurrentStreams(1000000)}
)

type EnvoyKey struct{}

func (h *EnvoyKey) ID(node *core.Node) string {
	return sdsClient
}

func main() {
	flag.Parse()
	ctx := contextutils.WithLogger(context.Background(), "sds_server")
	ctx = contextutils.WithLoggerValues(ctx, "version", version.Version)

	// Set up the gRPC server
	snapshotCache, err := runGrpcServer(ctx) // runs the grpc server in internal goroutines
	if err != nil {
		contextutils.LoggerFrom(ctx).Info("%v", err)
	}

	key, err = ioutil.ReadFile(sslKeyFile)
	if err != nil {
		contextutils.LoggerFrom(ctx).Info("err: ", err)
	}
	cert, err = ioutil.ReadFile(sslCertFile)
	if err != nil {
		contextutils.LoggerFrom(ctx).Info("err: ", err)
	}
	ca, err = ioutil.ReadFile(sslCaFile)
	if err != nil {
		contextutils.LoggerFrom(ctx).Info("err: ", err)
	}
	updateSDSConfig(ctx, cert, key, cert, snapshotCache)

	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		contextutils.LoggerFrom(ctx).Warn("error when setting up file watcher: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				contextutils.LoggerFrom(ctx).Info("received event: \n", event)
				key, err = ioutil.ReadFile(sslKeyFile)
				if err != nil {
					contextutils.LoggerFrom(ctx).Info("err: ", err)
				}
				cert, err = ioutil.ReadFile(sslCertFile)
				if err != nil {
					contextutils.LoggerFrom(ctx).Info("err: ", err)
				}
				ca, err = ioutil.ReadFile(sslCaFile)
				if err != nil {
					contextutils.LoggerFrom(ctx).Info("err: ", err)
				}
				updateSDSConfig(ctx, cert, key, ca, snapshotCache)

				// watch for errors
			case err := <-watcher.errors:
				contextutils.LoggerFrom(ctx).Info("Received error: \n", err)
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(sslCertFile); err != nil {
		contextutils.LoggerFrom(ctx).Warn(fmt.Sprintf("error adding watch to file %v: %v", sslCertFile, err))
	}
	if err := watcher.Add(sslKeyFile); err != nil {
		contextutils.LoggerFrom(ctx).Warn(fmt.Sprintf("error adding watch to file %v: %v", sslKeyFile, err))
	}
	if err := watcher.Add(sslCaFile); err != nil {
		contextutils.LoggerFrom(ctx).Warn(fmt.Sprintf("error adding watch to file %v: %v", sslCaFile, err))
	}

	<-done
}

func runGrpcServer(ctx context.Context) (cache.SnapshotCache, error) {
	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems.
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", *sdsServerAddress)
	if err != nil {
		return nil, err
	}
	hasher := &EnvoyKey{}
	snapshotCache := cache.NewSnapshotCache(false, hasher, nil)
	svr := server.NewServer(context.Background(), snapshotCache, nil)

	// register services
	sds.RegisterSecretDiscoveryServiceServer(grpcServer, svr)

	contextutils.LoggerFrom(ctx).Info(fmt.Sprintf("sds server listening on %s\n", *sdsServerAddress))
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			os.Exit(1)
		}
	}()
	go func() {
		<-ctx.Done()
		contextutils.LoggerFrom(ctx).Info(fmt.Sprintf("stopping sds server on %s\n", *sdsServerAddress))
		grpcServer.GracefulStop()
	}()
	return snapshotCache, nil
}

func updateSDSConfig(ctx context.Context, cert, key, validation []byte, snapshotCache cache.SnapshotCache) {
	hash := fnv.New64()
	hash.Write(cert)
	hash.Write(key)
	hash.Write(validation)
	items := []cache.Resource{
		&auth.Secret{
			Name: "server_cert",
			Type: &auth.Secret_TlsCertificate{
				TlsCertificate: &auth.TlsCertificate{
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{
							InlineBytes: cert,
						},
					},
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{
							InlineBytes: key,
						},
					},
				},
			},
		},
		&auth.Secret{
			Name: "validation_context",
			Type: &auth.Secret_ValidationContext{
				ValidationContext: &auth.CertificateValidationContext{
					TrustedCa: &core.DataSource{
						Specifier: &core.DataSource_InlineBytes{
							InlineBytes: cert,
						},
					},
				},
			},
		},
	}
	secretSnapshot := cache.Snapshot{}
	version := fmt.Sprintf("%d", hash.Sum64())
	contextutils.LoggerFrom(ctx).Info(fmt.Sprintf("snapshot version is %s", version))
	secretSnapshot.Resources[cache.Secret] = cache.NewResources(version, items)

	snapshotCache.SetSnapshot(sdsClient, secretSnapshot)
}
