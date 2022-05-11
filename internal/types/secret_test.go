package types

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExternalService_RedactedConfig(t *testing.T) {
	for i, tc := range []struct {
		kind string
		in   any
		out  any
	}{
		{
			kind: extsvc.KindGitHub,
			in:   schema.GitHubConnection{Token: "foobar", Url: "https://github.com"},
			out:  schema.GitHubConnection{Token: "REDACTED", Url: "https://github.com"},
		},
		{
			kind: extsvc.KindGitLab,
			in:   schema.GitLabConnection{Token: "foobar", Url: "https://gitlab.com"},
			out:  schema.GitLabConnection{Token: "REDACTED", Url: "https://gitlab.com"},
		},
		{
			kind: extsvc.KindBitbucketServer,
			in: schema.BitbucketServerConnection{
				Password: "foobar",
				Token:    "foobar",
				Url:      "https://bbs.org",
			},
			out: schema.BitbucketServerConnection{
				Password: "REDACTED",
				Token:    "REDACTED",
				Url:      "https://bbs.org",
			},
		},
		{
			kind: extsvc.KindBitbucketCloud,
			in:   schema.BitbucketCloudConnection{AppPassword: "foobar", Url: "https://bitbucket.com"},
			out:  schema.BitbucketCloudConnection{AppPassword: "REDACTED", Url: "https://bitbucket.com"},
		},
		{
			kind: extsvc.KindAWSCodeCommit,
			in: schema.AWSCodeCommitConnection{
				SecretAccessKey: "foobar",
				Region:          "us-east-9000z",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "username",
					Password: "password",
				},
			},
			out: schema.AWSCodeCommitConnection{
				SecretAccessKey: "REDACTED",
				Region:          "us-east-9000z",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "username",
					Password: "REDACTED",
				},
			},
		},
		{
			kind: extsvc.KindPhabricator,
			in:   schema.PhabricatorConnection{Token: "foobar", Url: "https://phabricator.biz"},
			out:  schema.PhabricatorConnection{Token: "REDACTED", Url: "https://phabricator.biz"},
		},
		{
			kind: extsvc.KindGitolite,
			in:   schema.GitoliteConnection{Host: "https://gitolite.ninja"},
			out:  schema.GitoliteConnection{Host: "https://gitolite.ninja"},
		},
		{
			kind: extsvc.KindPerforce,
			in:   schema.PerforceConnection{P4User: "foo", P4Passwd: "bar"},
			out:  schema.PerforceConnection{P4User: "foo", P4Passwd: "REDACTED"},
		},
		{
			kind: extsvc.KindPagure,
			in:   schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: "bar"},
			out:  schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: "REDACTED"},
		},
		{
			kind: extsvc.KindJVMPackages,
			in:   schema.JVMPackagesConnection{Maven: &schema.Maven{Credentials: "foobar", Dependencies: []string{"baz"}}},
			out:  schema.JVMPackagesConnection{Maven: &schema.Maven{Credentials: "REDACTED", Dependencies: []string{"baz"}}},
		},
		{
			kind: extsvc.KindNpmPackages,
			in:   schema.NpmPackagesConnection{Credentials: "foobar", Registry: "https://registry.npmjs.org"},
			out:  schema.NpmPackagesConnection{Credentials: "REDACTED", Registry: "https://registry.npmjs.org"},
		},
		{
			kind: extsvc.KindOther,
			in:   schema.OtherExternalServiceConnection{Url: "https://other.org"},
			out:  schema.OtherExternalServiceConnection{Url: "https://other.org"},
		},
		{
			kind: extsvc.KindOther,
			in:   schema.OtherExternalServiceConnection{Url: "https://user:pass@other.org"},
			out:  schema.OtherExternalServiceConnection{Url: "https://user:REDACTED@other.org"},
		},
		{
			kind: extsvc.KindGoModules,
			in: schema.GoModulesConnection{
				Dependencies: []string{"github.com/tsenart/vegeta"},
				Urls: []string{
					"https://user:password@athens.golang.org",
					"https://proxy.golang.org",
				},
			},
			out: schema.GoModulesConnection{
				Dependencies: []string{"github.com/tsenart/vegeta"},
				Urls: []string{
					"https://user:REDACTED@athens.golang.org",
					"https://proxy.golang.org",
				},
			},
		},
		{
			kind: extsvc.KindPythonPackages,
			in: schema.PythonPackagesConnection{
				Dependencies: []string{"requests=1.2.3"},
				Urls: []string{
					"https://user:password@pypi.corp/simple",
					"https://pypi.org/simple",
				},
			},
			out: schema.PythonPackagesConnection{
				Dependencies: []string{"requests=1.2.3"},
				Urls: []string{
					"https://user:REDACTED@pypi.corp/simple",
					"https://pypi.org/simple",
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s-%d", tc.kind, i), func(t *testing.T) {
			cfg, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			e := ExternalService{Kind: tc.kind, Config: string(cfg)}

			have, err := e.RedactedConfig()
			if err != nil {
				t.Fatal(err)
			}

			want, err := json.Marshal(tc.out)
			if err != nil {
				t.Fatal(err)
			}

			assert.JSONEq(t, string(want), have)
		})
	}
}

