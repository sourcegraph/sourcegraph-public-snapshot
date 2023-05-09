package rbac

import (
	"strings"

	"github.com/grafana/regexp"

	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var permissionDisplayNameRegex = regexp.MustCompile(`^\w+#\w+$`)

var invalidPermissionDisplayName = errors.New("permission display name is invalid.")

// ParsePermissionDisplayName parses a permission display name and returns the namespace and action.
// It returns an error if the displayName is invalid.
func ParsePermissionDisplayName(displayName string) (namespace rtypes.PermissionNamespace, action rtypes.NamespaceAction, err error) {
	if ok := permissionDisplayNameRegex.MatchString(displayName); ok {
		parts := strings.Split(displayName, "#")

		namespace = rtypes.PermissionNamespace(parts[0])
		action = rtypes.NamespaceAction(parts[1])
	} else {
		err = invalidPermissionDisplayName
	}
	return namespace, action, err
}
