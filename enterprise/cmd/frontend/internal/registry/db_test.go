package registry

import "sourcegraph.com/pkg/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "registry"
}
