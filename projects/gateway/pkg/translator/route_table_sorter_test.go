package translator_test

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gateway/pkg/translator"
	"github.com/solo-io/gloo/projects/gloo/pkg/defaults"
	"github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

var _ = Describe("RouteTableSorter", func() {

	var (
		sorter translator.RouteTableSorter

		noWeight1,
		noWeight2,
		noWeight3,
		weightMinus10,
		weightZero,
		weightTen,
		weightTwenty *v1.RouteTable

		routeTableWithWeight = func(weight *int32, name string) *v1.RouteTable {
			var w *types.Int32Value
			if weight != nil {
				w = &types.Int32Value{Value: *weight}
			}

			table := &v1.RouteTable{
				Metadata: core.Metadata{
					Name:      name,
					Namespace: defaults.GlooSystem,
				},
				Weight: w,
				Routes: []*v1.Route{},
			}

			return table
		}
	)

	BeforeEach(func() {
		sorter = translator.NewRouteTableSorter()

		minus10 := int32(-10)
		zero := int32(0)
		ten := int32(10)
		twenty := int32(20)

		noWeight1 = routeTableWithWeight(nil, "no-weight-1")
		noWeight2 = routeTableWithWeight(nil, "no-weight-2")
		noWeight3 = routeTableWithWeight(nil, "no-weight-3")
		weightMinus10 = routeTableWithWeight(&minus10, "minus-ten")
		weightZero = routeTableWithWeight(&zero, "zero")
		weightTen = routeTableWithWeight(&ten, "ten")
		weightTwenty = routeTableWithWeight(&twenty, "twenty")
	})

	When("an empty list is passed", func() {
		It("returns nothing without errors", func() {
			haveBeenSorted, errs := sorter.Sort(nil)
			Expect(errs).To(BeNil())
			Expect(haveBeenSorted).To(BeFalse())
		})
	})

	When("a single route table is passed", func() {

		It("returns the table without errors", func() {
			haveBeenSorted, errs := sorter.Sort(v1.RouteTableList{noWeight1})
			Expect(errs).To(BeNil())
			Expect(haveBeenSorted).To(BeFalse())
		})
	})

	Context("multiple route tables are passed", func() {

		When("no route tables have weights", func() {
			It("does not sort the tables and returns the correct flag", func() {
				tables := v1.RouteTableList{noWeight1, noWeight2, noWeight3}
				haveBeenSorted, errs := sorter.Sort(tables)
				Expect(errs).To(BeNil())
				Expect(haveBeenSorted).To(BeFalse())
				Expect(tables).To(Equal(v1.RouteTableList{noWeight1, noWeight2, noWeight3}))
			})
		})

		When("all route tables have weights", func() {
			It("sorts the tables in ascending order by weight", func() {
				tables := v1.RouteTableList{weightTen, weightMinus10, weightTwenty, weightZero}
				haveBeenSorted, errs := sorter.Sort(tables)
				Expect(errs).To(BeNil())
				Expect(haveBeenSorted).To(BeTrue())
				Expect(tables).To(Equal(v1.RouteTableList{weightMinus10, weightZero, weightTen, weightTwenty}))
			})
		})

		When("some route tables have weights and others don't", func() {
			It("sorts the ones with weight in ascending order by weight and appends the rest", func() {
				tables := v1.RouteTableList{weightTen, noWeight1, weightMinus10, weightTwenty, noWeight3, weightZero, noWeight2}
				haveBeenSorted, errs := sorter.Sort(tables)

				Expect(errs).To(ConsistOf(testutils.HaveInErrorChain(
					translator.WithAndWithoutWeightErr(
						v1.RouteTableList{weightTen, weightMinus10, weightTwenty, weightZero},
						v1.RouteTableList{noWeight1, noWeight3, noWeight2},
					),
				)), "warns about likely user error")

				Expect(haveBeenSorted).To(BeTrue())
				Expect(tables[:4]).To(Equal(v1.RouteTableList{weightMinus10, weightZero, weightTen, weightTwenty}))
				Expect(tables[4:]).To(Equal(v1.RouteTableList{noWeight1, noWeight3, noWeight2}))
			})
		})

		When("some route tables have the same weight", func() {
			It("sorts the routes and warns about likely user error", func() {
				weightTenClone := proto.Clone(weightTen).(*v1.RouteTable)
				weightTenClone.Metadata.Name = "ten-dup"
				tables := v1.RouteTableList{weightZero, weightTen, weightTwenty, weightTenClone}

				haveBeenSorted, errs := sorter.Sort(tables)

				Expect(errs).To(ConsistOf(testutils.HaveInErrorChain(
					translator.RouteTablesWithSameWeightErr(v1.RouteTableList{weightTen, weightTenClone}, 10),
				)))

				Expect(haveBeenSorted).To(BeTrue())
				Expect(tables).To(Equal(v1.RouteTableList{weightZero, weightTenClone, weightTen, weightTwenty}))
			})
		})

	})
})
