package resourceid

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// ID represents a resource. All terraform resources that comprise a resource
// should use (ID).TerraformID(...) to construct an ID. The ID itself should
// never be used as a Terraform resource ID to avoid collisions.
type ID struct{ id string }

// New constructs a base ID for grouping resources.
//
// ❗ Inputs MUST NOT reference CDKTF values, as CDKTF values cannot be used as
// resource identifiers.
func New(id string) ID { return ID{id: id} }

// TerraformID must be used to construct IDs for Terraform resources within a
// resource group. The IDs it creates are automatically prefixed with the parent
// ID. Argument is required to safely create non-conflicting IDs.
//
// ❗ Inputs MUST NOT reference CDKTF values, as CDKTF values cannot be used as
// resource identifiers.
func (id ID) TerraformID(format string, args ...any) *string {
	subID := fmt.Sprintf(format, args...)
	return pointers.Ptr(fmt.Sprintf("%s-%s", id.id, subID))
}

// Group can be used by resource groups that use other resource groups to safely
// create non-conflicting sub-IDs.
//
// ❗ Inputs MUST NOT reference CDKTF values, as CDKTF values cannot be used as
// resource identifiers.
func (id ID) Group(format string, args ...any) ID {
	subID := fmt.Sprintf(format, args...)
	return ID{id: fmt.Sprintf("%s-%s", id.id, subID)}
}

// Append combines 2 IDs.
//
// ❗ Inputs MUST NOT reference CDKTF values, as CDKTF values cannot be used as
// resource identifiers.
func (id ID) Append(next ID) ID {
	return ID{id: fmt.Sprintf("%s-%s", id.id, next.id)}
}

// DisplayName can be used for display name fields - it is the ID itself, as
// display names generally do not need to be unique.
func (id ID) DisplayName() string { return id.id }

var _ fmt.Stringer = ID{}

// String must never be used to render ID as a string. This guards against
// accidental misuse.
func (ID) String() string { panic("resourceid.ID must never be rendered as string") }
