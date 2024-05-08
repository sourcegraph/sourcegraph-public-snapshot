package validation

import (
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

var contributedWarnings []conf.Validator

// GetWarnings identifies problems with the configuration that a site
// admin should address, but do not prevent Sourcegraph from running.
func GetWarnings(db database.DB) (problems conf.Problems, err error) {
	c := *conf.Get()
	for i := range contributedWarnings {
		problems = append(problems, contributedWarnings[i](c)...)
	}
	// Currently this validator requires a DB connection, which we need to pass
	// down here. We don't want to encourage that, so calling this validator out of
	// band.

	problems = append(problems, validateAuthzProviders(c, db)...)

	return problems, nil
}

// contributeWarning adds the configuration validator function to the validation process.
// It is called to validate site configuration. Any problems it returns are shown as configuration
// warnings in the form of site alerts.
//
// It may only be called at init time.
func contributeWarning(f conf.Validator) {
	contributedWarnings = append(contributedWarnings, f)
}
