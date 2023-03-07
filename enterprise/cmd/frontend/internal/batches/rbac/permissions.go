package rbac

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

var BatchChangesWritePermission = fmt.Sprintf("%s#WRITE", types.BatchChangesNamespace)
