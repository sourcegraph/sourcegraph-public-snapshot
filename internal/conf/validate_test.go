package conf

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	executorsAccessToken                        = "executorsAccessToken"
	authOpenIDClientSecret                      = "authOpenIDClientSecret"
	authGitHubClientSecret                      = "authGitHubClientSecret"
	authGitLabClientSecret                      = "authGitLabClientSecret"
	emailSMTPPassword                           = "emailSMTPPassword"
	organizationInvitationsSigningKey           = "organizationInvitationsSigningKey"
	githubClientSecret                          = "githubClientSecret"
	dotcomGitHubAppCloudClientSecret            = "dotcomGitHubAppCloudClientSecret"
	dotcomGitHubAppCloudPrivateKey              = "dotcomGitHubAppCloudPrivateKey"
	authUnlockAccountLinkSigningKey             = "authUnlockAccountLinkSigningKey"
	dotcomSrcCliVersionCacheGitHubToken         = "dotcomSrcCliVersionCacheGitHubToken"
	dotcomSrcCliVersionCacheGitHubWebhookSecret = "dotcomSrcCliVersionCacheGitHubWebhookSecret"
)

func TestValidate(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		res, err := validate([]byte(schema.SiteSchemaJSON), []byte(`{"maxReposToSearch":123}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Errors()) != 0 {
			t.Errorf("errors: %v", res.Errors())
		}
	})

	t.Run("invalid", func(t *testing.T) {
		res, err := validate([]byte(schema.SiteSchemaJSON), []byte(`{"a":1}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Errors()) == 0 {
			t.Error("want invalid")
		}
	})
}

func TestValidateCustom(t *testing.T) {
	tests := map[string]struct {
		raw         string
		wantProblem string
		wantErr     string
	}{
		"unrecognized auth.providers": {
			raw:     `{"auth.providers":[{"type":"asdf"}]}`,
			wantErr: "tagged union type must have a",
		},
		"valid externalURL": {
			raw: `{"externalURL":"http://example.com"}`,
		},
		"valid externalURL ending with slash": {
			raw: `{"externalURL":"http://example.com/"}`,
		},
		"non-root externalURL": {
			raw:         `{"externalURL":"http://example.com/sourcegraph"}`,
			wantProblem: "externalURL must not be a non-root URL",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			problems, err := validateCustomRaw(conftypes.RawUnified{Site: test.raw})
			if err != nil {
				if test.wantErr == "" {
					t.Fatalf("got unexpected error: %v", err)
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatal(err)
				}
				return
			}
			if test.wantProblem == "" {
				if len(problems) > 0 {
					t.Fatalf("unexpected problems: %v", problems)
				}
				return
			}
			for _, p := range problems {
				if strings.Contains(p.String(), test.wantProblem) {
					return
				}
			}
			t.Fatalf("could not find problem %q in %v", test.wantProblem, problems)
		})
	}
}

