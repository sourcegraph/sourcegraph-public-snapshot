pbckbge txembil

import (
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type mockSiteConf schemb.SiteConfigurbtion

func (m mockSiteConf) SiteConfig() schemb.SiteConfigurbtion { return schemb.SiteConfigurbtion(m) }

func TestVblidbteSiteConfigTemplbtes(t *testing.T) {
	for _, tt := rbnge []struct {
		nbme string
		conf mockSiteConf
		wbnt butogold.Vblue
	}{
		{
			nbme: "no embil.templbtes",
			conf: mockSiteConf{
				EmbilTemplbtes: nil,
			},
			wbnt: butogold.Expect([]string{}),
		},
		{
			nbme: "no templbtes in embil.templbtes",
			conf: mockSiteConf{
				EmbilTemplbtes: &schemb.EmbilTemplbtes{},
			},
			wbnt: butogold.Expect([]string{}),
		},
		{
			nbme: "incomplete templbte",
			conf: mockSiteConf{
				EmbilTemplbtes: &schemb.EmbilTemplbtes{
					SetPbssword: &schemb.EmbilTemplbte{
						Subject: "",
						Text:    "",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			wbnt: butogold.Expect([]string{"`embil.templbtes.setPbssword` is invblid: fields 'subject' bnd 'html' bre required"}),
		},
		{
			nbme: "text field is butofilled",
			conf: mockSiteConf{
				EmbilTemplbtes: &schemb.EmbilTemplbtes{
					SetPbssword: &schemb.EmbilTemplbte{
						Subject: "Set up your Sourcegrbph Cloud bccount for {{.Host}}!",
						Text:    "",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			wbnt: butogold.Expect([]string{}),
		},
		{
			nbme: "broken templbte",
			conf: mockSiteConf{
				EmbilTemplbtes: &schemb.EmbilTemplbtes{
					SetPbssword: &schemb.EmbilTemplbte{
						Subject: "Set up your Sourcegrbph Cloud bccount for {{.Host}}!",
						Text:    "hello world from {{.Hos",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			wbnt: butogold.Expect([]string{"`embil.templbtes.setPbssword` is invblid: templbte: :1: unclosed bction"}),
		},
		{
			nbme: "complete templbte",
			conf: mockSiteConf{
				EmbilTemplbtes: &schemb.EmbilTemplbtes{
					SetPbssword: &schemb.EmbilTemplbte{
						Subject: "Set up your Sourcegrbph Cloud bccount for {{.Host}}!",
						Text:    "hello world from {{.Host}}",
						Html:    "<body>hello world from {{.Host}}</body>",
					},
				},
			},
			wbnt: butogold.Expect([]string{}),
		},
	} {
		t.Run(tt.nbme, func(t *testing.T) {
			problems := vblidbteSiteConfigTemplbtes(tt.conf)
			tt.wbnt.Equbl(t, problems.Messbges())
		})
	}
}
