package txemail

import (
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// FromSiteConfigTemplateWithDefault returns a valid txtypes.Templates from customTemplate
// if it is valid, otherwise it returns the given default.
func FromSiteConfigTemplateWithDefault(customTemplate *schema.EmailTemplate, defaultTemplate txtypes.Templates) txtypes.Templates {
	if customTemplate == nil {
		return defaultTemplate
	}

	if custom, err := FromSiteConfigTemplate(*customTemplate); err == nil {
		// If valid, use the custom template. If invalid, proceed with the default
		// template and discard the error - it will also be rendered in site config
		// problems.
		return *custom
	}

	return defaultTemplate
}

// FromSiteConfigTemplate validates and converts an email template configured in site
// configuration to a *txtypes.Templates.
func FromSiteConfigTemplate(input schema.EmailTemplate) (*txtypes.Templates, error) {
	if input.Subject == "" || input.Html == "" {
		return nil, errors.New("fields 'subject' and 'html' are required")
	}
	tpl := txtypes.Templates{
		Subject: input.Subject,
		Text:    input.Text,
		HTML:    input.Html,
	}
	if _, err := ParseTemplate(tpl); err != nil {
		return nil, err
	}
	return &tpl, nil
}
