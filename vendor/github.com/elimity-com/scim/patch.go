package scim

import "github.com/scim2/filter-parser/v2"

const (
	// PatchOperationAdd is used to add a new attribute value to an existing resource.
	PatchOperationAdd = "add"
	// PatchOperationRemove removes the value at the target location specified by the required attribute "path".
	PatchOperationRemove = "remove"
	// PatchOperationReplace replaces the value at the target location specified by the "path".
	PatchOperationReplace = "replace"
)

// PatchOperation represents a single PATCH operation.
type PatchOperation struct {
	// Op indicates the operation to perform and MAY be one of "add", "remove", or "replace".
	Op string
	// Path contains an attribute path describing the target of the operation. The "path" attribute is OPTIONAL for
	// "add" and "replace" and is REQUIRED for "remove" operations.
	Path *filter.Path
	// Value specifies the value to be added or replaced.
	Value interface{}
}
