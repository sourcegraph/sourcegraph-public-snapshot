package security

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TODO look at testcases here enterprise/cmd/frontend/internal/dotcom/productsubscription/license_expiration_test.go
func TestPasswordPolicy(t *testing.T) {
	//cfg := conf.Get()

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthMinPasswordLength: 9,
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				PasswordPolicy: &schema.PasswordPolicy{
					Enabled:                   true,
					MinimumLength:             10,
					NumberOfSpecialCharacters: 2,
					RequireUpperandLowerCase:  true,
					RequireAtLeastOneNumber:   true,
				},
			},
			AuthPasswordPolicy: &schema.AuthPasswordPolicy{
				Enabled:                   true,
				MinimumLength:             12,
				NumberOfSpecialCharacters: 2,
				RequireAtLeastOneNumber:   true,
				RequireUpperandLowerCase:  true,
			},
		},
	})

	t.Run("policy is parsed correctly", func(t *testing.T) {
		p := getPasswordPolicy()
		// policy is enabled
		assert.True(t, p.Enabled == true)

		// password length is from AuthPasswordPolicy
		assert.Equal(t, p.MinimumLength, 12)

	})

	// TODO t.Cleanup()
}
