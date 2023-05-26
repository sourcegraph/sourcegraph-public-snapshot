package rbac

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
)

func TestParsePermissionDisplayName(t *testing.T) {
	tests := []struct {
		displayName string
		name        string

		namespace     rtypes.PermissionNamespace
		action        rtypes.NamespaceAction
		expectedError error
	}{
		{
			name:          "valid display name",
			displayName:   fmt.Sprintf("%s#READ", rtypes.BatchChangesNamespace),
			namespace:     rtypes.BatchChangesNamespace,
			action:        "READ",
			expectedError: nil,
		},
		{
			name:          "display name without action",
			displayName:   "BATCH_CHANGES#",
			namespace:     "",
			action:        "",
			expectedError: invalidPermissionDisplayName,
		},
		{
			name:          "display name without namespace",
			displayName:   "#READ",
			namespace:     "",
			action:        "",
			expectedError: invalidPermissionDisplayName,
		},
		{
			name:          "display name without namespace and action",
			displayName:   "#",
			namespace:     "",
			action:        "",
			expectedError: invalidPermissionDisplayName,
		},
		{
			name:          "empty string",
			displayName:   "",
			namespace:     "",
			action:        "",
			expectedError: invalidPermissionDisplayName,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ns, action, err := ParsePermissionDisplayName(tc.displayName)

			require.Equal(t, ns, tc.namespace)
			require.Equal(t, action, tc.action)
			require.Equal(t, err, tc.expectedError)
		})
	}
}
