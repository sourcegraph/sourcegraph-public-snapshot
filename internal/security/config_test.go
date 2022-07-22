package security

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/assert"
)

// setMockPasswordPolicyConfig helper for returning customized mock config
func setMockPasswordPolicyConfig(policyEnabled bool, authPolicyEnabled bool, authMinPasswordLength int,
	authPolicySpChr int, reqNumber bool, reqCase bool) {

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthMinPasswordLength: authMinPasswordLength,
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				PasswordPolicy: &schema.PasswordPolicy{
					Enabled:                   policyEnabled,
					NumberOfSpecialCharacters: 3,
					// invert reqNumber and reqCase so it differs AuthPasswordPolicy
					RequireUpperandLowerCase: !reqNumber,
					RequireAtLeastOneNumber:  !reqCase,
				},
			},
			AuthPasswordPolicy: &schema.AuthPasswordPolicy{
				Enabled:                   authPolicyEnabled,
				NumberOfSpecialCharacters: authPolicySpChr,
				RequireAtLeastOneNumber:   reqNumber,
				RequireUpperandLowerCase:  reqCase,
			},
		},
	})
}

func TestGetPasswordPolicy(t *testing.T) {

	authPolicyLength := 12
	authPolicySpChr := 2

	setMockPasswordPolicyConfig(false, true,
		authPolicyLength, authPolicySpChr, true, true)

	t.Run("Fetch correct policy.", func(t *testing.T) {
		p := conf.AuthPasswordPolicy()

		assert.True(t, p.Enabled)
		assert.Equal(t, p.MinimumLength, authPolicyLength)
		assert.Equal(t, p.RequireUpperandLowerCase, true)

		// create experimental policy for testing backwards compatability
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				AuthMinPasswordLength: authPolicyLength,
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					PasswordPolicy: &schema.PasswordPolicy{
						Enabled:                   true,
						NumberOfSpecialCharacters: authPolicySpChr,
						RequireUpperandLowerCase:  true,
						RequireAtLeastOneNumber:   true,
					},
				},
			},
		})

		p = conf.AuthPasswordPolicy()

		assert.True(t, p.Enabled)
		assert.Equal(t, p.MinimumLength, authPolicyLength)
		assert.Equal(t, p.RequireUpperandLowerCase, true)
		assert.Equal(t, p.NumberOfSpecialCharacters, authPolicySpChr)

	})
}

func TestFetchPasswordPolicy_nil(t *testing.T) {

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthMinPasswordLength: 9,
		},
	})

	t.Run("When no policy is defined, only check password length ", func(t *testing.T) {
		p := GetPasswordPolicy()
		assert.False(t, p.Enabled)

		assert.Nil(t, ValidatePassword("idontneedanythingspecial"))
		assert.ErrorContains(t, ValidatePassword("abshort"), "Your password may not be less than 9 or be more than 256 characters.")

	})
}
