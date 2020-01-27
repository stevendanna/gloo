package server_test

import (
	"context"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
	"github.com/solo-io/gloo/projects/sds/pkg/server"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SDS Server", func() {
	It("correctly updates SDSConfig", func() {
		ctx, _ := context.WithCancel(context.Background())
		hasher := &server.EnvoyKey{}
		snapshotCache := cache.NewSnapshotCache(false, hasher, nil)
		var key, cert, ca []byte
		server.UpdateSDSConfig(ctx, key, cert, ca, snapshotCache)
		_, err := snapshotCache.GetSnapshot(hasher.ID(nil))
		Expect(err).To(BeNil())
	})
})