func TestExternalService_UnredactConfig(t *testing.T) {
	for i, tc := range []struct {
		kind string
		old  any
		in   any
		out  any
	}{
		{
			kind: extsvc.KindGitHub,
			old:  schema.GitHubConnection{Token: "foobar", Url: "https://github.com"},
			in:   schema.GitHubConnection{Token: "REDACTED", Url: "https://ghe.sgdev.org"},
			out:  schema.GitHubConnection{Token: "foobar", Url: "https://ghe.sgdev.org"},
		},
		{
			kind: extsvc.KindGitLab,
			old:  schema.GitLabConnection{Token: "foobar", Url: "https://gitlab.com"},
			in:   schema.GitLabConnection{Token: "REDACTED", Url: "https://gitlab.corp.com"},
			out:  schema.GitLabConnection{Token: "foobar", Url: "https://gitlab.corp.com"},
		},
		{
			kind: extsvc.KindBitbucketServer,
			old: schema.BitbucketServerConnection{
				Password: "foobar",
				Token:    "foobar",
				Url:      "https://bbs.org",
			},
			in: schema.BitbucketServerConnection{
				Password: "REDACTED",
				Token:    "REDACTED",
				Url:      "https://bbs.corp.org",
			},
			out: schema.BitbucketServerConnection{
				Password: "foobar",
				Token:    "foobar",
				Url:      "https://bbs.corp.org",
			},
		},
		{
			kind: extsvc.KindBitbucketCloud,
			old:  schema.BitbucketCloudConnection{AppPassword: "foobar", Url: "https://bitbucket.com"},
			in:   schema.BitbucketCloudConnection{AppPassword: "REDACTED", Url: "https://bitbucket.corp.com"},
			out:  schema.BitbucketCloudConnection{AppPassword: "foobar", Url: "https://bitbucket.corp.com"},
		},
		{
			kind: extsvc.KindAWSCodeCommit,
			old: schema.AWSCodeCommitConnection{
				SecretAccessKey: "foobar",
				Region:          "us-east-9000z",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "username",
					Password: "password",
				},
			},
			in: schema.AWSCodeCommitConnection{
				SecretAccessKey: "REDACTED",
				Region:          "us-west-9000z",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "username",
					Password: "REDACTED",
				},
			},
			out: schema.AWSCodeCommitConnection{
				SecretAccessKey: "foobar",
				Region:          "us-west-9000z",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "username",
					Password: "password",
				},
			},
		},
		{
			kind: extsvc.KindPhabricator,
			old:  schema.PhabricatorConnection{Token: "foobar", Url: "https://phabricator.biz"},
			in:   schema.PhabricatorConnection{Token: "REDACTED", Url: "https://phabricator.corp.biz"},
			out:  schema.PhabricatorConnection{Token: "foobar", Url: "https://phabricator.corp.biz"},
		},
		{
			kind: extsvc.KindGitolite,
			old:  schema.GitoliteConnection{Host: "https://gitolite.ninja"},
			in:   schema.GitoliteConnection{Host: "https://gitolite.corp.ninja"},
			out:  schema.GitoliteConnection{Host: "https://gitolite.corp.ninja"},
		},
		{
			kind: extsvc.KindPerforce,
			old:  schema.PerforceConnection{P4User: "foo", P4Passwd: "bar"},
			in:   schema.PerforceConnection{P4User: "baz", P4Passwd: "REDACTED"},
			out:  schema.PerforceConnection{P4User: "baz", P4Passwd: "bar"},
		},
		{
			// Tests that we can remove a secret field and that it won't appear redacted in the output
			kind: extsvc.KindPagure,
			old:  schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: "bar"},
			in:   schema.PagureConnection{Url: "https://src.fedoraproject.org"},
			out:  schema.PagureConnection{Url: "https://src.fedoraproject.org"},
		},
		{
			kind: extsvc.KindPagure,
			old:  schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: "bar"},
			in:   schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: "REDACTED"},
			out:  schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: "bar"},
		},
		{
			kind: extsvc.KindJVMPackages,
			old:  schema.JVMPackagesConnection{Maven: &schema.Maven{Credentials: "foobar", Dependencies: []string{"baz"}}},
			in:   schema.JVMPackagesConnection{Maven: &schema.Maven{Credentials: "REDACTED", Dependencies: []string{"bar"}}},
			out:  schema.JVMPackagesConnection{Maven: &schema.Maven{Credentials: "foobar", Dependencies: []string{"bar"}}},
		},
		{
			kind: extsvc.KindNpmPackages,
			old:  schema.NpmPackagesConnection{Credentials: "foobar", Registry: "https://registry.npmjs.org"},
			in:   schema.NpmPackagesConnection{Credentials: "REDACTED", Registry: "https://private-registry.npmjs.org"},
			out:  schema.NpmPackagesConnection{Credentials: "foobar", Registry: "https://private-registry.npmjs.org"},
		},
		{
			kind: extsvc.KindOther,
			old:  schema.OtherExternalServiceConnection{Url: "https://user:pass@other.org"},
			in:   schema.OtherExternalServiceConnection{Url: "https://user:REDACTED@other.org"},
			out:  schema.OtherExternalServiceConnection{Url: "https://user:pass@other.org"},
		},
		{
			kind: extsvc.KindOther,
			old:  schema.OtherExternalServiceConnection{Url: "https://user:pass@other.org"},
			in:   schema.OtherExternalServiceConnection{Url: "https://user:REDACTED@other.corp.org"},
			out:  schema.OtherExternalServiceConnection{Url: "https://user:pass@other.corp.org"},
		},
		{
			kind: extsvc.KindGoModules,
			old: schema.GoModulesConnection{
				Dependencies: []string{"github.com/tsenart/vegeta"},
				Urls: []string{
					"https://user:password@athens.golang.org",
					"https://proxy.golang.org",
				},
			},
			in: schema.GoModulesConnection{
				Dependencies: []string{"github.com/oklog/ulid"},
				Urls: []string{
					"https://user:REDACTED@athens.golang.org",
					"https://proxy.golang.org",
				},
			},
			out: schema.GoModulesConnection{
				Dependencies: []string{"github.com/oklog/ulid"},
				Urls: []string{
					"https://user:password@athens.golang.org",
					"https://proxy.golang.org",
				},
			},
		},
		{
			kind: extsvc.KindPythonPackages,
			old: schema.PythonPackagesConnection{
				Dependencies: []string{"requests==1.2.3"},
				Urls: []string{
					"https://user:password@artifactory.corp/simple",
					"https://pypi.org/simple",
				},
			},
			in: schema.PythonPackagesConnection{
				Dependencies: []string{"numpy==1.12.4"},
				Urls: []string{
					"https://user:REDACTED@artifactory.corp/simple",
					"https://pypi.org/simple",
				},
			},
			out: schema.PythonPackagesConnection{
				Dependencies: []string{"numpy==1.12.4"},
				Urls: []string{
					"https://user:password@artifactory.corp/simple",
					"https://pypi.org/simple",
				},
			},
		},
		{
			// Tests that swapping order of URLs doesn't affect correct unredaction.
			kind: extsvc.KindGoModules,
			old: schema.GoModulesConnection{
				Urls: []string{
					"https://user:password@athens.golang.org",
					"https://proxy.golang.org",
				},
			},
			in: schema.GoModulesConnection{
				Urls: []string{
					"https://proxy.golang.org",
					"https://user:REDACTED@athens.golang.org",
				},
			},
			out: schema.GoModulesConnection{
				Urls: []string{
					"https://proxy.golang.org",
					"https://user:password@athens.golang.org",
				},
			},
		},
		{
			kind: extsvc.KindGoModules,
			old: schema.GoModulesConnection{
				Urls: []string{
					"https://user:password@athens.golang.org",
				},
			},
			in: schema.GoModulesConnection{
				Urls: []string{
					"https://user:REDACTED@athens.golang.org",
					"https://proxy.golang.org",
				},
			},
			out: schema.GoModulesConnection{
				Urls: []string{
					"https://user:password@athens.golang.org",
					"https://proxy.golang.org",
				},
			},
		},
		{
			kind: extsvc.KindGoModules,
			old: schema.GoModulesConnection{
				Urls: []string{
					"https://user:password@athens.golang.org",
					"https://user:password1@proxy.golang.org",
				},
			},
			in: schema.GoModulesConnection{
				Urls: []string{
					"https://user:REDACTED@proxy.golang.org",
				},
			},
			out: schema.GoModulesConnection{
				Urls: []string{
					"https://user:password1@proxy.golang.org",
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("%s-%d", tc.kind, i), func(t *testing.T) {
			inCfg, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			oldCfg, err := json.Marshal(tc.old)
			if err != nil {
				t.Fatal(err)
			}

			old := ExternalService{Kind: tc.kind, Config: string(oldCfg)}
			in := ExternalService{Kind: tc.kind, Config: string(inCfg)}

			err = in.UnredactConfig(&old)
			if err != nil {
				t.Fatal(err)
			}

			want, err := json.Marshal(tc.out)
			if err != nil {
				t.Fatal(err)
			}

			assert.JSONEq(t, string(want), in.Config)
		})
	}
}
