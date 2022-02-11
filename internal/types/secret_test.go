package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

var newValue = "a different value"

func copyStrings(fields []jsonStringField) (out []string) {
	for _, field := range fields {
		out = append(out, *field.ptr)
	}
	return out
}

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
		GitCredentials: schema.AWSCodeCommitGitCredentials{
			Username: "username",
			Password: "password",
		},
	}
	phabricatorConfig := schema.PhabricatorConnection{
		Token: someSecret,
		Url:   "https://phabricator.biz",
	}
	gitoliteConfig := schema.GitoliteConnection{
		Host: "https://gitolite.ninja",
	}
	perforceConfig := schema.PerforceConnection{
		P4Passwd: someSecret,
		P4User:   "admin",
	}
	jvmPackagesConfig := schema.JVMPackagesConnection{
		Maven: &schema.Maven{
			Credentials:  "top secret credentials",
			Dependencies: []string{"placeholder"},
		},
	}
	pagureConfig := schema.PagureConnection{
		Url: "https://src.fedoraproject.org",
	}
	npmPackagesConfig := schema.NPMPackagesConnection{
		Credentials:  "npm credentials!",
		Dependencies: []string{"placeholder"},
	}
	otherConfig := schema.OtherExternalServiceConnection{
		Url:                   someSecret,
		RepositoryPathPattern: "foo",
	}
	var tc = []struct {
		kind      string
		config    interface{}               // the config for the service kind
		editField func(interface{}) *string // a pointer to a field on the config we can edit to simulate the user using the API
	}{
		{
			kind:      extsvc.KindGitHub,
			config:    &githubConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.GitHubConnection).Url },
		},
		{
			kind:      extsvc.KindGitLab,
			config:    &gitlabConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.GitLabConnection).Url },
		},
		{
			kind:      extsvc.KindBitbucketCloud,
			config:    &bitbucketCloudConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.BitbucketCloudConnection).Url },
		},
		// BitbucketServer can have a password OR token, not both
		{
			kind:      extsvc.KindBitbucketServer,
			config:    &bitbucketServerConfigWithPassword,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.BitbucketServerConnection).Url },
		},
		{
			kind:      extsvc.KindBitbucketServer,
			config:    &bitbucketServerConfigWithToken,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.BitbucketServerConnection).Url },
		},
		{
			kind:      extsvc.KindAWSCodeCommit,
			config:    &awsCodeCommitConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.AWSCodeCommitConnection).Region },
		},
		{
			kind:      extsvc.KindPhabricator,
			config:    &phabricatorConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.PhabricatorConnection).Url },
		},
		{
			kind:      extsvc.KindGitolite,
			config:    &gitoliteConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.GitoliteConnection).Host },
		},
		{
			kind:      extsvc.KindPerforce,
			config:    &perforceConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.PerforceConnection).P4User },
		},
		{
			kind:      extsvc.KindJVMPackages,
			config:    &jvmPackagesConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.JVMPackagesConnection).Maven.Dependencies[0] },
		},
		{
			// Unlike the other test cases, this test covers skipping redaction of missing optional fields.
			kind:      extsvc.KindPagure,
			config:    &pagureConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.PagureConnection).Pattern },
		},
		{
			kind:      extsvc.KindNPMPackages,
			config:    &npmPackagesConfig,
			editField: func(cfg interface{}) *string { return &cfg.(*schema.NPMPackagesConnection).Dependencies[0] },
		},
		{
			kind:   extsvc.KindOther,
			config: &otherConfig,
			editField: func(cfg interface{}) *string {
				return &cfg.(*schema.OtherExternalServiceConnection).RepositoryPathPattern
			},
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

			var unredactedFields []string
			// capture the field before redacting
			infos, err := redactionInfo(c.config)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			unredactedFields = copyStrings(infos)

			// first we redact the config as it was received from the DB, then write the redacted form to the user
			svc := ExternalService{
				Kind:   c.kind,
				Config: old,
			}
			redacted, err := svc.RedactConfigSecrets()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			// reset all fields on the config struct to prevent stale data
			if err := zeroFields(c.config); err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if err := json.Unmarshal([]byte(redacted), &c.config); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			infos, err = redactionInfo(c.config)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			for _, got := range copyStrings(infos) {
				if want := RedactedSecret; want != got {
					t.Errorf("want: %q, got: %q", want, got)
				}
			}

			// now we simulate a user updating their config, and writing it back to the API containing redacted secrets
			oldSvc := ExternalService{
				Kind:   c.kind,
				Config: old,
			}
			// edit a field
			*c.editField(c.config) = newValue
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
			if *c.editField(c.config) != newValue {
				t.Errorf("expected %s got %s", newValue, *c.editField(c.config))
			}
			// and the secrets is no longer redacted
			infos, err = redactionInfo(c.config)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if want, got := unredactedFields, copyStrings(infos); !reflect.DeepEqual(want, got) {
				t.Errorf("want: %q, got %q", want, got)
			}
		})
	}
}

func TestUnredactFieldsDeletion(t *testing.T) {
	oldJson := `{"url":"https://src.fedoraproject.org", "token": "oldtoken"}`
	newJson := `{"url":"https://src.fedoraproject.org"}`
	pagureConfig := schema.PagureConnection{}
	_, err := unredactFields(oldJson, newJson, &pagureConfig)
	assert.Nil(t, err)
}
