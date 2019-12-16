package version

import (
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/utils/gomodutils"
	"github.com/solo-io/go-utils/log"
	gitutils "github.com/solo-io/go-utils/versionutils/git"
)

// ensure that the given vendored dependencies are at the version declared in go.mod
func CheckVersions(modParser *gomodutils.ModParser, vendoredDependencies *gomodutils.VendoredDependencies) error {
	modFile, err := modParser.Parse()
	if err != nil {
		return err
	}

	importPaths := vendoredDependencies.GetImportPaths()

	for _, require := range modFile.Require {
		if importPaths.Has(require.Mod.Path) {
			importPath := require.Mod.Path
			modVersion := require.Mod.Version
			pathRelativeToGloo := vendoredDependencies.GetPathRelativeToProject(importPath)

			log.Printf("Checking expected version for %s...", pathRelativeToGloo)

			actualVersion, err := gitutils.GetGitRefInfo(pathRelativeToGloo)
			if err != nil {
				return err
			}

			if actualVersion.Tag != modVersion && actualVersion.Branch != modVersion && actualVersion.Hash != modVersion {
				return errors.Errorf("Expected %s version %s, found %+v in repo", pathRelativeToGloo, modVersion, actualVersion)
			}
		}
	}

	return nil
}
