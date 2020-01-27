package server

import (
	"context"
	"fmt"
	"hash/fnv"
	"net"
	"os"

	"github.com/solo-io/go-utils/contextutils"

	"google.golang.org/grpc"

	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	sds "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/envoyproxy/go-control-plane/pkg/server"
)

const (
	sdsClient        = "sds_client"
	sdsServerAddress = "0.0.0.0:8234"
)

var (
	grpcOptions = []grpc.ServerOption{grpc.MaxConcurrentStreams(10000)}
)

type EnvoyKey struct{}

func (h *EnvoyKey) ID(node *core.Node) string {
	return sdsClient
}

func RunSDSServer(ctx context.Context) (cache.SnapshotCache, error) {
	grpcServer := grpc.NewServer(grpcOptions...)

	lis, err := net.Listen("tcp", sdsServerAddress)
	if err != nil {
		return nil, err
	}
	hasher := &EnvoyKey{}
	snapshotCache := cache.NewSnapshotCache(false, hasher, nil)
	svr := server.NewServer(context.Background(), snapshotCache, nil)

	// register services
	sds.RegisterSecretDiscoveryServiceServer(grpcServer, svr)

	contextutils.LoggerFrom(ctx).Info(fmt.Sprintf("sds server listening on %s\n", sdsServerAddress))
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			contextutils.LoggerFrom(ctx).Error(fmt.Sprintf("Stopping sds server listening on %s\n", sdsServerAddress))
			os.Exit(1)
		}
	}()
	go func() {
		<-ctx.Done()
		contextutils.LoggerFrom(ctx).Info(fmt.Sprintf("stopping sds server on %s\n", sdsServerAddress))
		grpcServer.GracefulStop()
	}()
	return snapshotCache, nil
}

func UpdateSDSConfig(ctx context.Context, key, cert, ca []byte, snapshotCache cache.SnapshotCache) {
	hash := fnv.New64()
	hash.Write(cert)
	hash.Write(key)
	hash.Write(ca)
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
							InlineBytes: ca,
						},
					},
				},
			},
		},
	}
	secretSnapshot := cache.Snapshot{}
	snapshotVersion := fmt.Sprintf("%d", hash.Sum64())
	contextutils.LoggerFrom(ctx).Debug(fmt.Sprintf("snapshot snapshotVersion is %s", snapshotVersion))
	secretSnapshot.Resources[cache.Secret] = cache.NewResources(snapshotVersion, items)

	snapshotCache.SetSnapshot(sdsClient, secretSnapshot)
}
