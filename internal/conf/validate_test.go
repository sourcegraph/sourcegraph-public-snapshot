package conf

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

var executorsAccessToken = "executors-access-token"
var openIDSecret = "open-id-secret"
var gitlabClientSecret = "gitlab-client-secret"
var githubClientSecret = "github-client-secret"

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

func TestRedact(t *testing.T) {
	t.Run("redact secrets", func(t *testing.T) {
		input := getTestSiteWithSecrets(executorsAccessToken, openIDSecret, gitlabClientSecret, githubClientSecret)
		redacted, err := RedactSecrets(conftypes.RawUnified{
			Site: input,
		})
		if err != nil {
			t.Errorf("unexpected error redacting secrets: %s", err)
		}
		expected := getTestSiteWithRedactedSecrets()
		if !redacted.Equal(conftypes.RawUnified{Site: expected}) {
			t.Errorf("unexpected output of RedactSecrets")
		}
	})
}

func TestUnredact(t *testing.T) {
	t.Run("replaces REDACTED with corresponding secret", func(t *testing.T) {
		input := getTestSiteWithRedactedSecrets()
		previousSite := getTestSiteWithSecrets(executorsAccessToken, openIDSecret, githubClientSecret, gitlabClientSecret)
		unredactedSite, err := UnredactSecrets(input, conftypes.RawUnified{Site: previousSite})
		if err != nil {
			t.Errorf("unexpected error unredacting secrets: %s", err)
		}
		if strings.Contains(unredactedSite, RedactedSecret) {
			t.Errorf("expected unredacted to contain secrets")
		}
		expectedUnredacted := conftypes.RawUnified{Site: previousSite}
		unredacted := conftypes.RawUnified{Site: unredactedSite}
		if !unredacted.Equal(expectedUnredacted) {
			t.Errorf("unexpected output of UnredactSecrets")
		}
	})
	t.Run("unredacts secrets AND respects specified edits to secret", func(t *testing.T) {
		input := getTestSiteWithSecrets("new-access-token", RedactedSecret, "new-github-client-secret", RedactedSecret)
		previousSite := getTestSiteWithSecrets(executorsAccessToken, openIDSecret, githubClientSecret, gitlabClientSecret)
		unredactedSite, err := UnredactSecrets(input, conftypes.RawUnified{Site: previousSite})
		if err != nil {
			t.Errorf("unexpected error unredacting secrets: %s", err)
		}
		// Expect to have newly-specified secrets and to fill in "REDACTED" secrets w/ secrets from previous site
		expectedUnredacted := conftypes.RawUnified{Site: getTestSiteWithSecrets("new-access-token", openIDSecret, "new-github-client-secret", gitlabClientSecret)}
		unredacted := conftypes.RawUnified{Site: unredactedSite}
		if !unredacted.Equal(expectedUnredacted) {
			t.Errorf("unexpected output of UnredactSecrets")
		}
	})
	t.Run("unredacts secrets and respects edits to config", func(t *testing.T) {
		newEmail := "new_email@example.com"
		input := getTestSiteWithSecrets("new-access-token", RedactedSecret, "new-github-client-secret", RedactedSecret, newEmail)
		previousSite := getTestSiteWithSecrets(executorsAccessToken, openIDSecret, githubClientSecret, gitlabClientSecret)
		unredactedSite, err := UnredactSecrets(input, conftypes.RawUnified{Site: previousSite})
		if err != nil {
			t.Errorf("unexpected error unredacting secrets: %s", err)
		}
		// Expect new secrets and new email to show up in the unredacted version
		expectedUnredacted := conftypes.RawUnified{Site: getTestSiteWithSecrets("new-access-token", openIDSecret, "new-github-client-secret", gitlabClientSecret, newEmail)}
		unredacted := conftypes.RawUnified{Site: unredactedSite}
		if !unredacted.Equal(expectedUnredacted) {
			t.Errorf("unexpected output of UnredactSecrets")
		}
	})
}

func getTestSiteWithRedactedSecrets() string {
	return getTestSiteWithSecrets(RedactedSecret, RedactedSecret, RedactedSecret, RedactedSecret)
}

func getTestSiteWithSecrets(execAccessToken, openIDSecret, githubSecret, gitlabSecret string, optionalEdit ...string) string {
	email := "noreply+dev@sourcegraph.com"
	if len(optionalEdit) > 0 {
		email = optionalEdit[0]
	}
	return fmt.Sprintf(`
{
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
    "sampling":"selective"
  },
  "externalService.userMode": "all"
}
`, email, execAccessToken, openIDSecret, githubSecret, gitlabSecret)

}
