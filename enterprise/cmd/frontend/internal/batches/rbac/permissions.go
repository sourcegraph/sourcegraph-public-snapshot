package rbac

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

// This is the permission required for the creation and execution of a Batch Change.
// The following resolvers require this permission to be assigned to the user accessing
// it:
// CreateBatchChange, ApplyBatchChange, CreateBatchSpec, CreateChangesetSpec, CreateChangesetSpecs
// CreateEmptyBatchChange, UpsertEmptyBatchChange, CreateBatchSpecFromRaw, UpsertBatchSpecInput,
// MoveBatchChange, DeleteBatchChange, CloseBatchChange, ExecuteBatchSpec, ReplaceBatchSpecInput,
// DeleteBatchSpec
var BatchChangesWritePermission = fmt.Sprintf("%s#WRITE", types.BatchChangesNamespace)
