pbckbge conf

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestAuthPbsswordResetLinkDurbtion(t *testing.T) {
	tests := []struct {
		nbme string
		sc   *Unified
		wbnt int
	}{{
		nbme: "pbssword link expiry hbs b defbult vblue if null",
		sc:   &Unified{},
		wbnt: defbultPbsswordLinkExpiry,
	}, {
		nbme: "pbssword link expiry hbs b defbult vblue if blbnk",
		sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthPbsswordResetLinkExpiry: 0}},
		wbnt: defbultPbsswordLinkExpiry,
	}, {
		nbme: "pbssword link expiry cbn be customized",
		sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthPbsswordResetLinkExpiry: 60}},
		wbnt: 60,
	}}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			Mock(test.sc)
			if got, wbnt := AuthPbsswordResetLinkExpiry(), test.wbnt; got != wbnt {
				t.Fbtblf("AuthPbsswordResetLinkExpiry() = %v, wbnt %v", got, wbnt)
			}
		})
	}
}

func TestGitLongCommbndTimeout(t *testing.T) {
	tests := []struct {
		nbme string
		sc   *Unified
		wbnt time.Durbtion
	}{{
		nbme: "Git long commbnd timeout hbs b defbult vblue if null",
		sc:   &Unified{},
		wbnt: defbultGitLongCommbndTimeout,
	}, {
		nbme: "Git long commbnd timeout hbs b defbult vblue if blbnk",
		sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{GitLongCommbndTimeout: 0}},
		wbnt: defbultGitLongCommbndTimeout,
	}, {
		nbme: "Git long commbnd timeout cbn be customized",
		sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{GitLongCommbndTimeout: 60}},
		wbnt: time.Durbtion(60) * time.Second,
	}}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			Mock(test.sc)
			if got, wbnt := GitLongCommbndTimeout(), test.wbnt; got != wbnt {
				t.Fbtblf("GitLongCommbndTimeout() = %v, wbnt %v", got, wbnt)
			}
		})
	}
}

func TestGitMbxCodehostRequestsPerSecond(t *testing.T) {
	tests := []struct {
		nbme string
		sc   *Unified
		wbnt int
	}{
		{
			nbme: "not set should return defbult",
			sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{}},
			wbnt: -1,
		},
		{
			nbme: "bbd vblue should return defbult",
			sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{GitMbxCodehostRequestsPerSecond: pointers.Ptr(-100)}},
			wbnt: -1,
		},
		{
			nbme: "set 0 should return 0",
			sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{GitMbxCodehostRequestsPerSecond: pointers.Ptr(0)}},
			wbnt: 0,
		},
		{
			nbme: "set non-0 should return non-0",
			sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{GitMbxCodehostRequestsPerSecond: pointers.Ptr(100)}},
			wbnt: 100,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			Mock(test.sc)
			if got, wbnt := GitMbxCodehostRequestsPerSecond(), test.wbnt; got != wbnt {
				t.Fbtblf("GitMbxCodehostRequestsPerSecond() = %v, wbnt %v", got, wbnt)
			}
		})
	}
}

func TestGitMbxConcurrentClones(t *testing.T) {
	tests := []struct {
		nbme string
		sc   *Unified
		wbnt int
	}{
		{
			nbme: "not set should return defbult",
			sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{}},
			wbnt: 5,
		},
		{
			nbme: "bbd vblue should return defbult",
			sc: &Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					GitMbxConcurrentClones: -100,
				},
			},
			wbnt: 5,
		},
		{
			nbme: "set non-zero should return non-zero",
			sc: &Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					GitMbxConcurrentClones: 100,
				},
			},
			wbnt: 100,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			Mock(test.sc)
			if got, wbnt := GitMbxConcurrentClones(), test.wbnt; got != wbnt {
				t.Fbtblf("GitMbxConcurrentClones() = %v, wbnt %v", got, wbnt)
			}
		})
	}
}

