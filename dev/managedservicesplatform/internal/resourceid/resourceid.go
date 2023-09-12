package resourceid

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// ID represents a resource. All terraform resources that comprise a resource
// should use (ID).TerraformID(...) to construct an ID. The ID itself should
// never be used as a Terraform resource ID to avoid collisions.
type ID struct{ id string }

func New(id string) ID { return ID{id: id} }

// ResourceID must be used to construct IDs for Terraform resources within a
// resource group. The IDs it creates are automatically prefixed with the parent
// ID.
func (id ID) ResourceID(format string, args ...any) *string {
	subID := fmt.Sprintf(format, args...)
	return pointers.Ptr(fmt.Sprintf("%s-%s", id.id, subID))
}

// SubID can be used by resource groups that use other resource groups to safely
// create non-conflicting sub-IDs.
func (id ID) SubID(format string, args ...any) ID {
	subID := fmt.Sprintf(format, args...)
	return ID{id: fmt.Sprintf("%s-%s", id.id, subID)}
}

// DisplayName can be used for display name fields - it is the ID itself, as
// display names generally do not need to be unique.
func (id ID) DisplayName() string { return id.id }
