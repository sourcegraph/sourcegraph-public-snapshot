package validation

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	conf.ContributeValidator(validateSiteConfigTemplates)
}

// validateSiteConfigTemplates is a conf.Validator that ensures each configured email
// template is valid.
func validateSiteConfigTemplates(confQuerier conftypes.SiteConfigQuerier) (problems conf.Problems) {
	customTemplates := confQuerier.SiteConfig().EmailTemplates
	if customTemplates == nil {
		return nil
	}

	for _, tpl := range []struct {
		Name     string
		Template *schema.EmailTemplate
	}{
		// All templates should go here
		{Name: "resetPassword", Template: customTemplates.ResetPassword},
		{Name: "setPassword", Template: customTemplates.SetPassword},
	} {
		if tpl.Template == nil {
			continue
		}
		if _, err := txemail.FromSiteConfigTemplate(*tpl.Template); err != nil {
			message := fmt.Sprintf("`email.templates.%s` is invalid: %s",
				tpl.Name, err.Error())
			problems = append(problems, conf.NewSiteProblem(message))
		}
	}

	return problems
}
