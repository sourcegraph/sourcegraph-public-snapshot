package bg

import (
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/secret"
)

func init() {
	dbtesting.DBNameSuffix = "bgdb"

	secret.MustInit()
}
