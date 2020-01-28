package server

import (
	"context"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_service_discovery_v2 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"google.golang.org/grpc"
)

var _ = Describe("SDS Server", func() {
	var test = []byte("test")
	tls := TlsInfo{
		key:  test,
		cert: test,
		ca:   test,
	}

	Context("ReadSecretsFromFiles", func() {
		var fs afero.Fs
		var dir string
		var err error

		BeforeEach(func() {
			fs = afero.NewOsFs()
			dir, err = afero.TempDir(fs, "", "")
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			_ = fs.RemoveAll(dir)
		})
		It("correctly reads tls secrets from files", func() {
			fileString := `test`
			keyFile, err := afero.TempFile(fs, dir, "")
			Expect(err).To(BeNil())
			_, err = keyFile.WriteString(fileString)
			Expect(err).To(BeNil())
			certFile, err := afero.TempFile(fs, dir, "")
			Expect(err).To(BeNil())
			_, err = certFile.WriteString(fileString)
			Expect(err).To(BeNil())
			caFile, err := afero.TempFile(fs, dir, "")
			Expect(err).To(BeNil())
			_, err = caFile.WriteString(fileString)
			Expect(err).To(BeNil())
			tls, err := ReadSecretsFromFiles(keyFile.Name(), certFile.Name(), caFile.Name())
			Expect(err).To(BeNil())
			Expect(tls).To(Equal(TlsInfo{
				key:  test,
				cert: []byte("test"),
				ca:   []byte("test"),
			}))
		})
	})

	Context("UpdateSDSConfig", func() {
		It("correctly updates SDSConfig", func() {
			ctx, _ := context.WithCancel(context.Background())
			hasher := &EnvoyKey{}
			snapshotCache := cache.NewSnapshotCache(false, hasher, nil)
			UpdateSDSConfig(ctx, tls, snapshotCache)
			_, err := snapshotCache.GetSnapshot(hasher.ID(nil))
			Expect(err).To(BeNil())
		})
	})

	Context("Test gRPC Server", func() {
		var ctx context.Context
		var cancel context.CancelFunc
		var grpcServer *grpc.Server
		var snapshotCache cache.SnapshotCache

		BeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())
			grpcServer, snapshotCache = SetupEnvoySDS()
			err := RunSDSServer(ctx, grpcServer)
			Expect(err).To(BeNil())
			//err = UpdateSDSConfig(ctx, tls, snapshotCache)
			//Expect(err).To(BeNil())
			_ = snapshotCache
		})

		AfterEach(func() {
			cancel()
		})

		It("accepts client connections", func() {
			// Check that it's answering
			var conn *grpc.ClientConn

			// Initiate a connection with the server
			conn, err := grpc.Dial("0.0.0.0:8234", grpc.WithInsecure())
			Expect(err).To(BeNil())
			defer conn.Close()

			client := envoy_service_discovery_v2.NewSecretDiscoveryServiceClient(conn)

			resp, err := client.FetchSecrets(ctx, &envoy_api_v2.DiscoveryRequest{
				VersionInfo: "",
				Node:        nil,
			})
			Expect(err).To(BeNil())
			Expect(len(resp.GetResources())).To(Equal(2))
		})
	})
})
