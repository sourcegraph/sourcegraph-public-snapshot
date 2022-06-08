package gitdomain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBranchName(t *testing.T) {
	for _, tc := range []struct {
		name   string
		branch string
		valid  bool
	}{
		{name: "Valid branch", branch: "valid-branch", valid: true},
		{name: "Valid branch with slash", branch: "rgs/valid-branch", valid: true},
		{name: "Valid branch with @", branch: "valid@branch", valid: true},
		{name: "Path component with .", branch: "valid-/.branch", valid: false},
		{name: "Double dot", branch: "valid..branch", valid: false},
		{name: "End with .lock", branch: "valid-branch.lock", valid: false},
		{name: "No space", branch: "valid branch", valid: false},
		{name: "No tilde", branch: "valid~branch", valid: false},
		{name: "No carat", branch: "valid^branch", valid: false},
		{name: "No colon", branch: "valid:branch", valid: false},
		{name: "No question mark", branch: "valid?branch", valid: false},
		{name: "No asterisk", branch: "valid*branch", valid: false},
		{name: "No open bracket", branch: "valid[branch", valid: false},
		{name: "No trailing slash", branch: "valid-branch/", valid: false},
		{name: "No beginning slash", branch: "/valid-branch", valid: false},
		{name: "No double slash", branch: "valid//branch", valid: false},
		{name: "No trailing dot", branch: "valid-branch.", valid: false},
		{name: "Cannot contain @{", branch: "valid@{branch", valid: false},
		{name: "Cannot be @", branch: "@", valid: false},
		{name: "Cannot contain backslash", branch: "valid\\branch", valid: false},
		{name: "head not allowed", branch: "head", valid: false},
		{name: "Head not allowed", branch: "Head", valid: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			valid := ValidateBranchName(tc.branch)
			assert.Equal(t, tc.valid, valid)
		})
	}
}
