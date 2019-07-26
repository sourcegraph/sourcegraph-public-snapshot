package graphqlbackend

import "sourcegraph.com/pkg/db/dbtesting"

func init() {
	dbtesting.DBNameSuffix = "graphqlbackenddb"
}
