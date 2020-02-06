package translator

import (
	"sort"
	"strings"

	errors "github.com/rotisserie/eris"
	v1 "github.com/solo-io/gloo/projects/gateway/pkg/api/v1"
)

var (
	RouteTablesWithSameWeightErr = func(tables v1.RouteTableList, weight int32) error {
		return errors.Errorf("the following route tables have the same weight (%d): [%s]. This can result in "+
			"unintended ordering of the resulting routes on the Proxy resource", weight, collectNames(tables))
	}
	WithAndWithoutWeightErr = func(with, without v1.RouteTableList) error {
		return errors.Errorf("some route tables define a weight and some do not; this can result in "+
			"unintended ordering of the resulting routes on the Proxy resource. Tables with weight: [%s], tables without "+
			"weight: [%s]", collectNames(with), collectNames(without))
	}
)

type RouteTableSorter interface {
	// Sorts the given route table list according to their weights (if present).
	// - `haveBeenSorted` indicates whether we were able to sort the routes.
	// - `errs` represent potential issues with the sorting result.
	Sort(routeTables v1.RouteTableList) (haveBeenSorted bool, errs []error)
}

func NewRouteTableSorter() RouteTableSorter {
	return &sorter{}
}

type sorter struct{}

func (sorter) Sort(routeTables v1.RouteTableList) (bool, []error) {

	// No need to sort if we do not have multiple tables
	if len(routeTables) <= 1 {
		return false, nil
	}

	// Separate tables that have weights from the ones that don't
	var withWeight, withoutWeight v1.RouteTableList
	for _, table := range routeTables {
		if weight := table.GetWeight(); weight != nil {
			withWeight = append(withWeight, table)
		} else {
			withoutWeight = append(withoutWeight, table)
		}
	}

	// If none of the tables have a weight, we have no way of sorting them
	if len(withoutWeight) == len(routeTables) {
		return false, nil
	}

	// Sort tables by weight. Tables with a weight are always "less" than tables without a weight; hence, route tables
	// without a weight will always be at the tail of the sorted slice (in no particular order).
	sort.SliceStable(routeTables, func(i, j int) bool {
		left, right := routeTables[i], routeTables[j]
		if left.Weight == nil {
			return false
		} else if right.Weight == nil {
			return true
		} else {
			return left.Weight.Value <= right.Weight.Value
		}
	})

	return true, validate(withWeight, withoutWeight)
}

func validate(withWeight, withoutWeight v1.RouteTableList) []error {
	var warnings []error

	// Index by weight
	withWeightMap := map[int32]v1.RouteTableList{}
	for _, rt := range withWeight {
		withWeightMap[rt.Weight.Value] = append(withWeightMap[rt.Weight.Value], rt)
	}

	// Warn if multiple tables have the same weight
	for weight, tablesForWeight := range withWeightMap {
		if len(tablesForWeight) > 1 {
			warnings = append(warnings, RouteTablesWithSameWeightErr(tablesForWeight, weight))
		}
	}

	// Warn if some tables have weight and others don't
	if len(withWeight) > 0 && len(withoutWeight) > 0 {
		warnings = append(warnings, WithAndWithoutWeightErr(withWeight, withoutWeight))
	}

	return warnings
}

func collectNames(routeTables v1.RouteTableList) string {
	var names []string
	for _, t := range routeTables {
		names = append(names, t.Metadata.Ref().Key())
	}
	return strings.Join(names, ", ")
}
