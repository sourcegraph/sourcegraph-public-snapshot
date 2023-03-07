package resolvers

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

var batchChangesCreatePermission = fmt.Sprintf("%s#WRITE", types.BatchChangesNamespace)
