package spec

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
)

func TestEnvironmentPrivateAccessServerSpecValidate(t *testing.T) {
	for _, tc := range []struct {
		name           string
		spec           *EnvironmentPrivateAccessPerimeterSpec
		expectedErrors autogold.Value
	}{
		{
			name: "nil spec",
			spec: nil,
		},
		{
			name: "empty spec",
			spec: &EnvironmentPrivateAccessPerimeterSpec{},
		},
		{
			name: "valid identities",
			spec: &EnvironmentPrivateAccessPerimeterSpec{
				AllowlistedIdentities: []string{
					"serviceAccount:project-id@service-account.com",
					"user:user@example.com",
				},
			},
		},
		{
			name: "invalid identity",
			spec: &EnvironmentPrivateAccessPerimeterSpec{
				AllowlistedIdentities: []string{
					"invalid:identity",
				},
			},
			expectedErrors: autogold.Expect([]string{"allowlistedIdentities[0]: identity invalid:identity must be of one of the allowed types: [serviceAccount:, user:]"}),
		},
		{
			name: "mixed identities",
			spec: &EnvironmentPrivateAccessPerimeterSpec{
				AllowlistedIdentities: []string{
					"serviceAccount:project-id@service-account.com",
					"invalid:identity",
					"group:group@example.com",
				},
			},
			expectedErrors: autogold.Expect([]string{
				"allowlistedIdentities[1]: identity invalid:identity must be of one of the allowed types: [serviceAccount:, user:]",
				"allowlistedIdentities[2]: identity group:group@example.com must be of one of the allowed types: [serviceAccount:, user:]",
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.spec.Validate()
			if tc.expectedErrors == nil {
				assert.Empty(t, errs)
			} else {
				tc.expectedErrors.Equal(t, errorMessages(errs))
			}
		})
	}
}
