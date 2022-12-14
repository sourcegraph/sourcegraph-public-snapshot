package txemail

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
		{Name: "setPassword", Template: customTemplates.SetPassword},
	} {
		if tpl.Template == nil {
			continue
		}
		if _, err := FromSiteConfigTemplate(*tpl.Template); err != nil {
			message := fmt.Sprintf("`email.templates.%s` is invalid: %s",
				tpl.Name, err.Error())
			problems = append(problems, conf.NewSiteProblem(message))
		}
	}

	return problems
}

// FromSiteConfigTemplate validates and converts an email template configured in site
// configuration to a *txtypes.Templates.
func FromSiteConfigTemplate(input schema.EmailTemplate) (*txtypes.Templates, error) {
	if input.Subject == "" || input.Html == "" || input.Text == "" {
		return nil, errors.New("fields 'subject', 'text', and 'html' are all required")
	}
	if _, err := ParseSiteConfigTemplate(input); err != nil {
		return nil, err
	}
	return &txtypes.Templates{
		Subject: input.Subject,
		Text:    input.Text,
		HTML:    input.Html,
	}, nil
}

// ParseSiteConfigTemplate calls ParseTemplate over an email template configured in site
// configuration.
func ParseSiteConfigTemplate(input schema.EmailTemplate) (*txtypes.ParsedTemplates, error) {
	return ParseTemplate(txtypes.Templates{
		Subject: input.Subject,
		Text:    input.Text,
		HTML:    input.Html,
	})
}
