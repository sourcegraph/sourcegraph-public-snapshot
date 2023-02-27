package rbac

import "github.com/sourcegraph/sourcegraph/internal/types"

// Schema refers to the RBAC structure which acts as a source of truth for permissions within
// the RBAC system.
type Schema struct {
	Namespaces []Namespace `json:"namespaces"`
}

// Namespace represents a feature to be guarded by RBAC. (example: Batch Changes, Code Insights e.t.c)
type Namespace struct {
	Name    types.PermissionNamespace `json:"name"`
	Actions []string                  `json:"actions"`
}
