package security

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// GenericPasswordPolicy a generic password policy that holds password requirements
type GenericPasswordPolicy struct {
	Enabled                   bool
	MinimumLength             int
	NumberOfSpecialCharacters int
	RequireAtLeastOneNumber   bool
	RequireUpperandLowerCase  bool
}

// passwordPolicyEnabled reports whether the PasswordPolicy feature is enabled (per site config).
func passwordPolicyEnabled() bool {
	pc := getPasswordPolicy()
	return pc.Enabled == true
}

// getPasswordPolicyConfig fetches the possible password policies as defined in the site config
// first it tries to fetch the AuthPasswordPolicy, if not available it tries to fetch the policy
// from the ExperimentalFeatures
func getPasswordPolicyConfig() interface{} {
	pl := conf.Get().AuthPasswordPolicy

	if pl == nil {
		fmt.Println("pl is nil")
		ep := conf.ExperimentalFeatures().PasswordPolicy
		fmt.Println("printing ep")
		fmt.Println(ep)
		return ep
	}

	return pl
}

// getPasswordPolicy converts a AuthPasswordPolicy or a PasswordPolicy into a GenericPasswordPolicy
func getPasswordPolicy() GenericPasswordPolicy {

	p := getPasswordPolicyConfig()
	fmt.Println(p)
	var gp GenericPasswordPolicy

	if p == nil {
		// this means no password policy exists, we return a default Policy that is disabled.
		gp = GenericPasswordPolicy{
			Enabled:                   false,
			MinimumLength:             0,
			NumberOfSpecialCharacters: 0,
			RequireAtLeastOneNumber:   false,
			RequireUpperandLowerCase:  false,
		}

		fmt.Println("returning gp")
		return gp
	}

	fmt.Println(p)
	switch p := p.(type) {

	case *schema.AuthPasswordPolicy:
		gp = GenericPasswordPolicy{
			Enabled:                   p.Enabled,
			MinimumLength:             p.MinimumLength,
			NumberOfSpecialCharacters: p.NumberOfSpecialCharacters,
			RequireAtLeastOneNumber:   p.RequireAtLeastOneNumber,
			RequireUpperandLowerCase:  p.RequireUpperandLowerCase,
		}

	case *schema.PasswordPolicy:
		gp = GenericPasswordPolicy{
			Enabled:                   p.Enabled,
			MinimumLength:             p.MinimumLength,
			NumberOfSpecialCharacters: p.NumberOfSpecialCharacters,
			RequireAtLeastOneNumber:   p.RequireAtLeastOneNumber,
			RequireUpperandLowerCase:  p.RequireUpperandLowerCase,
		}
	}

	return gp
}
