package rbac

import (
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var permissionDisplayNameRegex = regexp.MustCompile(`^\w+#\w+$`)

var invalidPermissionDisplayName = errors.New("permission display name is invalid.")

// ParsePermissionDisplayName parses a permission display name and returns the namespace and action.
// It returns an error if the displayName is invalid.
func ParsePermissionDisplayName(displayName string) (namespace types.PermissionNamespace, action string, err error) {
	if ok := permissionDisplayNameRegex.MatchString(displayName); ok {
		parts := strings.Split(displayName, "#")

		namespace = types.PermissionNamespace(parts[0])
		action = parts[1]
	} else {
		err = invalidPermissionDisplayName
	}
	return namespace, action, err
}
