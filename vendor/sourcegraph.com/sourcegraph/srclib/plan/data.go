package plan

import (
	"path/filepath"

	"sourcegraph.com/sourcegraph/srclib/buildstore"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func RepositoryCommitDataFilename(emptyData interface{}) string {
	return buildstore.DataTypeSuffix(emptyData)
}

func SourceUnitDataFilename(emptyData interface{}, u *unit.SourceUnit) string {
	if u.Name == "" {
		return u.Type + "." + buildstore.DataTypeSuffix(emptyData)
	}
	return filepath.Join(u.Name, u.Type+"."+buildstore.DataTypeSuffix(emptyData))
}
