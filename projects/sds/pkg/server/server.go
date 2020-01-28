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

type TlsInfo struct {
	key  []byte
	cert []byte
	ca   []byte
}

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

func Sync(ctx context.Context, sslKeyFile, sslCertFile, sslCaFile string, snapshotCache cache.SnapshotCache) error {
	tls, err := ReadSecretsFromFiles(sslKeyFile, sslCertFile, sslCaFile)
	if err != nil {
		return err
	}
	err = UpdateSDSConfig(ctx, tls, snapshotCache)
	if err != nil {
		return err
	}
	return nil
}

func ReadSecretsFromFiles(sslKeyFile, sslCertFile, sslCaFile string) (TlsInfo, error) {
	var err error
	key, err := ioutil.ReadFile(sslKeyFile)
	if err != nil {
		return TlsInfo{}, err
	}
	cert, err := ioutil.ReadFile(sslCertFile)
	if err != nil {
		return TlsInfo{}, err
	}
	ca, err := ioutil.ReadFile(sslCaFile)
	if err != nil {
		return TlsInfo{}, err
	}
	return TlsInfo{
		key:  key,
		cert: cert,
		ca:   ca,
	}, nil
}

func UpdateSDSConfig(ctx context.Context, tls TlsInfo, snapshotCache cache.SnapshotCache) error {
	hash := fnv.New64()
	hash.Write(tls.key)
	hash.Write(tls.cert)
	hash.Write(tls.ca)
	snapshotVersion := fmt.Sprintf("%d", hash.Sum64())
	contextutils.LoggerFrom(ctx).Debug(fmt.Sprintf("snapshot snapshotVersion is %s", snapshotVersion))

	items := []cache.Resource{
		serverCertSecret(tls.cert, tls.key),
		validationContextSecret(tls.ca),
	}
	secretSnapshot := cache.Snapshot{}
	secretSnapshot.Resources[cache.Secret] = cache.NewResources(snapshotVersion, items)
	return snapshotCache.SetSnapshot(sdsClient, secretSnapshot)
}

func serverCertSecret(cert, key []byte) cache.Resource {
	return &auth.Secret{
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
	}
}

func validationContextSecret(ca []byte) cache.Resource {
	return &auth.Secret{
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
	}
}
