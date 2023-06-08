package rbac

import rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"

// Schema refers to the RBAC structure which acts as a source of truth for permissions within
// the RBAC system.
type Schema struct {
	Namespaces          []Namespace                  `yaml:"namespaces"`
	ExcludeFromUserRole []rtypes.PermissionNamespace `yaml:"excludeFromUserRole"`
}

// Namespace represents a feature to be guarded by RBAC. (example: Batch Changes, Code Insights e.t.c)
type Namespace struct {
	Name    rtypes.PermissionNamespace `yaml:"name"`
	Actions []rtypes.NamespaceAction   `yaml:"actions"`
}
