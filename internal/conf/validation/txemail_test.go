package validation

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/schema"
)

type mockSiteConf schema.SiteConfiguration

func (m mockSiteConf) SiteConfig() schema.SiteConfiguration { return schema.SiteConfiguration(m) }

func TestValidateSiteConfigTemplates(t *testing.T) {
	for _, tt := range []struct {
		name string
		conf mockSiteConf
		want autogold.Value
	}{
		{
			name: "no email.templates",
			conf: mockSiteConf{
				EmailTemplates: nil,
			},
			want: autogold.Expect([]string{}),
		},
		{
			name: "no templates in email.templates",
			conf: mockSiteConf{
				EmailTemplates: &schema.EmailTemplates{},
			},
			want: autogold.Expect([]string{}),
		},
		{
			name: "incomplete template",
			conf: mockSiteConf{
				EmailTemplates: &schema.EmailTemplates{
					SetPassword: &schema.EmailTemplate{
						Subject: "",
						Text:    "",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			want: autogold.Expect([]string{"`email.templates.setPassword` is invalid: fields 'subject' and 'html' are required"}),
		},
		{
			name: "text field is autofilled",
			conf: mockSiteConf{
				EmailTemplates: &schema.EmailTemplates{
					SetPassword: &schema.EmailTemplate{
						Subject: "Set up your Sourcegraph Cloud account for {{.Host}}!",
						Text:    "",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			want: autogold.Expect([]string{}),
		},
		{
			name: "broken template",
			conf: mockSiteConf{
				EmailTemplates: &schema.EmailTemplates{
					SetPassword: &schema.EmailTemplate{
						Subject: "Set up your Sourcegraph Cloud account for {{.Host}}!",
						Text:    "hello world from {{.Hos",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			want: autogold.Expect([]string{"`email.templates.setPassword` is invalid: template: :1: unclosed action"}),
		},
		{
			name: "complete template",
			conf: mockSiteConf{
				EmailTemplates: &schema.EmailTemplates{
					SetPassword: &schema.EmailTemplate{
						Subject: "Set up your Sourcegraph Cloud account for {{.Host}}!",
						Text:    "hello world from {{.Host}}",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			want: autogold.Expect([]string{}),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			problems := validateSiteConfigTemplates(tt.conf)
			tt.want.Equal(t, problems.Messages())
		})
	}
}
