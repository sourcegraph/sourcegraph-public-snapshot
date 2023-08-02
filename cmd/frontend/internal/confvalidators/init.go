package confvalidators

import "github.com/sourcegraph/sourcegraph/internal/conf"

// RegisterValidators registers all known validators that will check if the site-config
// is valid.
//
// To add a new validator, create a new function in this package and register it here.
func RegisterValidators() {
	conf.ContributeValidator(validateSiteConfigTemplates)
	conf.ContributeValidator(validateAuthSessionExpiry)
	conf.ContributeValidator(validateLicenseKey)
	conf.ContributeValidator(validateHighlightSettings)
	conf.ContributeValidator(validateKeyringConfig)
	conf.ContributeValidator(validateGitCloneURLMappings)
	conf.ContributeValidator(validateBatchChangeRolloutWindows)
	conf.ContributeValidator(validateAuthProviders)
	conf.ContributeValidator(validateHttpHeaderAuth)
	conf.ContributeValidator(validateOIDCConfig)
	conf.ContributeValidator(validateUserPasswdAuth)
	conf.ContributeValidator(validateSourcegraphOperatorAuth)
	conf.ContributeValidator(validateSAMLAuth)
	conf.ContributeValidator(validateSymbolsConfig)
	conf.ContributeValidator(completionsConfigValidator)
	conf.ContributeValidator(embeddingsConfigValidator)

	conf.ContributeWarning(validateExternalURL)
	conf.ContributeWarning(validateTracerConfig)
	conf.ContributeWarning(validateAuthzProviders)
	conf.ContributeWarning(validatePrometheusConnection)
}
