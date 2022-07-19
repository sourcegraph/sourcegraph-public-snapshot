package security

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFetchPasswordPolicy(t *testing.T) {
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

func TestFetchPasswordPolicy_nil(t *testing.T) {
	//cfg := conf.Get()

	fmt.Println("start nil test")

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			AuthMinPasswordLength: 9,
		},
	})

	t.Run("policy is parsed correctly", func(t *testing.T) {
		fmt.Println("hello 3")
		p := getPasswordPolicy()
		// policy is enabled
		fmt.Println(p)
		assert.True(t, p.Enabled == true)

		// password length is from AuthPasswordPolicy
		assert.Equal(t, p.MinimumLength, 12)

	})

	// TODO t.Cleanup()
}
