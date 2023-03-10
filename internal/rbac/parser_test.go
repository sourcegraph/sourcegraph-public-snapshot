package rbac

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestParsePermissionDisplayName(t *testing.T) {
	tests := []struct {
		displayName string

		namespace     types.PermissionNamespace
		action        string
		expectedError error

		name string
	}{
		{
			name:          "valid display name",
			displayName:   fmt.Sprintf("%s#READ", types.BatchChangesNamespace),
			namespace:     types.BatchChangesNamespace,
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
