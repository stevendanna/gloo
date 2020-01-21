package main

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/grpc"
	"hash/fnv"
	"io/ioutil"
	"net"

	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	sds "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
)

const (
	SslKeyFile  = "/etc/envoy/tls.key"
	SslCertFile = "/etc/envoy/tls.crt"
	SslCaFile = "/etc/envoy/tls.crt"
)

var (
	key, cert, ca []byte
)

type EnvoyKeyHasher struct{}

func (h *EnvoyKeyHasher) ID(node *core.Node) string {
	return node.GetId()
}

func main() {
	ctx := context.TODO()

	// Set up the gRPC server
	snapshotCache, err := runGrpcServer(ctx) // runs the grpc server in internal goroutines
	if err != nil {
		fmt.Printf("%v", err)
	}

	key, err = ioutil.ReadFile(SslKeyFile)
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	cert, err = ioutil.ReadFile(SslCertFile)
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	ca, err = ioutil.ReadFile(SslCaFile)
	if err != nil {
		fmt.Printf("err: %v", err)
	}
	updateSDSConfig(cert, key, cert, snapshotCache)

	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				fmt.Printf("EVENT! %#v\n", event)
				key, err = ioutil.ReadFile(SslKeyFile)
				if err != nil {
					fmt.Printf("err: %v", err)
				}
				cert, err = ioutil.ReadFile(SslCertFile)
				if err != nil {
					fmt.Printf("err: %v", err)
				}
				ca, err = ioutil.ReadFile(SslCaFile)
				if err != nil {
					fmt.Printf("err: %v", err)
				}
				updateSDSConfig(cert, key, ca, snapshotCache)

				// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(SslCertFile); err != nil {
		fmt.Println("ERROR", err)
	}
	if err := watcher.Add(SslKeyFile); err != nil {
		fmt.Println("ERROR", err)
	}
	if err := watcher.Add(SslCaFile); err != nil {
		fmt.Println("ERROR", err)
	}

	<-done
}

func runGrpcServer(ctx context.Context) (cache.SnapshotCache, error) {
	// gRPC golang library sets a very small upper bound for the number gRPC/h2
	// streams over a single TCP connection. If a proxy multiplexes requests over
	// a single connection to the management server, then it might lead to
	// availability problems.
	var grpcOptions []grpc.ServerOption
	grpcOptions = append(grpcOptions, grpc.MaxConcurrentStreams(1000000))
	grpcServer := grpc.NewServer(grpcOptions...)
	address := fmt.Sprintf("127.0.0.1:8234")

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	hasher := &EnvoyKeyHasher{}
	snapshotCache := cache.NewSnapshotCache(false, hasher, nil)
	svr := server.NewServer(context.Background(), snapshotCache, nil)

	// register services
	sds.RegisterSecretDiscoveryServiceServer(grpcServer, svr)

	fmt.Printf("sds server listening on %s\n", address)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			fmt.Println(err)
		}
	}()
	go func(){
		<-ctx.Done()
		fmt.Printf("Stopping sds server on %d\n", address)
		grpcServer.GracefulStop()
	}()
	return snapshotCache, nil
}

func updateSDSConfig(cert, key, validation [] byte, snapshotCache cache.SnapshotCache) {
	hash := fnv.New64()
	hash.Write(cert)
	hash.Write(key)
	hash.Write(validation)
	items := []cache.Resource{
		&auth.Secret{
			Name:"server_cert",
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
	secretSnapshot:=cache.Snapshot{}
	version := fmt.Sprintf("%d",hash.Sum64())
	fmt.Printf("Snapshot version is %s", version)
	secretSnapshot.Resources[cache.Secret] = cache.NewResources(version, items)

	snapshotCache.SetSnapshot("sds_client", secretSnapshot)
}