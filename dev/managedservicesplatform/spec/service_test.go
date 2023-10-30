package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidIAMRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		ok   bool
	}{
		{
			name: "project custom role",
			role: "projects/project-id/roles/custom_role",
			ok:   true,
		},
		{
			name: "organization custom role",
			role: "organizations/org-id/roles/custom_role",
			ok:   true,
		},
		{
			name: "predefined role",
			role: "roles/iam.getIamPolicy",
			ok:   true,
		},
		{
			name: "invalid role",
			role: "invalid-role",
		},
		{
			name: "invalid role",
			role: "projects/project-id/roles/a-b-c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.ok, validIAMRole(tt.role))
		})
	}
}
