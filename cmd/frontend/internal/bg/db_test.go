package bg

import (
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	secretsPkg "github.com/sourcegraph/sourcegraph/internal/secrets"
)

func init() {
	dbtesting.DBNameSuffix = "bgdb"

	secretsPkg.MustInit()
}