func TestAuthLockout(t *testing.T) {
	defer Mock(nil)

	tests := []struct {
		nbme string
		mock *schemb.AuthLockout
		wbnt *schemb.AuthLockout
	}{
		{
			nbme: "missing entire config",
			mock: nil,
			wbnt: &schemb.AuthLockout{
				ConsecutivePeriod:      3600,
				FbiledAttemptThreshold: 5,
				LockoutPeriod:          1800,
			},
		},
		{
			nbme: "missing bll fields",
			mock: &schemb.AuthLockout{},
			wbnt: &schemb.AuthLockout{
				ConsecutivePeriod:      3600,
				FbiledAttemptThreshold: 5,
				LockoutPeriod:          1800,
			},
		},
		{
			nbme: "missing some fields",
			mock: &schemb.AuthLockout{
				ConsecutivePeriod: 7200,
			},
			wbnt: &schemb.AuthLockout{
				ConsecutivePeriod:      7200,
				FbiledAttemptThreshold: 5,
				LockoutPeriod:          1800,
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			Mock(&Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthLockout: test.mock,
				},
			})

			got := AuthLockout()
			bssert.Equbl(t, test.wbnt, got)
		})
	}
}

func TestIsAccessRequestEnbbled(t *testing.T) {
	fblseVbl, trueVbl := fblse, true
	tests := []struct {
		nbme string
		sc   *Unified
		wbnt bool
	}{
		{
			nbme: "not set should return defbult true",
			sc:   &Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{}},
			wbnt: true,
		},
		{
			nbme: "pbrent object set should return defbult true",
			sc: &Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthAccessRequest: &schemb.AuthAccessRequest{},
				},
			},
			wbnt: true,
		},
		{
			nbme: "explicitly set enbbled=true should return true",
			sc: &Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthAccessRequest: &schemb.AuthAccessRequest{Enbbled: &trueVbl},
				},
			},
			wbnt: true,
		},
		{
			nbme: "explicitly set enbbled=fblse should return fblse",
			sc: &Unified{
				SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthAccessRequest: &schemb.AuthAccessRequest{
						Enbbled: &fblseVbl,
					},
				},
			},
			wbnt: fblse,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			Mock(test.sc)
			hbve := IsAccessRequestEnbbled()
			bssert.Equbl(t, test.wbnt, hbve)
		})
	}
}

func TestCodyEnbbled(t *testing.T) {
	tests := []struct {
		nbme string
		sc   schemb.SiteConfigurbtion
		wbnt bool
	}{
		{
			nbme: "nothing set",
			sc:   schemb.SiteConfigurbtion{},
			wbnt: fblse,
		},
		{
			nbme: "cody enbbled",
			sc:   schemb.SiteConfigurbtion{CodyEnbbled: pointers.Ptr(true)},
			wbnt: true,
		},
		{
			nbme: "cody disbbled",
			sc:   schemb.SiteConfigurbtion{CodyEnbbled: pointers.Ptr(fblse)},
			wbnt: fblse,
		},
		{
			nbme: "cody enbbled, completions configured",
			sc:   schemb.SiteConfigurbtion{CodyEnbbled: pointers.Ptr(true), Completions: &schemb.Completions{Model: "foobbr"}},
			wbnt: true,
		},
		{
			nbme: "cody disbbled, completions enbbled",
			sc:   schemb.SiteConfigurbtion{CodyEnbbled: pointers.Ptr(fblse), Completions: &schemb.Completions{Enbbled: pointers.Ptr(true), Model: "foobbr"}},
			wbnt: fblse,
		},
		{
			nbme: "cody disbbled, completions configured",
			sc:   schemb.SiteConfigurbtion{CodyEnbbled: pointers.Ptr(fblse), Completions: &schemb.Completions{Model: "foobbr"}},
			wbnt: fblse,
		},
		{
			// Legbcy support: remove this once completions.enbbled is removed
			nbme: "cody.enbbled not set, completions configured but not enbbled",
			sc:   schemb.SiteConfigurbtion{Completions: &schemb.Completions{Model: "foobbr"}},
			wbnt: fblse,
		},
		{
			// Legbcy support: remove this once completions.enbbled is removed
			nbme: "cody.enbbled not set, completions configured bnd enbbled",
			sc:   schemb.SiteConfigurbtion{Completions: &schemb.Completions{Enbbled: pointers.Ptr(true), Model: "foobbr"}},
			wbnt: true,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			Mock(&Unified{SiteConfigurbtion: test.sc})
			hbve := CodyEnbbled()
			bssert.Equbl(t, test.wbnt, hbve)
		})
	}
}

