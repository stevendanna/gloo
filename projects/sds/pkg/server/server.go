package server

import (
	"context"
	"fmt"
	"hash/fnv"
	"io/ioutil"
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

func SetupEnvoySDS() (*grpc.Server, cache.SnapshotCache) {
	grpcServer := grpc.NewServer(grpcOptions...)
	hasher := &EnvoyKey{}
	snapshotCache := cache.NewSnapshotCache(false, hasher, nil)
	svr := server.NewServer(context.Background(), snapshotCache, nil)

	// register services
	sds.RegisterSecretDiscoveryServiceServer(grpcServer, svr)
	return grpcServer, snapshotCache
}

func RunSDSServer(ctx context.Context, grpcServer *grpc.Server) error {
	lis, err := net.Listen("tcp", sdsServerAddress)
	if err != nil {
		return err
	}
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
	return nil
}

func GetSnapshotVersion(sslKeyFile, sslCertFile, sslCaFile string) (string, error) {
	var err error
	key, err := ioutil.ReadFile(sslKeyFile)
	if err != nil {
		return "", err
	}
	cert, err := ioutil.ReadFile(sslCertFile)
	if err != nil {
		return "", err
	}
	ca, err := ioutil.ReadFile(sslCaFile)
	if err != nil {
		return "", err
	}
	hash := fnv.New64()
	hash.Write(key)
	hash.Write(cert)
	hash.Write(ca)
	snapshotVersion := fmt.Sprintf("%d", hash.Sum64())
	return snapshotVersion, nil
}

func UpdateSDSConfig(ctx context.Context, sslKeyFile, sslCertFile, sslCaFile string, snapshotCache cache.SnapshotCache) error {
	snapshotVersion, err := GetSnapshotVersion(sslKeyFile, sslCertFile, sslCaFile)
	if err != nil {
		return err
	}
	contextutils.LoggerFrom(ctx).Info(fmt.Sprintf("Updating SDS config. Snapshot version is %s", snapshotVersion))

	items := []cache.Resource{
		serverCertSecret(sslCertFile, sslKeyFile),
		validationContextSecret(sslCaFile),
	}
	secretSnapshot := cache.Snapshot{}
	secretSnapshot.Resources[cache.Secret] = cache.NewResources(snapshotVersion, items)
	return snapshotCache.SetSnapshot(sdsClient, secretSnapshot)
}

func serverCertSecret(certFile, keyFile string) cache.Resource {
	return &auth.Secret{
		Name: "server_cert",
		Type: &auth.Secret_TlsCertificate{
			TlsCertificate: &auth.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_Filename{
						Filename: certFile,
					},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_Filename{
						Filename: keyFile,
					},
				},
			},
		},
	}
}

func validationContextSecret(caFile string) cache.Resource {
	return &auth.Secret{
		Name: "validation_context",
		Type: &auth.Secret_ValidationContext{
			ValidationContext: &auth.CertificateValidationContext{
				TrustedCa: &core.DataSource{
					Specifier: &core.DataSource_Filename{
						Filename: caFile,
					},
				},
			},
		},
	}
}
