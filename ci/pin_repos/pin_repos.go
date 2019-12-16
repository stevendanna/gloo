package main

import (
	"github.com/solo-io/gloo/pkg/utils/gomodutils"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/go-utils/versionutils/git"
)

func main() {
	fatalCheck(PinVersions(new(gomodutils.ModParser), gomodutils.GlooVendoredDependencies), "error pinning versions")
}

// pin the given vendored dependencies to the version found in the mod file
func PinVersions(modParser *gomodutils.ModParser, vendoredDependencies *gomodutils.VendoredDependencies) error {
	gomodFile, err := modParser.Parse()
	if err != nil {
		return err
	}

	importPaths := vendoredDependencies.GetImportPaths()

	// collect the gomod versions of all those dependencies
	importPathToRequiredVersion := map[string]string{}
	for _, require := range gomodFile.Require {
		if importPaths.Has(require.Mod.Path) {
			importPathToRequiredVersion[require.Mod.Path] = require.Mod.Version
		}
	}

	// pin the repo to that version
	for importPath, modVersion := range importPathToRequiredVersion {
		relativePath := vendoredDependencies.GetPathRelativeToProject(importPath)
		fatalCheck(git.PinDependencyVersion(relativePath, modVersion), "consider git fetching in "+relativePath)
	}

	return nil
}

func fatalCheck(err error, msg string) {
	if err != nil {
		log.Fatalf("Error (%v) unable to pin repos!: %v", msg, err)
	}
}