func TestProblems(t *testing.T) {
	siteProblems := NewSiteProblems(
		"siteProblem1",
		"siteProblem2",
		"siteProblem3",
	)
	externalServiceProblems := NewExternalServiceProblems(
		"externalServiceProblem1",
		"externalServiceProblem2",
		"externalServiceProblem3",
	)

	var problems Problems
	problems = append(problems, siteProblems...)
	problems = append(problems, externalServiceProblems...)

	{
		messages := make([]string, 0, len(problems))
		messages = append(messages, siteProblems.Messages()...)
		messages = append(messages, externalServiceProblems.Messages()...)

		want := strings.Join(messages, "\n")
		got := strings.Join(problems.Messages(), "\n")
		if want != got {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	{
		want := strings.Join(siteProblems.Messages(), "\n")
		got := strings.Join(problems.Site().Messages(), "\n")
		if want != got {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	{
		want := strings.Join(externalServiceProblems.Messages(), "\n")
		got := strings.Join(problems.ExternalService().Messages(), "\n")
		if want != got {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestRedactSecrets(t *testing.T) {
	redacted, err := RedactSecrets(
		conftypes.RawUnified{
			Site: getTestSiteWithSecrets(
				executorsAccessToken,
				authOpenIDClientSecret, authGitLabClientSecret, authGitHubClientSecret,
				emailSMTPPassword,
				organizationInvitationsSigningKey,
				githubClientSecret,
				dotcomGitHubAppCloudClientSecret,
				dotcomGitHubAppCloudPrivateKey,
				dotcomSrcCliVersionCacheGitHubToken,
				dotcomSrcCliVersionCacheGitHubWebhookSecret,
				authUnlockAccountLinkSigningKey,
			),
		},
	)
	require.NoError(t, err)

	want := getTestSiteWithRedactedSecrets()
	assert.Equal(t, want, redacted.Site)
}

func TestRedactSecrets_AuthProvidersSectionNotAdded(t *testing.T) {
	const cfgWithoutAuthProviders = `{
  "executors.accessToken": "%s"
}`
	redacted, err := RedactSecrets(
		conftypes.RawUnified{
			Site: fmt.Sprintf(cfgWithoutAuthProviders, executorsAccessToken),
		},
	)
	require.NoError(t, err)

	want := fmt.Sprintf(cfgWithoutAuthProviders, "REDACTED")
	assert.Equal(t, want, redacted.Site)
}

func TestUnredactSecrets(t *testing.T) {
	previousSite := getTestSiteWithSecrets(
		executorsAccessToken,
		authOpenIDClientSecret, authGitLabClientSecret, authGitHubClientSecret,
		emailSMTPPassword,
		organizationInvitationsSigningKey,
		githubClientSecret,
		dotcomGitHubAppCloudClientSecret,
		dotcomGitHubAppCloudPrivateKey,
		dotcomSrcCliVersionCacheGitHubToken,
		dotcomSrcCliVersionCacheGitHubWebhookSecret,
		authUnlockAccountLinkSigningKey,
	)

	t.Run("replaces REDACTED with corresponding secret", func(t *testing.T) {
		input := getTestSiteWithRedactedSecrets()
		unredactedSite, err := UnredactSecrets(input, conftypes.RawUnified{Site: previousSite})
		require.NoError(t, err)
		assert.NotContains(t, unredactedSite, redactedSecret)
		assert.Equal(t, previousSite, unredactedSite)
	})

	t.Run("unredacts secrets AND respects specified edits to secret", func(t *testing.T) {
		input := getTestSiteWithSecrets(
			"new"+executorsAccessToken,
			redactedSecret, "new"+authGitLabClientSecret, redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
		)
		unredactedSite, err := UnredactSecrets(input, conftypes.RawUnified{Site: previousSite})
		require.NoError(t, err)

		// Expect to have newly-specified secrets and to fill in "REDACTED" secrets with secrets from previous site
		want := getTestSiteWithSecrets(
			"new"+executorsAccessToken,
			authOpenIDClientSecret, "new"+authGitLabClientSecret, authGitHubClientSecret,
			emailSMTPPassword,
			organizationInvitationsSigningKey,
			githubClientSecret,
			dotcomGitHubAppCloudClientSecret,
			dotcomGitHubAppCloudPrivateKey,
			dotcomSrcCliVersionCacheGitHubToken,
			dotcomSrcCliVersionCacheGitHubWebhookSecret,
			authUnlockAccountLinkSigningKey,
		)
		assert.Equal(t, want, unredactedSite)
	})

	t.Run("unredacts secrets and respects edits to config", func(t *testing.T) {
		const newEmail = "new_email@example.com"
		input := getTestSiteWithSecrets(
			"new"+executorsAccessToken,
			redactedSecret, "new"+authGitLabClientSecret, redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			redactedSecret,
			newEmail,
		)
		unredactedSite, err := UnredactSecrets(input, conftypes.RawUnified{Site: previousSite})
		require.NoError(t, err)

		// Expect new secrets and new email to show up in the unredacted version
		want := getTestSiteWithSecrets(
			"new"+executorsAccessToken,
			authOpenIDClientSecret, "new"+authGitLabClientSecret, authGitHubClientSecret,
			emailSMTPPassword,
			organizationInvitationsSigningKey,
			githubClientSecret,
			dotcomGitHubAppCloudClientSecret,
			dotcomGitHubAppCloudPrivateKey,
			dotcomSrcCliVersionCacheGitHubToken,
			dotcomSrcCliVersionCacheGitHubWebhookSecret,
			authUnlockAccountLinkSigningKey,
			newEmail,
		)
		assert.Equal(t, want, unredactedSite)
	})
}

func getTestSiteWithRedactedSecrets() string {
	return getTestSiteWithSecrets(redactedSecret, redactedSecret, redactedSecret, redactedSecret, redactedSecret, redactedSecret, redactedSecret, redactedSecret, redactedSecret, redactedSecret, redactedSecret, redactedSecret)
}

func getTestSiteWithSecrets(
	executorsAccessToken,
	authOpenIDClientSecret, authGitHubClientSecret, authGitLabClientSecret,
	emailSMTPPassword,
	organizationInvitationsSigningKey,
	githubClientSecret,
	dotcomGitHubAppCloudClientSecret, dotcomGitHubAppCloudPrivateKey,
	dotcomSrcCliVersionCacheGitHubToken, dotcomSrcCliVersionCacheGitHubWebhookSecret,
	authUnlockAccountLinkSigningKey string,
	optionalEdit ...string,
) string {
	email := "noreply+dev@sourcegraph.com"
	if len(optionalEdit) > 0 {
		email = optionalEdit[0]
	}
	return fmt.Sprintf(`{
  "disablePublicRepoRedirects": true,
  "repoListUpdateInterval": 1,
  "email.address": "%s",
  "executors.accessToken": "%s",
  "externalURL": "https://sourcegraph.test:3443",
  "update.channel": "release",
  "auth.providers": [
    {
      "allowSignup": true,
      "type": "builtin"
    },
    {
      "clientID": "sourcegraph-client-openid",
      "clientSecret": "%s",
      "displayName": "Keycloak local OpenID Connect #1 (dev)",
      "issuer": "http://localhost:3220/auth/realms/master",
      "type": "openidconnect"
    },
    {
      "clientID": "sourcegraph-client-github",
      "clientSecret": "%s",
      "displayName": "GitHub.com (dev)",
      "type": "github"
    },
    {
      "clientID": "sourcegraph-client-gitlab",
      "clientSecret": "%s",
      "displayName": "GitLab.com",
      "type": "gitlab",
      "url": "https://gitlab.com"
    }
  ],
  "observability.tracing": {
    "sampling": "selective"
  },
  "externalService.userMode": "all",
  "email.smtp": {
    "password": "%s"
  },
  "organizationInvitations": {
    "signingKey": "%s"
  },
  "githubClientSecret": "%s",
  "dotcom": {
    "githubApp.cloud": {
      "clientSecret": "%s",
      "privateKey": "%s"
    },
    "srcCliVersionCache": {
      "github": {
        "token": "%s",
        "webhookSecret": "%s"
      }
    }
  },
  "auth.unlockAccountLinkSigningKey": "%s",
}`,
		email,
		executorsAccessToken,
		authOpenIDClientSecret, authGitHubClientSecret, authGitLabClientSecret,
		emailSMTPPassword,
		organizationInvitationsSigningKey,
		githubClientSecret,
		dotcomGitHubAppCloudClientSecret, dotcomGitHubAppCloudPrivateKey,
		dotcomSrcCliVersionCacheGitHubToken, dotcomSrcCliVersionCacheGitHubWebhookSecret,
		authUnlockAccountLinkSigningKey,
	)

}
