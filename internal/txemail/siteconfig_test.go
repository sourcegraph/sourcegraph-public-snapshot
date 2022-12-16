package txemail

import (
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/schema"
)

type mockSiteConf schema.SiteConfiguration

func (m mockSiteConf) SiteConfig() schema.SiteConfiguration { return schema.SiteConfiguration(m) }

func TestValidateSiteConfigTemplates(t *testing.T) {
	for _, tt := range []struct {
		conf mockSiteConf
		want autogold.Value
	}{
		{
			conf: mockSiteConf{
				EmailTemplates: nil,
			},
			want: autogold.Want("no email.templates", []string{}),
		},
		{
			conf: mockSiteConf{
				EmailTemplates: &schema.EmailTemplates{},
			},
			want: autogold.Want("no templates in email.templates", []string{}),
		},
		{
			conf: mockSiteConf{
				EmailTemplates: &schema.EmailTemplates{
					SetPassword: &schema.EmailTemplate{
						Subject: "Set up your Sourcegraph Cloud account for {{.Host}}!",
						Text:    "",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			want: autogold.Want("incomplete template", []string{"`email.templates.setPassword` is invalid: fields 'subject', 'text', and 'html' are all required"}),
		},
		{
			conf: mockSiteConf{
				EmailTemplates: &schema.EmailTemplates{
					SetPassword: &schema.EmailTemplate{
						Subject: "Set up your Sourcegraph Cloud account for {{.Host}}!",
						Text:    "hello world from {{.Hos",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			want: autogold.Want("broken template", []string{"`email.templates.setPassword` is invalid: template: :1: unclosed action"}),
		},
		{
			conf: mockSiteConf{
				EmailTemplates: &schema.EmailTemplates{
					SetPassword: &schema.EmailTemplate{
						Subject: "Set up your Sourcegraph Cloud account for {{.Host}}!",
						Text:    "hello world from {{.Host}}",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			want: autogold.Want("complete template", []string{}),
		},
	} {
		t.Run(tt.want.Name(), func(t *testing.T) {
			problems := validateSiteConfigTemplates(tt.conf)
			tt.want.Equal(t, problems.Messages())
		})
	}
}
