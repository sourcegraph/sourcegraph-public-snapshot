package types

import (
	"encoding/json"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

var newValue = "a different value"

func TestRoundTripRedactExternalServiceConfig(t *testing.T) {
	someSecret := "this is a secret, i hope no one steals it"
	githubConfig := schema.GitHubConnection{
		Token: someSecret,
		Url:   "https://github.com",
	}
	gitlabConfig := schema.GitLabConnection{
		Token: someSecret,
		Url:   "https://gitlab.com",
	}
	bitbucketCloudConfig := schema.BitbucketCloudConnection{
		AppPassword: someSecret,
		Url:         "https://bitbucket.com",
	}
	bitbucketServerConfigWithPassword := schema.BitbucketServerConnection{
		Password: someSecret,
		Url:      "https://bitbucket.com",
	}
	bitbucketServerConfigWithToken := schema.BitbucketServerConnection{
		Token: someSecret,
		Url:   "https://bitbucket.com",
	}
	awsCodeCommitConfig := schema.AWSCodeCommitConnection{
		SecretAccessKey: someSecret,
		Region:          "us-east-9000z",
	}
	phabricatorConfig := schema.PhabricatorConnection{
		Token: someSecret,
		Url:   "https://phabricator.biz",
	}
	perforceConfig := schema.PerforceConnection{
		P4Passwd: someSecret,
		P4User:   "admin",
	}
	otherConfig := schema.OtherExternalServiceConnection{
		Url:                   someSecret,
		RepositoryPathPattern: "foo",
	}
	var tc = []struct {
		kind        string
		config      interface{} // the config for the service kind
		editField   *string     // a pointer to a field on the config we can edit to simulate the user using the API
		secretField *string     // a pointer to the field we expect to be obfuscated
	}{
		{
			kind:        extsvc.KindGitHub,
			config:      &githubConfig,
			editField:   &githubConfig.Url,
			secretField: &githubConfig.Token,
		},
		{
			kind:        extsvc.KindGitLab,
			config:      &gitlabConfig,
			editField:   &gitlabConfig.Url,
			secretField: &gitlabConfig.Token,
		},
		{
			kind:        extsvc.KindBitbucketCloud,
			config:      &bitbucketCloudConfig,
			editField:   &bitbucketCloudConfig.Url,
			secretField: &bitbucketCloudConfig.AppPassword,
		},
		// BitbucketServer can have a password OR token, not both
		{
			kind:        extsvc.KindBitbucketServer,
			config:      &bitbucketServerConfigWithPassword,
			editField:   &bitbucketServerConfigWithPassword.Url,
			secretField: &bitbucketServerConfigWithPassword.Password,
		},
		{
			kind:        extsvc.KindBitbucketServer,
			config:      &bitbucketServerConfigWithToken,
			editField:   &bitbucketServerConfigWithToken.Url,
			secretField: &bitbucketServerConfigWithToken.Token,
		},
		{
			kind:        extsvc.KindAWSCodeCommit,
			config:      &awsCodeCommitConfig,
			editField:   &awsCodeCommitConfig.Region,
			secretField: &awsCodeCommitConfig.SecretAccessKey,
		},
		{
			kind:        extsvc.KindPhabricator,
			config:      &phabricatorConfig,
			editField:   &phabricatorConfig.Url,
			secretField: &phabricatorConfig.Token,
		},
		{
			kind:        extsvc.KindPerforce,
			config:      &perforceConfig,
			editField:   &perforceConfig.P4User,
			secretField: &perforceConfig.P4Passwd,
		},
		{
			kind:        extsvc.KindOther,
			config:      &otherConfig,
			editField:   &otherConfig.RepositoryPathPattern,
			secretField: &otherConfig.Url,
		},
	}
	for _, c := range tc {
		t.Run(c.kind, func(t *testing.T) {
			// this test simulates the round trip of a user editing external service config via our APIs
			buf, err := json.Marshal(c.config)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			old := string(buf)

			// capture the field before redacting
			unredactedField := *c.secretField

			// first we redact the config as it was received from the DB, then write the redacted form to the user
			svc := ExternalService{
				Kind:   c.kind,
				Config: old,
			}
			if err := svc.RedactConfigSecrets(); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// reset all fields on the config struct to prevent stale data
			if err := zeroFields(c.config); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			if err := json.Unmarshal([]byte(svc.Config), &c.config); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if want, got := RedactedSecret, *c.secretField; want != got {
				t.Errorf("want: %q, got: %q", want, got)
			}

			// now we simulate a user updating their config, and writing it back to the API containing redacted secrets
			oldSvc := ExternalService{
				Kind:   c.kind,
				Config: old,
			}
			// edit a field
			*c.editField = newValue
			buf, err = json.Marshal(c.config)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// this config now contains a redacted token
			newSvc := ExternalService{
				Kind:   c.kind,
				Config: string(buf),
			}
			// unredact fields in newSvc config
			if err := newSvc.UnredactConfig(&oldSvc); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			// reset all fields on the config struct to prevent stale data
			if err := zeroFields(c.config); err != nil {
				t.Errorf("unexpected error: %s", err)
			}

			// the config is now safe to write to the DB, let's unmarshal it again to make sure that no fields are redacted
			// still, and that our updated fields are there
			if err := json.Unmarshal([]byte(newSvc.Config), &c.config); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// our updated fields are still here
			if *c.editField != newValue {
				t.Errorf("expected %s got %s", newValue, *c.editField)
			}
			// and the secret is no longer redacted
			if want, got := unredactedField, *c.secretField; want != got {
				t.Errorf("want: %q, got %q", want, got)
			}
		})
	}

}
