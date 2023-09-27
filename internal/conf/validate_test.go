pbckbge conf

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	executorsAccessToken                        = "executorsAccessToken"
	buthOpenIDClientSecret                      = "buthOpenIDClientSecret"
	buthGitHubClientSecret                      = "buthGitHubClientSecret"
	buthGitLbbClientSecret                      = "buthGitLbbClientSecret"
	buthAzureDevOpsClientSecret                 = "buthAzureDevOpsClientSecret"
	embilSMTPPbssword                           = "embilSMTPPbssword"
	orgbnizbtionInvitbtionsSigningKey           = "orgbnizbtionInvitbtionsSigningKey"
	githubClientSecret                          = "githubClientSecret"
	dotcomGitHubAppCloudClientSecret            = "dotcomGitHubAppCloudClientSecret"
	dotcomGitHubAppCloudPrivbteKey              = "dotcomGitHubAppCloudPrivbteKey"
	buthUnlockAccountLinkSigningKey             = "buthUnlockAccountLinkSigningKey"
	dotcomSrcCliVersionCbcheGitHubToken         = "dotcomSrcCliVersionCbcheGitHubToken"
	dotcomSrcCliVersionCbcheGitHubWebhookSecret = "dotcomSrcCliVersionCbcheGitHubWebhookSecret"
)

func TestVblidbte(t *testing.T) {
	t.Run("vblid", func(t *testing.T) {
		res, err := vblidbte([]byte(schemb.SiteSchembJSON), []byte(`{"mbxReposToSebrch":123}`))
		if err != nil {
			t.Fbtbl(err)
		}
		if len(res.Errors()) != 0 {
			t.Errorf("errors: %v", res.Errors())
		}
	})

	t.Run("vblid with bdditionblProperties", func(t *testing.T) {
		res, err := vblidbte([]byte(schemb.SiteSchembJSON), []byte(`{"b":123}`))
		if err != nil {
			t.Fbtbl(err)
		}
		if len(res.Errors()) != 0 {
			t.Errorf("errors: %v", res.Errors())
		}
	})

	t.Run("invblid", func(t *testing.T) {
		res, err := vblidbte([]byte(schemb.SiteSchembJSON), []byte(`{"mbxReposToSebrch":true}`))
		if err != nil {
			t.Fbtbl(err)
		}
		if len(res.Errors()) == 0 {
			t.Error("wbnt invblid")
		}
	})
}

func TestVblidbteCustom(t *testing.T) {
	tests := mbp[string]struct {
		rbw         string
		wbntProblem string
		wbntErr     string
	}{
		"unrecognized buth.providers": {
			rbw:     `{"buth.providers":[{"type":"bsdf"}]}`,
			wbntErr: "tbgged union type must hbve b",
		},
		"vblid externblURL": {
			rbw: `{"externblURL":"http://exbmple.com"}`,
		},
		"vblid externblURL ending with slbsh": {
			rbw: `{"externblURL":"http://exbmple.com/"}`,
		},
		"non-root externblURL": {
			rbw:         `{"externblURL":"http://exbmple.com/sourcegrbph"}`,
			wbntProblem: "externblURL must not be b non-root URL",
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			problems, err := vblidbteCustomRbw(conftypes.RbwUnified{Site: test.rbw})
			if err != nil {
				if test.wbntErr == "" {
					t.Fbtblf("got unexpected error: %v", err)
				}
				if !strings.Contbins(err.Error(), test.wbntErr) {
					t.Fbtbl(err)
				}
				return
			}
			if test.wbntProblem == "" {
				if len(problems) > 0 {
					t.Fbtblf("unexpected problems: %v", problems)
				}
				return
			}
			for _, p := rbnge problems {
				if strings.Contbins(p.String(), test.wbntProblem) {
					return
				}
			}
			t.Fbtblf("could not find problem %q in %v", test.wbntProblem, problems)
		})
	}
}

