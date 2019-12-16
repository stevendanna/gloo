package gomodutils

import "k8s.io/apimachinery/pkg/util/sets"

// a set of dependencies imported to a project by vendoring
type VendoredDependencies struct {
	importPathToRelativePath map[string]string
}

// return a set of strings that look like `"github.com/solo-io/solo-kit"`
func (v *VendoredDependencies) GetImportPaths() sets.String {
	vendoredDependencyNames := sets.NewString()
	for k, _ := range v.importPathToRelativePath {
		vendoredDependencyNames.Insert(k)
	}

	return vendoredDependencyNames
}

// given an import path, return the relative path from the toplevel current project to that dependency
func (v *VendoredDependencies) GetPathRelativeToProject(importPath string) string {
	return v.importPathToRelativePath[importPath]
}

var GlooVendoredDependencies = &VendoredDependencies{
	importPathToRelativePath: map[string]string{
		"github.com/solo-io/solo-kit": "../solo-kit",
	},
}
