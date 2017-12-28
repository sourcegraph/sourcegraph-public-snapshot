package conf

import (
	"encoding/json"
	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// ValidateCustom validates the site config using custom validation steps that are not
// able to be expressed in the JSON Schema.
func ValidateCustom(normalizedInput []byte) (validationErrors []string, err error) {
	var cfg schema.SiteConfiguration
	if err := json.Unmarshal(normalizedInput, &cfg); err != nil {
		return nil, err
	}

	invalid := func(msg string) {
		validationErrors = append(validationErrors, msg)
	}

	if cfg.AuthAllowSignup && !(cfg.AuthProvider == "builtin" || cfg.AuthProvider == "auth0") {
		invalid(fmt.Sprintf("auth.allowSignup requires auth.provider == \"builtin\" or \"auth0\" (got %q)", cfg.AuthProvider))
	}

	return validationErrors, nil
}