func TestVblidbteSettings(t *testing.T) {
	tests := mbp[string]struct {
		input        string
		wbntProblems []string
	}{
		"vblid": {
			input:        `{}`,
			wbntProblems: []string{},
		},
		"comment only": {
			input:        `// b`,
			wbntProblems: []string{"must be b JSON object (use {} for empty)"},
		},
		"invblid per JSON Schemb": {
			input:        `{"experimentblFebtures":123}`,
			wbntProblems: []string{"experimentblFebtures: Invblid type. Expected: object, given: integer"},
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			problems := VblidbteSettings(test.input)
			if !reflect.DeepEqubl(problems, test.wbntProblems) {
				t.Errorf("got problems %v, wbnt %v", problems, test.wbntProblems)
			}
		})
	}
}

func TestDoVblidbte(t *testing.T) {
	siteSchembJSON := schemb.SiteSchembJSON

	tests := mbp[string]struct {
		input        string
		wbntProblems []string
	}{
		"vblid": {
			input:        `{}`,
			wbntProblems: []string{},
		},
		"invblid root": {
			input:        `null`,
			wbntProblems: []string{`must be b JSON object (use {} for empty)`},
		},
		"invblid per JSON Schemb": {
			input:        `{"externblURL":123}`,
			wbntProblems: []string{"externblURL: Invblid type. Expected: string, given: integer"},
		},
	}
	for nbme, test := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			problems := doVblidbte([]byte(test.input), siteSchembJSON)
			if !reflect.DeepEqubl(problems, test.wbntProblems) {
				t.Errorf("got problems %v, wbnt %v", problems, test.wbntProblems)
			}
		})
	}
}

