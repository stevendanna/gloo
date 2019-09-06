package translator

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

var _ = Describe("RouteOwner", func() {

	It("appends owners to the route metadata", func() {
		route := &gloov1.Route{}
		err := appendOwner(route, &v1.VirtualService{
			Metadata: core.Metadata{Name: "taco", Namespace: "pizza", Generation: 5},
		})
		Expect(err).NotTo(HaveOccurred())
		meta, err := getRouteMeta(route)
		Expect(err).NotTo(HaveOccurred())
		Expect(meta.Owners).To(Equal([]OwnerRef{{
			ResourceKind:       "*v1.VirtualService",
			ResourceRef:        core.ResourceRef{Namespace: "pizza", Name: "taco"},
			ObservedGeneration: 5,
		}}))
	})
})
