package registry

import "github.com/sourcegraph/sourcegraph/cmd/internal/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "registry"
}