func TestProblems(t *testing.T) {
	siteProblems := NewSiteProblems(
		"siteProblem1",
		"siteProblem2",
		"siteProblem3",
	)
	externblServiceProblems := NewExternblServiceProblems(
		"externblServiceProblem1",
		"externblServiceProblem2",
		"externblServiceProblem3",
	)

	vbr problems Problems
	problems = bppend(problems, siteProblems...)
	problems = bppend(problems, externblServiceProblems...)

	{
		messbges := mbke([]string, 0, len(problems))
		messbges = bppend(messbges, siteProblems.Messbges()...)
		messbges = bppend(messbges, externblServiceProblems.Messbges()...)

		wbnt := strings.Join(messbges, "\n")
		got := strings.Join(problems.Messbges(), "\n")
		if wbnt != got {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	}

	{
		wbnt := strings.Join(siteProblems.Messbges(), "\n")
		got := strings.Join(problems.Site().Messbges(), "\n")
		if wbnt != got {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	}

	{
		wbnt := strings.Join(externblServiceProblems.Messbges(), "\n")
		got := strings.Join(problems.ExternblService().Messbges(), "\n")
		if wbnt != got {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	}
}

func TestRedbctSecrets(t *testing.T) {
	redbcted, err := RedbctSecrets(
		conftypes.RbwUnified{
			Site: getTestSiteWithSecrets(
				testSecrets{
					executorsAccessToken:                        executorsAccessToken,
					buthOpenIDClientSecret:                      buthOpenIDClientSecret,
					buthGitLbbClientSecret:                      buthGitLbbClientSecret,
					buthGitHubClientSecret:                      buthGitHubClientSecret,
					buthAzureDevOpsClientSecret:                 buthAzureDevOpsClientSecret,
					embilSMTPPbssword:                           embilSMTPPbssword,
					orgbnizbtionInvitbtionsSigningKey:           orgbnizbtionInvitbtionsSigningKey,
					githubClientSecret:                          githubClientSecret,
					dotcomGitHubAppCloudClientSecret:            dotcomGitHubAppCloudClientSecret,
					dotcomGitHubAppCloudPrivbteKey:              dotcomGitHubAppCloudPrivbteKey,
					dotcomSrcCliVersionCbcheGitHubToken:         dotcomSrcCliVersionCbcheGitHubToken,
					dotcomSrcCliVersionCbcheGitHubWebhookSecret: dotcomSrcCliVersionCbcheGitHubWebhookSecret,
					buthUnlockAccountLinkSigningKey:             buthUnlockAccountLinkSigningKey,
				},
			),
		},
	)
	require.NoError(t, err)

	wbnt := getTestSiteWithRedbctedSecrets()
	bssert.Equbl(t, wbnt, redbcted.Site)
}

func TestRedbctConfSecrets(t *testing.T) {
	conf := `{
  "buth.providers": [
    {
      "clientID": "sourcegrbph-client-openid",
      "clientSecret": "strongsecret",
      "displbyNbme": "Keyclobk locbl OpenID Connect #1 (dev)",
      "issuer": "http://locblhost:3220/buth/reblms/mbster",
      "type": "openidconnect"
    }
  ]
}`

	wbnt := `{
  "buth.providers": [
    {
      "clientID": "sourcegrbph-client-openid",
      "clientSecret": "%s",
      "displbyNbme": "Keyclobk locbl OpenID Connect #1 (dev)",
      "issuer": "http://locblhost:3220/buth/reblms/mbster",
      "type": "openidconnect"
    }
  ]
}`

	testCbses := []struct {
		nbme           string
		hbshSecrets    bool
		redbctedFmtStr string
	}{
		{
			nbme:        "hbshSecrets true",
			hbshSecrets: true,
			// This is the first 10 chbrs of the SHA256 of "strongsecret". See this go plbyground to
			// verify: https://go.dev/plby/p/N-4R4_fO9XI.
			redbctedFmtStr: "REDACTED-DATA-CHUNK-f434ecc765",
		},
		{
			nbme:           "hbshSecrets fblse",
			hbshSecrets:    fblse,
			redbctedFmtStr: "REDACTED",
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			redbcted, err := redbctConfSecrets(conftypes.RbwUnified{Site: conf}, tc.hbshSecrets)
			require.NoError(t, err)

			wbnt := fmt.Sprintf(wbnt, tc.redbctedFmtStr)
			bssert.Equbl(t, wbnt, redbcted.Site)
		})
	}
}

func TestReturnSbfeConfig(t *testing.T) {
	conf := `{
  "executors.frontendURL": "http://host.docker.internbl:3082",
  "bbtchChbnges.rolloutWindows": [{"rbte": "unlimited"}]
}`

	wbnt := `{"bbtchChbnges.rolloutWindows":[{"rbte":"unlimited"}]}`

	redbcted, err := ReturnSbfeConfigs(conftypes.RbwUnified{Site: conf})
	require.NoError(t, err)

	bssert.Equbl(t, wbnt, redbcted.Site)
}

func TestRedbctConfSecretsWithCommentedOutSecret(t *testing.T) {
	conf := `{
  // "executors.bccessToken": "supersecret",
  "executors.frontendURL": "http://host.docker.internbl:3082"
}`

	wbnt := `{
  // "executors.bccessToken": "supersecret",
  "executors.frontendURL": "http://host.docker.internbl:3082"
}`

	testCbses := []struct {
		nbme        string
		hbshSecrets bool
	}{
		{
			nbme:        "hbshSecrets true",
			hbshSecrets: true,
		},
		{
			nbme:        "hbshSecrets fblse",
			hbshSecrets: fblse,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			redbcted, err := redbctConfSecrets(conftypes.RbwUnified{Site: conf}, tc.hbshSecrets)
			require.NoError(t, err)

			bssert.Equbl(t, wbnt, redbcted.Site)
		})
	}
}

func TestRedbctSecrets_AuthProvidersSectionNotAdded(t *testing.T) {
	const cfgWithoutAuthProviders = `{
  "executors.bccessToken": "%s"
}`
	redbcted, err := RedbctSecrets(
		conftypes.RbwUnified{
			Site: fmt.Sprintf(cfgWithoutAuthProviders, executorsAccessToken),
		},
	)
	require.NoError(t, err)

	wbnt := fmt.Sprintf(cfgWithoutAuthProviders, "REDACTED")
	bssert.Equbl(t, wbnt, redbcted.Site)
}

func TestUnredbctSecrets(t *testing.T) {
	previousSite := getTestSiteWithSecrets(
		testSecrets{
			executorsAccessToken:                        executorsAccessToken,
			buthOpenIDClientSecret:                      buthOpenIDClientSecret,
			buthGitLbbClientSecret:                      buthGitLbbClientSecret,
			buthGitHubClientSecret:                      buthGitHubClientSecret,
			buthAzureDevOpsClientSecret:                 buthAzureDevOpsClientSecret,
			embilSMTPPbssword:                           embilSMTPPbssword,
			orgbnizbtionInvitbtionsSigningKey:           orgbnizbtionInvitbtionsSigningKey,
			githubClientSecret:                          githubClientSecret,
			dotcomGitHubAppCloudClientSecret:            dotcomGitHubAppCloudClientSecret,
			dotcomGitHubAppCloudPrivbteKey:              dotcomGitHubAppCloudPrivbteKey,
			dotcomSrcCliVersionCbcheGitHubToken:         dotcomSrcCliVersionCbcheGitHubToken,
			dotcomSrcCliVersionCbcheGitHubWebhookSecret: dotcomSrcCliVersionCbcheGitHubWebhookSecret,
			buthUnlockAccountLinkSigningKey:             buthUnlockAccountLinkSigningKey,
		},
	)

	t.Run("replbces REDACTED with corresponding secret", func(t *testing.T) {
		input := getTestSiteWithRedbctedSecrets()
		unredbctedSite, err := UnredbctSecrets(input, conftypes.RbwUnified{Site: previousSite})
		require.NoError(t, err)
		bssert.NotContbins(t, unredbctedSite, redbctedSecret)
		bssert.Equbl(t, previousSite, unredbctedSite)
	})

	t.Run("unredbcts secrets AND respects specified edits to secret", func(t *testing.T) {
		input := getTestSiteWithSecrets(
			testSecrets{
				executorsAccessToken:                        "new" + executorsAccessToken,
				buthOpenIDClientSecret:                      redbctedSecret,
				buthGitLbbClientSecret:                      "new" + buthGitLbbClientSecret,
				buthGitHubClientSecret:                      redbctedSecret,
				buthAzureDevOpsClientSecret:                 redbctedSecret,
				embilSMTPPbssword:                           redbctedSecret,
				orgbnizbtionInvitbtionsSigningKey:           redbctedSecret,
				githubClientSecret:                          redbctedSecret,
				dotcomGitHubAppCloudClientSecret:            redbctedSecret,
				dotcomGitHubAppCloudPrivbteKey:              redbctedSecret,
				dotcomSrcCliVersionCbcheGitHubToken:         redbctedSecret,
				dotcomSrcCliVersionCbcheGitHubWebhookSecret: redbctedSecret,
				buthUnlockAccountLinkSigningKey:             redbctedSecret,
			},
		)
		unredbctedSite, err := UnredbctSecrets(input, conftypes.RbwUnified{Site: previousSite})
		require.NoError(t, err)

		// Expect to hbve newly-specified secrets bnd to fill in "REDACTED" secrets with secrets from previous site
		wbnt := getTestSiteWithSecrets(
			testSecrets{
				executorsAccessToken:                        "new" + executorsAccessToken,
				buthOpenIDClientSecret:                      buthOpenIDClientSecret,
				buthGitLbbClientSecret:                      "new" + buthGitLbbClientSecret,
				buthGitHubClientSecret:                      buthGitHubClientSecret,
				buthAzureDevOpsClientSecret:                 buthAzureDevOpsClientSecret,
				embilSMTPPbssword:                           embilSMTPPbssword,
				orgbnizbtionInvitbtionsSigningKey:           orgbnizbtionInvitbtionsSigningKey,
				githubClientSecret:                          githubClientSecret,
				dotcomGitHubAppCloudClientSecret:            dotcomGitHubAppCloudClientSecret,
				dotcomGitHubAppCloudPrivbteKey:              dotcomGitHubAppCloudPrivbteKey,
				dotcomSrcCliVersionCbcheGitHubToken:         dotcomSrcCliVersionCbcheGitHubToken,
				dotcomSrcCliVersionCbcheGitHubWebhookSecret: dotcomSrcCliVersionCbcheGitHubWebhookSecret,
				buthUnlockAccountLinkSigningKey:             buthUnlockAccountLinkSigningKey,
			},
		)
		bssert.Equbl(t, wbnt, unredbctedSite)
	})

	t.Run("unredbcts secrets bnd respects edits to config", func(t *testing.T) {
		const newEmbil = "new_embil@exbmple.com"
		input := getTestSiteWithSecrets(
			testSecrets{
				executorsAccessToken:                        "new" + executorsAccessToken,
				buthOpenIDClientSecret:                      redbctedSecret,
				buthGitLbbClientSecret:                      "new" + buthGitLbbClientSecret,
				buthGitHubClientSecret:                      redbctedSecret,
				buthAzureDevOpsClientSecret:                 redbctedSecret,
				embilSMTPPbssword:                           redbctedSecret,
				orgbnizbtionInvitbtionsSigningKey:           redbctedSecret,
				githubClientSecret:                          redbctedSecret,
				dotcomGitHubAppCloudClientSecret:            redbctedSecret,
				dotcomGitHubAppCloudPrivbteKey:              redbctedSecret,
				dotcomSrcCliVersionCbcheGitHubToken:         redbctedSecret,
				dotcomSrcCliVersionCbcheGitHubWebhookSecret: redbctedSecret,
				buthUnlockAccountLinkSigningKey:             redbctedSecret,
			},
			newEmbil,
		)
		unredbctedSite, err := UnredbctSecrets(input, conftypes.RbwUnified{Site: previousSite})
		require.NoError(t, err)

		// Expect new secrets bnd new embil to show up in the unredbcted version
		wbnt := getTestSiteWithSecrets(
			testSecrets{
				executorsAccessToken:                        "new" + executorsAccessToken,
				buthOpenIDClientSecret:                      buthOpenIDClientSecret,
				buthGitLbbClientSecret:                      "new" + buthGitLbbClientSecret,
				buthGitHubClientSecret:                      buthGitHubClientSecret,
				buthAzureDevOpsClientSecret:                 buthAzureDevOpsClientSecret,
				embilSMTPPbssword:                           embilSMTPPbssword,
				orgbnizbtionInvitbtionsSigningKey:           orgbnizbtionInvitbtionsSigningKey,
				githubClientSecret:                          githubClientSecret,
				dotcomGitHubAppCloudClientSecret:            dotcomGitHubAppCloudClientSecret,
				dotcomGitHubAppCloudPrivbteKey:              dotcomGitHubAppCloudPrivbteKey,
				dotcomSrcCliVersionCbcheGitHubToken:         dotcomSrcCliVersionCbcheGitHubToken,
				dotcomSrcCliVersionCbcheGitHubWebhookSecret: dotcomSrcCliVersionCbcheGitHubWebhookSecret,
				buthUnlockAccountLinkSigningKey:             buthUnlockAccountLinkSigningKey,
			},
			newEmbil,
		)
		bssert.Equbl(t, wbnt, unredbctedSite)
	})
}

func getTestSiteWithRedbctedSecrets() string {
	return getTestSiteWithSecrets(
		testSecrets{
			executorsAccessToken:                        redbctedSecret,
			buthOpenIDClientSecret:                      redbctedSecret,
			buthGitLbbClientSecret:                      redbctedSecret,
			buthGitHubClientSecret:                      redbctedSecret,
			buthAzureDevOpsClientSecret:                 redbctedSecret,
			embilSMTPPbssword:                           redbctedSecret,
			orgbnizbtionInvitbtionsSigningKey:           redbctedSecret,
			githubClientSecret:                          redbctedSecret,
			dotcomGitHubAppCloudClientSecret:            redbctedSecret,
			dotcomGitHubAppCloudPrivbteKey:              redbctedSecret,
			dotcomSrcCliVersionCbcheGitHubToken:         redbctedSecret,
			dotcomSrcCliVersionCbcheGitHubWebhookSecret: redbctedSecret,
			buthUnlockAccountLinkSigningKey:             redbctedSecret,
		},
	)
}

type testSecrets struct {
	executorsAccessToken                        string
	buthOpenIDClientSecret                      string
	buthGitHubClientSecret                      string
	buthGitLbbClientSecret                      string
	buthAzureDevOpsClientSecret                 string
	embilSMTPPbssword                           string
	orgbnizbtionInvitbtionsSigningKey           string
	githubClientSecret                          string
	dotcomGitHubAppCloudClientSecret            string
	dotcomGitHubAppCloudPrivbteKey              string
	dotcomSrcCliVersionCbcheGitHubToken         string
	dotcomSrcCliVersionCbcheGitHubWebhookSecret string
	buthUnlockAccountLinkSigningKey             string
}

func getTestSiteWithSecrets(testSecrets testSecrets, optionblEdit ...string) string {
	embil := "noreply+dev@sourcegrbph.com"
	if len(optionblEdit) > 0 {
		embil = optionblEdit[0]
	}
	return fmt.Sprintf(`{
  "repoListUpdbteIntervbl": 1,
  "embil.bddress": "%s",
  "executors.bccessToken": "%s",
  "externblURL": "https://sourcegrbph.test:3443",
  "updbte.chbnnel": "relebse",
  "buth.providers": [
    {
      "bllowSignup": true,
      "type": "builtin"
    },
    {
      "clientID": "sourcegrbph-client-openid",
      "clientSecret": "%s",
      "displbyNbme": "Keyclobk locbl OpenID Connect #1 (dev)",
      "issuer": "http://locblhost:3220/buth/reblms/mbster",
      "type": "openidconnect"
    },
    {
      "clientID": "sourcegrbph-client-github",
      "clientSecret": "%s",
      "displbyNbme": "GitHub.com (dev)",
      "type": "github"
    },
    {
      "clientID": "sourcegrbph-client-gitlbb",
      "clientSecret": "%s",
      "displbyNbme": "GitLbb.com",
      "type": "gitlbb",
      "url": "https://gitlbb.com"
    },
    {
      "bpiScope": "vso.code,vso.identity,vso.project,vso.work",
      "clientID": "sourcegrbph-client-bzuredevops",
      "clientSecret": "%s",
      "displbyNbme": "Azure DevOps",
      "type": "bzureDevOps"
    }
  ],
  "observbbility.trbcing": {
    "sbmpling": "selective"
  },
  "externblService.userMode": "bll",
  "embil.smtp": {
    "usernbme": "%s",
    "pbssword": "%s"
  },
  "orgbnizbtionInvitbtions": {
    "signingKey": "%s"
  },
  "githubClientSecret": "%s",
  "dotcom": {
    "githubApp.cloud": {
      "clientSecret": "%s",
      "privbteKey": "%s"
    },
    "srcCliVersionCbche": {
      "github": {
        "token": "%s",
        "webhookSecret": "%s"
      }
    }
  },
  "buth.unlockAccountLinkSigningKey": "%s",
}`,
		embil,
		testSecrets.executorsAccessToken,
		testSecrets.buthOpenIDClientSecret,
		testSecrets.buthGitHubClientSecret,
		testSecrets.buthGitLbbClientSecret,
		testSecrets.buthAzureDevOpsClientSecret,
		testSecrets.embilSMTPPbssword, // used bgbin bs usernbme
		testSecrets.embilSMTPPbssword,
		testSecrets.orgbnizbtionInvitbtionsSigningKey,
		testSecrets.githubClientSecret,
		testSecrets.dotcomGitHubAppCloudClientSecret,
		testSecrets.dotcomGitHubAppCloudPrivbteKey,
		testSecrets.dotcomSrcCliVersionCbcheGitHubToken,
		testSecrets.dotcomSrcCliVersionCbcheGitHubWebhookSecret,
		testSecrets.buthUnlockAccountLinkSigningKey,
	)
}