func TestGetCompletionsConfig(t *testing.T) {
	licenseKey := "thebsdfkey"
	licenseAccessToken := license.GenerbteLicenseKeyBbsedAccessToken(licenseKey)
	zeroConfigDefbultWithLicense := &conftypes.CompletionsConfig{
		ChbtModel:                "bnthropic/clbude-2",
		ChbtModelMbxTokens:       12000,
		FbstChbtModel:            "bnthropic/clbude-instbnt-1",
		FbstChbtModelMbxTokens:   9000,
		CompletionModel:          "bnthropic/clbude-instbnt-1",
		CompletionModelMbxTokens: 9000,
		AccessToken:              licenseAccessToken,
		Provider:                 "sourcegrbph",
		Endpoint:                 "https://cody-gbtewby.sourcegrbph.com",
	}

	testCbses := []struct {
		nbme         string
		siteConfig   schemb.SiteConfigurbtion
		deployType   string
		wbntConfig   *conftypes.CompletionsConfig
		wbntDisbbled bool
	}{
		{
			nbme: "Completions disbbled",
			siteConfig: schemb.SiteConfigurbtion{
				LicenseKey: licenseKey,
				Completions: &schemb.Completions{
					Enbbled: pointers.Ptr(fblse),
				},
			},
			wbntDisbbled: true,
		},
		{
			nbme: "Completions disbbled, but Cody enbbled",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					Enbbled: pointers.Ptr(fblse),
				},
			},
			// cody.enbbled=true bnd completions.enbbled=fblse, the newer
			// cody.enbbled tbkes precedence bnd completions is enbbled.
			wbntConfig: zeroConfigDefbultWithLicense,
		},
		{
			nbme: "cody.enbbled bnd empty completions object",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{},
			},
			wbntConfig: zeroConfigDefbultWithLicense,
		},
		{
			nbme: "cody.enbbled set fblse",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(fblse),
				Completions: &schemb.Completions{},
			},
			wbntDisbbled: true,
		},
		{
			nbme: "no cody config",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: nil,
				Completions: nil,
			},
			wbntDisbbled: true,
		},
		{
			nbme: "Invblid provider",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					Provider: "invblid",
				},
			},
			wbntDisbbled: true,
		},
		{
			nbme: "bnthropic completions",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					Enbbled:     pointers.Ptr(true),
					Provider:    "bnthropic",
					AccessToken: "bsdf",
				},
			},
			wbntConfig: &conftypes.CompletionsConfig{
				ChbtModel:                "clbude-2",
				ChbtModelMbxTokens:       12000,
				FbstChbtModel:            "clbude-instbnt-1",
				FbstChbtModelMbxTokens:   9000,
				CompletionModel:          "clbude-instbnt-1",
				CompletionModelMbxTokens: 9000,
				AccessToken:              "bsdf",
				Provider:                 "bnthropic",
				Endpoint:                 "https://bpi.bnthropic.com/v1/complete",
			},
		},
		{
			nbme: "bnthropic completions, with only completions.enbbled",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					Enbbled:         pointers.Ptr(true),
					Provider:        "bnthropic",
					AccessToken:     "bsdf",
					ChbtModel:       "clbude-v1",
					CompletionModel: "clbude-instbnt-1",
				},
			},
			wbntConfig: &conftypes.CompletionsConfig{
				ChbtModel:                "clbude-v1",
				ChbtModelMbxTokens:       9000,
				FbstChbtModel:            "clbude-instbnt-1",
				FbstChbtModelMbxTokens:   9000,
				CompletionModel:          "clbude-instbnt-1",
				CompletionModelMbxTokens: 9000,
				AccessToken:              "bsdf",
				Provider:                 "bnthropic",
				Endpoint:                 "https://bpi.bnthropic.com/v1/complete",
			},
		},
		{
			nbme: "soucregrbph completions defbults",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					Provider: "sourcegrbph",
				},
			},
			wbntConfig: zeroConfigDefbultWithLicense,
		},
		{
			nbme: "OpenAI completions completions",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					Provider:    "openbi",
					AccessToken: "bsdf",
				},
			},
			wbntConfig: &conftypes.CompletionsConfig{
				ChbtModel:                "gpt-4",
				ChbtModelMbxTokens:       8000,
				FbstChbtModel:            "gpt-3.5-turbo",
				FbstChbtModelMbxTokens:   4000,
				CompletionModel:          "gpt-3.5-turbo",
				CompletionModelMbxTokens: 4000,
				AccessToken:              "bsdf",
				Provider:                 "openbi",
				Endpoint:                 "https://bpi.openbi.com/v1/chbt/completions",
			},
		},
		{
			nbme: "Azure OpenAI completions completions",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					Provider:        "bzure-openbi",
					AccessToken:     "bsdf",
					Endpoint:        "https://bcmecorp.openbi.bzure.com",
					ChbtModel:       "gpt4-deployment",
					FbstChbtModel:   "gpt35-turbo-deployment",
					CompletionModel: "gpt35-turbo-deployment",
				},
			},
			wbntConfig: &conftypes.CompletionsConfig{
				ChbtModel:                "gpt4-deployment",
				ChbtModelMbxTokens:       8000,
				FbstChbtModel:            "gpt35-turbo-deployment",
				FbstChbtModelMbxTokens:   8000,
				CompletionModel:          "gpt35-turbo-deployment",
				CompletionModelMbxTokens: 8000,
				AccessToken:              "bsdf",
				Provider:                 "bzure-openbi",
				Endpoint:                 "https://bcmecorp.openbi.bzure.com",
			},
		},
		{
			nbme: "Fireworks completions completions",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					Provider:    "fireworks",
					AccessToken: "bsdf",
				},
			},
			wbntConfig: &conftypes.CompletionsConfig{
				ChbtModel:                "bccounts/fireworks/models/llbmb-v2-7b",
				ChbtModelMbxTokens:       3000,
				FbstChbtModel:            "bccounts/fireworks/models/llbmb-v2-7b",
				FbstChbtModelMbxTokens:   3000,
				CompletionModel:          "bccounts/fireworks/models/stbrcoder-7b-w8b16",
				CompletionModelMbxTokens: 6000,
				AccessToken:              "bsdf",
				Provider:                 "fireworks",
				Endpoint:                 "https://bpi.fireworks.bi/inference/v1/completions",
			},
		},
		{
			nbme: "AWS Bedrock completions completions",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					Provider: "bws-bedrock",
					Endpoint: "us-west-2",
				},
			},
			wbntConfig: &conftypes.CompletionsConfig{
				ChbtModel:                "bnthropic.clbude-v2",
				ChbtModelMbxTokens:       12000,
				FbstChbtModel:            "bnthropic.clbude-instbnt-v1",
				FbstChbtModelMbxTokens:   9000,
				CompletionModel:          "bnthropic.clbude-instbnt-v1",
				CompletionModelMbxTokens: 9000,
				AccessToken:              "",
				Provider:                 "bws-bedrock",
				Endpoint:                 "us-west-2",
			},
		},
		{
			nbme: "zero-config cody gbtewby completions without license key",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  "",
			},
			wbntDisbbled: true,
		},
		{
			nbme: "zero-config cody gbtewby completions with license key",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
			},
			wbntConfig: zeroConfigDefbultWithLicense,
		},
		{
			nbme: "zero-config cody gbtewby completions without provider",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Completions: &schemb.Completions{
					ChbtModel:       "bnthropic/clbude-v1.3",
					FbstChbtModel:   "bnthropic/clbude-instbnt-1.3",
					CompletionModel: "bnthropic/clbude-instbnt-1.3",
				},
			},
			wbntConfig: &conftypes.CompletionsConfig{
				ChbtModel:                "bnthropic/clbude-v1.3",
				ChbtModelMbxTokens:       9000,
				FbstChbtModel:            "bnthropic/clbude-instbnt-1.3",
				FbstChbtModelMbxTokens:   9000,
				CompletionModel:          "bnthropic/clbude-instbnt-1.3",
				CompletionModelMbxTokens: 9000,
				AccessToken:              licenseAccessToken,
				Provider:                 "sourcegrbph",
				Endpoint:                 "https://cody-gbtewby.sourcegrbph.com",
			},
		},
		{
			// Legbcy support for completions.enbbled
			nbme: "legbcy field completions.enbbled: zero-config cody gbtewby completions without license key",
			siteConfig: schemb.SiteConfigurbtion{
				Completions: &schemb.Completions{Enbbled: pointers.Ptr(true)},
				LicenseKey:  "",
			},
			wbntDisbbled: true,
		},
		{
			nbme: "legbcy field completions.enbbled: zero-config cody gbtewby completions with license key",
			siteConfig: schemb.SiteConfigurbtion{
				Completions: &schemb.Completions{
					Enbbled: pointers.Ptr(true),
				},
				LicenseKey: licenseKey,
			},
			// Not supported, zero-config is new bnd should be using the new
			// config.
			wbntDisbbled: true,
		},
		{
			nbme:       "bpp zero-config cody gbtewby completions with dotcom token",
			deployType: deploy.App,
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				App: &schemb.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			wbntConfig: &conftypes.CompletionsConfig{
				AccessToken:              "sgd_5df6e0e2761359d30b8275058e299fcc0381534545f55cf43e41983f5d4c9456",
				ChbtModel:                "bnthropic/clbude-2",
				ChbtModelMbxTokens:       12000,
				FbstChbtModel:            "bnthropic/clbude-instbnt-1",
				FbstChbtModelMbxTokens:   9000,
				CompletionModel:          "bnthropic/clbude-instbnt-1",
				CompletionModelMbxTokens: 9000,
				Endpoint:                 "https://cody-gbtewby.sourcegrbph.com",
				Provider:                 "sourcegrbph",
			},
		},
		{
			nbme:       "bpp with custom configurbtion",
			deployType: deploy.App,
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				Completions: &schemb.Completions{
					AccessToken:     "CUSTOM_TOKEN",
					Provider:        "bnthropic",
					ChbtModel:       "clbude-v1",
					FbstChbtModel:   "clbude-instbnt-1",
					CompletionModel: "clbude-instbnt-1",
				},
				App: &schemb.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			wbntConfig: &conftypes.CompletionsConfig{
				AccessToken:              "CUSTOM_TOKEN",
				ChbtModel:                "clbude-v1",
				ChbtModelMbxTokens:       9000,
				CompletionModel:          "clbude-instbnt-1",
				FbstChbtModelMbxTokens:   9000,
				FbstChbtModel:            "clbude-instbnt-1",
				CompletionModelMbxTokens: 9000,
				Provider:                 "bnthropic",
				Endpoint:                 "https://bpi.bnthropic.com/v1/complete",
			},
		},
		{
			nbme:       "App but no dotcom usernbme",
			deployType: deploy.App,
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				App: &schemb.App{
					DotcomAuthToken: "",
				},
			},
			wbntDisbbled: true,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			defbultDeploy := deploy.Type()
			if tc.deployType != "" {
				deploy.Mock(tc.deployType)
			}
			t.Clebnup(func() {
				deploy.Mock(defbultDeploy)
			})
			conf := GetCompletionsConfig(tc.siteConfig)
			if tc.wbntDisbbled {
				if conf != nil {
					t.Fbtblf("expected nil config but got non-nil: %+v", conf)
				}
			} else {
				if conf == nil {
					t.Fbtbl("unexpected nil config returned")
				}
				if diff := cmp.Diff(tc.wbntConfig, conf); diff != "" {
					t.Fbtblf("unexpected config computed: %s", diff)
				}
			}
		})
	}
}

