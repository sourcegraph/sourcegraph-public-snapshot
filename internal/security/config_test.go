package security

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func setMockPasswordPolicyConfig(policyEnabled bool, authPolicyEnabled bool, authMinPasswordLength int,
	policyMinLength int, authPolicyLength int, authPolicySpChr int, reqNumber bool, reqCase bool) {

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthMinPasswordLength: authMinPasswordLength,
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				PasswordPolicy: &schema.PasswordPolicy{
					Enabled:                   policyEnabled,
					NumberOfSpecialCharacters: authPolicySpChr,
					RequireUpperandLowerCase:  reqNumber,
					RequireAtLeastOneNumber:   reqCase,
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

	setMockPasswordPolicyConfig(false, true, 10, 12, authPolicyLength,
		2, true, true)

	t.Run("Policy retrieved is correct.", func(t *testing.T) {
		p := getPasswordPolicy()

		assert.True(t, p.Enabled == true)
		assert.Equal(t, p.MinimumLength, authPolicyLength)

	})
}

func TestFetchPasswordPolicy_nil(t *testing.T) {

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthMinPasswordLength: 9,
		},
	})

	t.Run("When no policy is defined, only check password length ", func(t *testing.T) {
		p := getPasswordPolicy()
		assert.True(t, p.Enabled == false)

		assert.Nil(t, ValidatePassword("idontneedanythingspecial"))
		assert.ErrorContains(t, ValidatePassword("abshort"), "Your password may not be less than 9 or be more than 256 characters.")

	})
}
