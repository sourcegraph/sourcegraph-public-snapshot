package types

import (
	"context"
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
			out:  schema.GitHubConnection{Token: RedactedSecret, Url: "https://github.com"},
		},
		{
			kind: extsvc.KindGitLab,
			in:   schema.GitLabConnection{Token: "foobar", Url: "https://gitlab.com", TokenOauthRefresh: "refresh-it"},
			out:  schema.GitLabConnection{Token: RedactedSecret, Url: "https://gitlab.com", TokenOauthRefresh: RedactedSecret},
		},
		{
			kind: extsvc.KindBitbucketServer,
			in: schema.BitbucketServerConnection{
				Password: "foobar",
				Token:    "foobar",
				Url:      "https://bbs.org",
			},
			out: schema.BitbucketServerConnection{
				Password: RedactedSecret,
				Token:    RedactedSecret,
				Url:      "https://bbs.org",
			},
		},
		{
			kind: extsvc.KindBitbucketCloud,
			in:   schema.BitbucketCloudConnection{AppPassword: "foobar", Url: "https://bitbucket.org"},
			out:  schema.BitbucketCloudConnection{AppPassword: RedactedSecret, Url: "https://bitbucket.org"},
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
				SecretAccessKey: RedactedSecret,
				Region:          "us-east-9000z",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "username",
					Password: RedactedSecret,
				},
			},
		},
		{
			kind: extsvc.KindPhabricator,
			in:   schema.PhabricatorConnection{Token: "foobar", Url: "https://phabricator.biz"},
			out:  schema.PhabricatorConnection{Token: RedactedSecret, Url: "https://phabricator.biz"},
		},
		{
			kind: extsvc.KindGitolite,
			in:   schema.GitoliteConnection{Host: "https://gitolite.ninja"},
			out:  schema.GitoliteConnection{Host: "https://gitolite.ninja"},
		},
		{
			kind: extsvc.KindPerforce,
			in:   schema.PerforceConnection{P4User: "foo", P4Passwd: "bar"},
			out:  schema.PerforceConnection{P4User: "foo", P4Passwd: RedactedSecret},
		},
		{
			kind: extsvc.KindPagure,
			in:   schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: "bar"},
			out:  schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: RedactedSecret},
		},
		{
			kind: extsvc.KindJVMPackages,
			in:   schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: "foobar", Dependencies: []string{"baz"}}},
			out:  schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: RedactedSecret, Dependencies: []string{"baz"}}},
		},
		{
			kind: extsvc.KindNpmPackages,
			in:   schema.NpmPackagesConnection{Credentials: "foobar", Registry: "https://registry.npmjs.org"},
			out:  schema.NpmPackagesConnection{Credentials: RedactedSecret, Registry: "https://registry.npmjs.org"},
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
			kind: extsvc.KindGoPackages,
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

			e := ExternalService{Kind: tc.kind, Config: extsvc.NewUnencryptedConfig(string(cfg))}

			ctx := context.Background()
			have, err := e.RedactedConfig(ctx)
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
		kind    string
		old     any
		in      any
		out     any
		wantErr error
	}{
		{
			kind:    extsvc.KindGitHub,
			old:     schema.GitHubConnection{Token: "foobar", Url: "https://github.com"},
			in:      schema.GitHubConnection{Token: RedactedSecret, Url: "https://ghe.sgdev.org"},
			out:     schema.GitHubConnection{Token: "foobar", Url: "https://ghe.sgdev.org"},
			wantErr: errCodeHostIdentityChanged{"url", "token"},
		},
		{
			kind:    extsvc.KindGitLab,
			old:     schema.GitLabConnection{Token: "foobar", Url: "https://gitlab.com", TokenOauthRefresh: "refresh-it"},
			in:      schema.GitLabConnection{Token: RedactedSecret, Url: "https://gitlab.corp.com", TokenOauthRefresh: RedactedSecret},
			out:     schema.GitLabConnection{Token: "foobar", Url: "https://gitlab.corp.com", TokenOauthRefresh: "refresh-it"},
			wantErr: errCodeHostIdentityChanged{"url", "token"},
		},
		{
			kind: extsvc.KindBitbucketServer,
			old: schema.BitbucketServerConnection{
				Password: "foobar",
				Token:    "foobar",
				Url:      "https://bbs.org",
			},
			in: schema.BitbucketServerConnection{
				Password: RedactedSecret,
				Token:    RedactedSecret,
				Url:      "https://bbs.corp.org",
			},
			out: schema.BitbucketServerConnection{
				Password: "foobar",
				Token:    "foobar",
				Url:      "https://bbs.corp.org",
			},
			wantErr: errCodeHostIdentityChanged{"url", "token"},
		},
		{
			kind: extsvc.KindBitbucketServer,
			old: schema.BitbucketServerConnection{
				Token: "foobar",
				Url:   "https://bbs.org",
			},
			in: schema.BitbucketServerConnection{
				Token: RedactedSecret,
				Url:   "https://bbs.corp.org",
			},
			out: schema.BitbucketServerConnection{
				Token: "foobar",
				Url:   "https://bbs.corp.org",
			},
			wantErr: errCodeHostIdentityChanged{"url", "token"},
		},
		{
			kind:    extsvc.KindBitbucketCloud,
			old:     schema.BitbucketCloudConnection{AppPassword: "foobar", Url: "https://bitbucket.org"},
			in:      schema.BitbucketCloudConnection{AppPassword: RedactedSecret, Url: "https://bitbucket.corp.com"},
			out:     schema.BitbucketCloudConnection{AppPassword: "foobar", Url: "https://bitbucket.corp.com"},
			wantErr: errCodeHostIdentityChanged{"apiUrl", "appPassword"},
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
				SecretAccessKey: RedactedSecret,
				Region:          "us-west-9000z",
				GitCredentials: schema.AWSCodeCommitGitCredentials{
					Username: "username",
					Password: RedactedSecret,
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
			kind:    extsvc.KindPhabricator,
			old:     schema.PhabricatorConnection{Token: "foobar", Url: "https://phabricator.biz"},
			in:      schema.PhabricatorConnection{Token: RedactedSecret, Url: "https://phabricator.corp.biz"},
			out:     schema.PhabricatorConnection{Token: "foobar", Url: "https://phabricator.corp.biz"},
			wantErr: errCodeHostIdentityChanged{"url", "token"},
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
			in:   schema.PerforceConnection{P4User: "baz", P4Passwd: RedactedSecret},
			out:  schema.PerforceConnection{P4User: "baz", P4Passwd: "bar"},
		},
		{
			kind:    extsvc.KindPerforce,
			old:     schema.PerforceConnection{P4Port: "tcp://es.ninja", P4User: "foo", P4Passwd: "bar"},
			in:      schema.PerforceConnection{P4Port: "tcp://vr.ninja", P4User: "foo", P4Passwd: RedactedSecret},
			out:     schema.PerforceConnection{P4User: "baz", P4Passwd: "bar"},
			wantErr: errCodeHostIdentityChanged{"p4.port", "p4.passwd"},
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
			in:   schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: RedactedSecret},
			out:  schema.PagureConnection{Url: "https://src.fedoraproject.org", Token: "bar"},
		},
		{
			kind: extsvc.KindJVMPackages,
			old:  schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: "foobar", Dependencies: []string{"baz"}}},
			in:   schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: RedactedSecret, Dependencies: []string{"bar"}}},
			out:  schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: "foobar", Dependencies: []string{"bar"}}},
		},
		{
			kind:    extsvc.KindJVMPackages,
			old:     schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: "foobar", Repositories: []string{"foo", "baz"}}},
			in:      schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: RedactedSecret, Repositories: []string{"bar"}}},
			out:     schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: "foobar", Repositories: []string{"bar"}}},
			wantErr: errCodeHostIdentityChanged{"repositories", "credentials"},
		},
		{
			kind: extsvc.KindJVMPackages,
			old:  schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: "foobar", Repositories: []string{"foo", "baz"}}},
			in:   schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: RedactedSecret, Repositories: []string{"baz", "foo"}}},
			out:  schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: "foobar", Repositories: []string{"baz", "foo"}}},
		},
		{
			kind: extsvc.KindJVMPackages,
			old:  schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: "foobar", Repositories: []string{"foo", "baz"}}},
			in:   schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: RedactedSecret, Repositories: []string{"baz"}}},
			out:  schema.JVMPackagesConnection{Maven: schema.Maven{Credentials: "foobar", Repositories: []string{"baz"}}},
		},
		{
			kind:    extsvc.KindNpmPackages,
			old:     schema.NpmPackagesConnection{Credentials: "foobar", Registry: "https://registry.npmjs.org"},
			in:      schema.NpmPackagesConnection{Credentials: RedactedSecret, Registry: "https://private-registry.npmjs.org"},
			out:     schema.NpmPackagesConnection{Credentials: "foobar", Registry: "https://private-registry.npmjs.org"},
			wantErr: errCodeHostIdentityChanged{"registry", "credentials"},
		},
		{
			kind: extsvc.KindOther,
			old:  schema.OtherExternalServiceConnection{Url: "https://user:pass@other.org"},
			in:   schema.OtherExternalServiceConnection{Url: "https://user:REDACTED@other.org"},
			out:  schema.OtherExternalServiceConnection{Url: "https://user:pass@other.org"},
		},
		{
			kind:    extsvc.KindOther,
			old:     schema.OtherExternalServiceConnection{Url: "https://user:pass@other.org"},
			in:      schema.OtherExternalServiceConnection{Url: "https://user:REDACTED@other.corp.org"},
			out:     schema.OtherExternalServiceConnection{Url: "https://user:pass@other.corp.org"},
			wantErr: errCodeHostIdentityChanged{"url", "password"},
		},
		{
			kind: extsvc.KindGoPackages,
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
			kind: extsvc.KindGoPackages,
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
					"https://user:REDACTED@athens.notgolang.org",
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
			wantErr: errCodeHostIdentityChanged{"url", "password"},
		},
		{
			// Tests that swapping order of URLs doesn't affect correct unredaction.
			kind: extsvc.KindGoPackages,
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
			kind: extsvc.KindGoPackages,
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
			kind: extsvc.KindGoPackages,
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

			old := ExternalService{Kind: tc.kind, Config: extsvc.NewUnencryptedConfig(string(oldCfg))}
			in := ExternalService{Kind: tc.kind, Config: extsvc.NewUnencryptedConfig(string(inCfg))}

			ctx := context.Background()
			err = in.UnredactConfig(ctx, &old)

			if err != nil {
				if tc.wantErr == nil {
					t.Fatal(err)
				} else if tc.wantErr.Error() != err.Error() {
					t.Fatal("received error, but not equals to expected one")
				} else {
					// we expected an error, so we're done here
					return
				}
			}

			if err == nil {
				if tc.wantErr != nil {
					t.Fatal("expected an error, got nil")
				}
			}

			cfg, err := in.Config.Decrypt(ctx)
			if err != nil {
				t.Fatal(err)
			}

			want, err := json.Marshal(tc.out)
			if err != nil {
				t.Fatal(err)
			}

			assert.JSONEq(t, string(want), cfg)
		})
	}
}