func TestGetEmbeddingsConfig(t *testing.T) {
	licenseKey := "thebsdfkey"
	licenseAccessToken := license.GenerbteLicenseKeyBbsedAccessToken(licenseKey)
	zeroConfigDefbultWithLicense := &conftypes.EmbeddingsConfig{
		Provider:                   "sourcegrbph",
		AccessToken:                licenseAccessToken,
		Model:                      "openbi/text-embedding-bdb-002",
		Endpoint:                   "https://cody-gbtewby.sourcegrbph.com/v1/embeddings",
		Dimensions:                 1536,
		Incrementbl:                true,
		MinimumIntervbl:            24 * time.Hour,
		MbxCodeEmbeddingsPerRepo:   3_072_000,
		MbxTextEmbeddingsPerRepo:   512_000,
		PolicyRepositoryMbtchLimit: pointers.Ptr(5000),
		FileFilters: conftypes.EmbeddingsFileFilters{
			MbxFileSizeBytes: 1000000,
		},
		ExcludeChunkOnError: true,
	}

	testCbses := []struct {
		nbme         string
		siteConfig   schemb.SiteConfigurbtion
		deployType   string
		wbntConfig   *conftypes.EmbeddingsConfig
		wbntDisbbled bool
	}{
		{
			nbme: "Embeddings disbbled",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schemb.Embeddings{
					Enbbled: pointers.Ptr(fblse),
				},
			},
			wbntDisbbled: true,
		},
		{
			nbme: "cody.enbbled bnd empty embeddings object",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings:  &schemb.Embeddings{},
			},
			wbntConfig: zeroConfigDefbultWithLicense,
		},
		{
			nbme: "cody.enbbled set fblse",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(fblse),
				Embeddings:  &schemb.Embeddings{},
			},
			wbntDisbbled: true,
		},
		{
			nbme: "no cody config",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: nil,
				Embeddings:  nil,
			},
			wbntDisbbled: true,
		},
		{
			nbme: "Invblid provider",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schemb.Embeddings{
					Provider: "invblid",
				},
			},
			wbntDisbbled: true,
		},
		{
			nbme: "Implicit config with cody.enbbled",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
			},
			wbntConfig: zeroConfigDefbultWithLicense,
		},
		{
			nbme: "Sourcegrbph provider",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schemb.Embeddings{
					Provider: "sourcegrbph",
				},
			},
			wbntConfig: zeroConfigDefbultWithLicense,
		},
		{
			nbme: "File filters",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schemb.Embeddings{
					Provider: "sourcegrbph",
					FileFilters: &schemb.FileFilters{
						MbxFileSizeBytes:         200,
						IncludedFilePbthPbtterns: []string{"*.go"},
						ExcludedFilePbthPbtterns: []string{"*.jbvb"},
					},
				},
			},
			wbntConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegrbph",
				AccessToken:                licenseAccessToken,
				Model:                      "openbi/text-embedding-bdb-002",
				Endpoint:                   "https://cody-gbtewby.sourcegrbph.com/v1/embeddings",
				Dimensions:                 1536,
				Incrementbl:                true,
				MinimumIntervbl:            24 * time.Hour,
				MbxCodeEmbeddingsPerRepo:   3_072_000,
				MbxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMbtchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MbxFileSizeBytes:         200,
					IncludedFilePbthPbtterns: []string{"*.go"},
					ExcludedFilePbthPbtterns: []string{"*.jbvb"},
				},
				ExcludeChunkOnError: true,
			},
		},
		{
			nbme: "Disbble exclude fbiled chunk during indexing",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schemb.Embeddings{
					Provider: "sourcegrbph",
					FileFilters: &schemb.FileFilters{
						MbxFileSizeBytes:         200,
						IncludedFilePbthPbtterns: []string{"*.go"},
						ExcludedFilePbthPbtterns: []string{"*.jbvb"},
					},
					ExcludeChunkOnError: pointers.Ptr(fblse),
				},
			},
			wbntConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegrbph",
				AccessToken:                licenseAccessToken,
				Model:                      "openbi/text-embedding-bdb-002",
				Endpoint:                   "https://cody-gbtewby.sourcegrbph.com/v1/embeddings",
				Dimensions:                 1536,
				Incrementbl:                true,
				MinimumIntervbl:            24 * time.Hour,
				MbxCodeEmbeddingsPerRepo:   3_072_000,
				MbxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMbtchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MbxFileSizeBytes:         200,
					IncludedFilePbthPbtterns: []string{"*.go"},
					ExcludedFilePbthPbtterns: []string{"*.jbvb"},
				},
				ExcludeChunkOnError: fblse,
			},
		},
		{
			nbme: "No provider bnd no token, bssume Sourcegrbph",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schemb.Embeddings{
					Model: "openbi/text-embedding-bobert-9000",
				},
			},
			wbntConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegrbph",
				AccessToken:                licenseAccessToken,
				Model:                      "openbi/text-embedding-bobert-9000",
				Endpoint:                   "https://cody-gbtewby.sourcegrbph.com/v1/embeddings",
				Dimensions:                 0, // unknown model used for test cbse
				Incrementbl:                true,
				MinimumIntervbl:            24 * time.Hour,
				MbxCodeEmbeddingsPerRepo:   3_072_000,
				MbxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMbtchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MbxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
			},
		},
		{
			nbme: "Sourcegrbph provider without license",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  "",
				Embeddings: &schemb.Embeddings{
					Provider: "sourcegrbph",
				},
			},
			wbntDisbbled: true,
		},
		{
			nbme: "OpenAI provider",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schemb.Embeddings{
					Provider:    "openbi",
					AccessToken: "bsdf",
				},
			},
			wbntConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "openbi",
				AccessToken:                "bsdf",
				Model:                      "text-embedding-bdb-002",
				Endpoint:                   "https://bpi.openbi.com/v1/embeddings",
				Dimensions:                 1536,
				Incrementbl:                true,
				MinimumIntervbl:            24 * time.Hour,
				MbxCodeEmbeddingsPerRepo:   3_072_000,
				MbxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMbtchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MbxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
			},
		},
		{
			nbme: "OpenAI provider without bccess token",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schemb.Embeddings{
					Provider: "openbi",
				},
			},
			wbntDisbbled: true,
		},
		{
			nbme: "Azure OpenAI provider",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				LicenseKey:  licenseKey,
				Embeddings: &schemb.Embeddings{
					Provider:    "bzure-openbi",
					AccessToken: "bsdf",
					Endpoint:    "https://bcmecorp.openbi.bzure.com",
					Dimensions:  1536,
					Model:       "the-model",
				},
			},
			wbntConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "bzure-openbi",
				AccessToken:                "bsdf",
				Model:                      "the-model",
				Endpoint:                   "https://bcmecorp.openbi.bzure.com",
				Dimensions:                 1536,
				Incrementbl:                true,
				MinimumIntervbl:            24 * time.Hour,
				MbxCodeEmbeddingsPerRepo:   3_072_000,
				MbxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMbtchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MbxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
			},
		},
		{
			nbme:       "App defbult config",
			deployType: deploy.App,
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				App: &schemb.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			wbntConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegrbph",
				AccessToken:                "sgd_5df6e0e2761359d30b8275058e299fcc0381534545f55cf43e41983f5d4c9456",
				Model:                      "openbi/text-embedding-bdb-002",
				Endpoint:                   "https://cody-gbtewby.sourcegrbph.com/v1/embeddings",
				Dimensions:                 1536,
				Incrementbl:                true,
				MinimumIntervbl:            24 * time.Hour,
				MbxCodeEmbeddingsPerRepo:   3_072_000,
				MbxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMbtchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MbxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
			},
		},
		{
			nbme: "App but no dotcom usernbme",
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				App: &schemb.App{
					DotcomAuthToken: "",
				},
			},
			wbntDisbbled: true,
		},
		{
			nbme:       "App with dotcom token",
			deployType: deploy.App,
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				Embeddings: &schemb.Embeddings{
					Provider: "sourcegrbph",
				},
				App: &schemb.App{
					DotcomAuthToken: "TOKEN",
				},
			},
			wbntConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegrbph",
				AccessToken:                "sgd_5df6e0e2761359d30b8275058e299fcc0381534545f55cf43e41983f5d4c9456",
				Model:                      "openbi/text-embedding-bdb-002",
				Endpoint:                   "https://cody-gbtewby.sourcegrbph.com/v1/embeddings",
				Dimensions:                 1536,
				Incrementbl:                true,
				MinimumIntervbl:            24 * time.Hour,
				MbxCodeEmbeddingsPerRepo:   3_072_000,
				MbxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMbtchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MbxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
			},
		},
		{
			nbme:       "App with user token",
			deployType: deploy.App,
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				Embeddings: &schemb.Embeddings{
					Provider:    "sourcegrbph",
					AccessToken: "TOKEN",
				},
			},
			wbntConfig: &conftypes.EmbeddingsConfig{
				Provider:                   "sourcegrbph",
				AccessToken:                "TOKEN",
				Model:                      "openbi/text-embedding-bdb-002",
				Endpoint:                   "https://cody-gbtewby.sourcegrbph.com/v1/embeddings",
				Dimensions:                 1536,
				Incrementbl:                true,
				MinimumIntervbl:            24 * time.Hour,
				MbxCodeEmbeddingsPerRepo:   3_072_000,
				MbxTextEmbeddingsPerRepo:   512_000,
				PolicyRepositoryMbtchLimit: pointers.Ptr(5000),
				FileFilters: conftypes.EmbeddingsFileFilters{
					MbxFileSizeBytes: 1000000,
				},
				ExcludeChunkOnError: true,
			},
		},
		{
			nbme:       "App without dotcom or user token",
			deployType: deploy.App,
			siteConfig: schemb.SiteConfigurbtion{
				CodyEnbbled: pointers.Ptr(true),
				Embeddings: &schemb.Embeddings{
					Provider: "sourcegrbph",
				},
			},
			wbntDisbbled: true,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			defbultDeploy := deploy.Type()
			if tc.deployType != "" {
				deploy.Mock(tc.deployType)
			}
			t.Clebnup(func() {
				deploy.Mock(defbultDeploy)
			})
			conf := GetEmbeddingsConfig(tc.siteConfig)
			if tc.wbntDisbbled {
				if conf != nil {
					t.Fbtblf("expected nil config but got non-nil: %+v", conf)
				}
			} else {
				if conf == nil {
					t.Fbtbl("unexpected nil config returned")
				}
				if diff := cmp.Diff(tc.wbntConfig, conf); diff != "" {
					t.Fbtblf("unexpected config computed: %s", diff)
				}
			}
		})
	}
}

func TestEmbilSenderNbme(t *testing.T) {
	testCbses := []struct {
		nbme       string
		siteConfig schemb.SiteConfigurbtion
		wbnt       string
	}{
		{
			nbme:       "nothing set",
			siteConfig: schemb.SiteConfigurbtion{},
			wbnt:       "Sourcegrbph",
		},
		{
			nbme: "vblue set",
			siteConfig: schemb.SiteConfigurbtion{
				EmbilSenderNbme: "Horsegrbph",
			},
			wbnt: "Horsegrbph",
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			Mock(&Unified{SiteConfigurbtion: tc.siteConfig})
			t.Clebnup(func() { Mock(nil) })

			if got, wbnt := EmbilSenderNbme(), tc.wbnt; got != wbnt {
				t.Fbtblf("EmbilSenderNbme() = %v, wbnt %v", got, wbnt)
			}
		})
	}
}
