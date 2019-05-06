package db

import (
	"reflect"
	"sort"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/kylelemons/godebug/pretty"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// This test lives in cmd/enterprise because it tests a proprietary
// super-set of the validation performed by the OSS version.
func TestExternalServices_ValidateConfig(t *testing.T) {
	// Assertion helpers
	equals := func(want ...string) func(testing.TB, []string) {
		sort.Strings(want)
		return func(t testing.TB, have []string) {
			t.Helper()
			sort.Strings(have)
			if !reflect.DeepEqual(have, want) {
				t.Error(pretty.Compare(have, want))
			}
		}
	}

	// Set difference: a - b
	diff := func(a, b []string) (difference []string) {
		set := make(map[string]struct{}, len(b))
		for _, err := range b {
			set[err] = struct{}{}
		}
		for _, err := range a {
			if _, ok := set[err]; !ok {
				difference = append(difference, err)
			}
		}
		return
	}

	includes := func(want ...string) func(testing.TB, []string) {
		return func(t testing.TB, have []string) {
			t.Helper()
			for _, err := range diff(want, have) {
				t.Errorf("%q not found in set:\n%s", err, pretty.Sprint(have))
			}
		}
	}

	excludes := func(want ...string) func(testing.TB, []string) {
		return func(t testing.TB, have []string) {
			t.Helper()
			for _, err := range diff(want, diff(want, have)) {
				t.Errorf("%q found in set:\n%s", err, pretty.Sprint(have))
			}
		}
	}

	// Test table
	for _, tc := range []struct {
		kind   string
		desc   string
		config string
		ps     []schema.AuthProviders
		assert func(testing.TB, []string)
	}{
		{
			kind:   "AWSCODECOMMIT",
			desc:   "without region, accessKeyID, secretAccessKey",
			config: `{}`,
			assert: includes(
				"region: region is required",
				"accessKeyID: accessKeyID is required",
				"secretAccessKey: secretAccessKey is required",
			),
		},
		{
			kind:   "AWSCODECOMMIT",
			desc:   "invalid region",
			config: `{"region": "foo", "accessKeyID": "bar", "secretAccessKey": "baz"}`,
			assert: includes(
				`region: region must be one of the following: "ap-northeast-1", "ap-northeast-2", "ap-south-1", "ap-southeast-1", "ap-southeast-2", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "eu-west-3", "sa-east-1", "us-east-1", "us-east-2", "us-west-1", "us-west-2"`,
			),
		},
		{
			kind:   "AWSCODECOMMIT",
			desc:   "valid",
			config: `{"region": "eu-west-2", "accessKeyID": "bar", "secretAccessKey": "baz"}`,
			assert: equals("<nil>"),
		},
		{
			kind: "AWSCODECOMMIT",
			desc: "valid exclude",
			config: `
			{
				"region": "eu-west-1",
				"accessKeyID": "bar",
				"secretAccessKey": "baz",
				"exclude": [
					{
					  "name": "git-codecommit.eu-central-3.amazonaws.com/foobar-barfoo_bazbar",
					  "id": "arn:aws:codecommit:eu-central-3:999999999999:foobar-barfoo_bazbar"
					},
					{
					  "name": "git-codecommit-fips.us-west-1.amazonaws.com/foobar-barfoo_bazbar",
					  "id": "arn:aws:codecommit:us-west-1:999999999999:foobar-barfoo_bazbar"
					},
				]
			}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "AWSCODECOMMIT",
			desc:   "invalid empty exclude",
			config: `{"exclude": []}`,
			assert: includes(`exclude: Array must have at least 1 items`),
		},
		{
			kind:   "AWSCODECOMMIT",
			desc:   "invalid empty exclude item",
			config: `{"exclude": [{}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "AWSCODECOMMIT",
			desc:   "invalid exclude item",
			config: `{"exclude": [{"foo": "bar"}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "AWSCODECOMMIT",
			desc:   "invalid exclude item name",
			config: `{"exclude": [{"name": "bar"}]}`,
			assert: includes(`exclude.0.name: Does not match pattern '^(git-codecommit-fips|git-codecommit)\.[\w.\d-]+\.amazonaws\.com/[\w.-]+$'`),
		},
		{
			kind:   "AWSCODECOMMIT",
			desc:   "invalid exclude item id",
			config: `{"exclude": [{"id": "bar"}]}`,
			assert: includes(`exclude.0.id: Does not match pattern '^arn:aws:codecommit:[\w\d-]+:\d+:[\w.-]+$'`),
		},
		{
			kind: "AWSCODECOMMIT",
			desc: "invalid additional exclude item properties",
			config: `{"exclude": [{
				"id": "arn:aws:codecommit:us-west-1:999999999999:foobar",
				"bar": "baz"
			}]}`,
			assert: includes(`bar: Additional property bar is not allowed`),
		},
		{
			kind: "AWSCODECOMMIT",
			desc: "both name and id can be specified in exclude",
			config: `
			{
				"region": "eu-west-1",
				"accessKeyID": "bar",
				"secretAccessKey": "baz",
				"exclude": [
					{
					  "name": "git-codecommit.eu-central-3.amazonaws.com/foobar",
					  "id": "arn:aws:codecommit:eu-central-3:999999999999:foobar"
					},
					{
					  "name": "git-codecommit-fips.us-west-1.amazonaws.com/foobar",
					  "id": "arn:aws:codecommit:us-west-1:999999999999:foobar"
					},
				]
			}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "GITOLITE",
			desc:   "witout prefix nor host",
			config: `{}`,
			assert: includes(
				"prefix: prefix is required",
				"host: host is required",
			),
		},
		{
			kind:   "GITOLITE",
			desc:   "with example.com defaults",
			config: `{"prefix": "gitolite.example.com/", "host": "git@gitolite.example.com"}`,
			assert: includes(
				"prefix: Must not validate the schema (not)",
				"host: Must not validate the schema (not)",
			),
		},
		{
			kind:   "GITOLITE",
			desc:   "witout prefix nor host",
			config: `{}`,
			assert: includes(
				"prefix: prefix is required",
				"host: host is required",
			),
		},
		{
			kind:   "GITOLITE",
			desc:   "bad blacklist regex",
			config: `{"blacklist": "]["}`,
			assert: includes("blacklist: Does not match format 'regex'"),
		},
		{
			kind:   "GITOLITE",
			desc:   "phabricator without url nor callsignCommand",
			config: `{"phabricator": {}}`,
			assert: includes(
				"url: url is required",
				"callsignCommand: callsignCommand is required",
			),
		},
		{
			kind:   "GITOLITE",
			desc:   "phabricator with invalid url",
			config: `{"phabricator": {"url": "not-a-url"}}`,
			assert: includes("phabricator.url: Does not match format 'uri'"),
		},
		{
			kind:   "GITOLITE",
			desc:   "invalid empty exclude",
			config: `{"exclude": []}`,
			assert: includes(`exclude: Array must have at least 1 items`),
		},
		{
			kind:   "GITOLITE",
			desc:   "invalid empty exclude item",
			config: `{"exclude": [{}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "GITOLITE",
			desc:   "invalid exclude item",
			config: `{"exclude": [{"foo": "bar"}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "GITOLITE",
			desc:   "invalid exclude item name",
			config: `{"exclude": [{"name": ""}]}`,
			assert: includes(`exclude.0.name: String length must be greater than or equal to 1`),
		},
		{
			kind:   "GITOLITE",
			desc:   "invalid additional exclude item properties",
			config: `{"exclude": [{"name": "foo", "bar": "baz"}]}`,
			assert: includes(`bar: Additional property bar is not allowed`),
		},
		{
			kind: "GITOLITE",
			desc: "name can be specified in exclude",
			config: `
			{
				"prefix": "/",
				"host": "gitolite.mycorp.com",
				"exclude": [
					{"name": "bar"},
				]
			}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "without url, username nor repositoryQuery",
			config: `{}`,
			assert: includes(
				"url: url is required",
				"username: username is required",
				"repositoryQuery: repositoryQuery is required",
			),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "without username",
			config: `{}`,
			assert: includes("username: username is required"),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "example url",
			config: `{"url": "https://bitbucket.example.com"}`,
			assert: includes("url: Must not validate the schema (not)"),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "bad url scheme",
			config: `{"url": "badscheme://bitbucket.com"}`,
			assert: includes("url: Does not match pattern '^https?://'"),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "with token AND password",
			config: `{"token": "foo", "password": "bar"}`,
			assert: includes(
				"(root): Must validate one and only one schema (oneOf)",
				"password: Invalid type. Expected: null, given: string",
			),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid token",
			config: `{"token": ""}`,
			assert: includes(`token: String length must be greater than or equal to 1`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid git url type",
			config: `{"gitURLType": "bad"}`,
			assert: includes(`gitURLType: gitURLType must be one of the following: "http", "ssh"`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid certificate",
			config: `{"certificate": ""}`,
			assert: includes("certificate: Does not match pattern '^-----BEGIN CERTIFICATE-----\n'"),
		},
		{
			kind: "BITBUCKETSERVER",
			desc: "valid",
			config: `
			{
				"url": "https://bitbucket.com/",
				"username": "admin",
				"token": "secret-token",
				"repositoryQuery": ["none"]
			}`,
			assert: equals("<nil>"),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "empty repositoryQuery",
			config: `{"repositoryQuery": []}`,
			assert: includes(`repositoryQuery: Array must have at least 1 items`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "empty repositoryQuery item",
			config: `{"repositoryQuery": [""]}`,
			assert: includes(`repositoryQuery.0: String length must be greater than or equal to 1`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid empty exclude",
			config: `{"exclude": []}`,
			assert: includes(`exclude: Array must have at least 1 items`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid empty exclude item",
			config: `{"exclude": [{}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid exclude item",
			config: `{"exclude": [{"foo": "bar"}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid exclude item name",
			config: `{"exclude": [{"name": "bar"}]}`,
			assert: includes(`exclude.0.name: Does not match pattern '^[\w-]+/[\w.-]+$'`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid additional exclude item properties",
			config: `{"exclude": [{"id": 1234, "bar": "baz"}]}`,
			assert: includes(`bar: Additional property bar is not allowed`),
		},
		{
			kind: "BITBUCKETSERVER",
			desc: "both name and id can be specified in exclude",
			config: `
			{
				"url": "https://bitbucketserver.corp.com",
				"username": "admin",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"exclude": [
					{"name": "foo/bar", "id": 1234},
					{"pattern": "^private/.*"}
				]
			}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid empty repos",
			config: `{"repos": []}`,
			assert: includes(`repos: Array must have at least 1 items`),
		},
		{
			kind:   "BITBUCKETSERVER",
			desc:   "invalid empty repos item",
			config: `{"repos": [""]}`,
			assert: includes(`repos.0: Does not match pattern '^[\w-]+/[\w.-]+$'`),
		},
		{
			kind: "BITBUCKETSERVER",
			desc: "invalid exclude pattern",
			config: `
			{
				"url": "https://bitbucketserver.corp.com",
				"username": "admin",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"exclude": [
					{"pattern": "["}
				]
			}`,
			assert: includes(`exclude.0.pattern: Does not match format 'regex'`),
		},
		{
			kind: "BITBUCKETSERVER",
			desc: "valid repos",
			config: `
			{
				"url": "https://bitbucketserver.corp.com",
				"username": "admin",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"repos": [
					"foo/bar",
					"bar/baz"
				]
			}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "GITHUB",
			desc:   "without url, token nor repositoryQuery",
			config: `{}`,
			assert: includes(
				"url: url is required",
				"token: token is required",
				"repositoryQuery: repositoryQuery is required",
			),
		},
		{
			kind:   "GITHUB",
			desc:   "with example.com url and badscheme",
			config: `{"url": "badscheme://github-enterprise.example.com"}`,
			assert: includes(
				"url: Must not validate the schema (not)",
				"url: Does not match pattern '^https?://'",
			),
		},
		{
			kind:   "GITHUB",
			desc:   "with invalid gitURLType",
			config: `{"gitURLType": "git"}`,
			assert: includes(`gitURLType: gitURLType must be one of the following: "http", "ssh"`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid token",
			config: `{"token": ""}`,
			assert: includes(`token: String length must be greater than or equal to 1`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid certificate",
			config: `{"certificate": ""}`,
			assert: includes("certificate: Does not match pattern '^-----BEGIN CERTIFICATE-----\n'"),
		},
		{
			kind:   "GITHUB",
			desc:   "empty repositoryQuery",
			config: `{"repositoryQuery": []}`,
			assert: includes(`repositoryQuery: Array must have at least 1 items`),
		},
		{
			kind:   "GITHUB",
			desc:   "empty repositoryQuery item",
			config: `{"repositoryQuery": [""]}`,
			assert: includes(`repositoryQuery.0: String length must be greater than or equal to 1`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid repos",
			config: `{"repos": [""]}`,
			assert: includes(`repos.0: Does not match pattern '^[\w-]+/[\w.-]+$'`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid authorization ttl",
			config: `{"authorization": {"ttl": "foo"}}`,
			assert: includes(`authorization.ttl: time: invalid duration foo`),
		},
		{
			kind:   "GITHUB",
			desc:   "valid authorization ttl 0",
			config: `{"authorization": {"ttl": "0"}}`,
			assert: excludes(`authorization.ttl: time: invalid duration 0`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid empty exclude",
			config: `{"exclude": []}`,
			assert: includes(`exclude: Array must have at least 1 items`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid empty exclude item",
			config: `{"exclude": [{}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid exclude item",
			config: `{"exclude": [{"foo": "bar"}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid exclude item name",
			config: `{"exclude": [{"name": "bar"}]}`,
			assert: includes(`exclude.0.name: Does not match pattern '^[\w-]+/[\w.-]+$'`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid empty exclude item id",
			config: `{"exclude": [{"id": ""}]}`,
			assert: includes(`exclude.0.id: String length must be greater than or equal to 1`),
		},
		{
			kind:   "GITHUB",
			desc:   "invalid additional exclude item properties",
			config: `{"exclude": [{"id": "foo", "bar": "baz"}]}`,
			assert: includes(`bar: Additional property bar is not allowed`),
		},
		{
			kind: "GITHUB",
			desc: "both name and id can be specified in exclude",
			config: `
			{
				"url": "https://github.corp.com",
				"token": "very-secret-token",
				"repositoryQuery": ["none"],
				"exclude": [
					{"name": "foo/bar", "id": "AAAAA="}
				]
			}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "GITLAB",
			desc:   "empty projectQuery",
			config: `{"projectQuery": []}`,
			assert: includes(`projectQuery: Array must have at least 1 items`),
		},
		{
			kind:   "GITLAB",
			desc:   "empty projectQuery item",
			config: `{"projectQuery": [""]}`,
			assert: includes(`projectQuery.0: String length must be greater than or equal to 1`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid empty exclude",
			config: `{"exclude": []}`,
			assert: includes(`exclude: Array must have at least 1 items`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid empty exclude item",
			config: `{"exclude": [{}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid exclude item",
			config: `{"exclude": [{"foo": "bar"}]}`,
			assert: includes(`exclude.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid exclude item name",
			config: `{"exclude": [{"name": "bar"}]}`,
			assert: includes(`exclude.0.name: Does not match pattern '^[\w-]+/[\w.-]+$'`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid additional exclude item properties",
			config: `{"exclude": [{"id": 1234, "bar": "baz"}]}`,
			assert: includes(`bar: Additional property bar is not allowed`),
		},
		{
			kind: "GITLAB",
			desc: "both name and id can be specified in exclude",
			config: `
			{
				"url": "https://gitlab.corp.com",
				"token": "very-secret-token",
				"projectQuery": ["none"],
				"exclude": [
					{"name": "foo/bar", "id": 1234}
				]
			}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid empty projects",
			config: `{"projects": []}`,
			assert: includes(`projects: Array must have at least 1 items`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid empty projects item",
			config: `{"projects": [{}]}`,
			assert: includes(`projects.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid projects item",
			config: `{"projects": [{"foo": "bar"}]}`,
			assert: includes(`projects.0: Must validate at least one schema (anyOf)`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid projects item name",
			config: `{"projects": [{"name": "bar"}]}`,
			assert: includes(`projects.0.name: Does not match pattern '^[\w-]+/[\w.-]+$'`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid additional projects item properties",
			config: `{"projects": [{"id": 1234, "bar": "baz"}]}`,
			assert: includes(`bar: Additional property bar is not allowed`),
		},
		{
			kind: "GITLAB",
			desc: "both name and id can be specified in projects",
			config: `
			{
				"url": "https://gitlab.corp.com",
				"token": "very-secret-token",
				"projectQuery": ["none"],
				"projects": [
					{"name": "foo/bar", "id": 1234}
				]
			}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "GITLAB",
			desc:   "without url, token nor projectQuery",
			config: `{}`,
			assert: includes(
				"url: url is required",
				"token: token is required",
				"projectQuery: projectQuery is required",
			),
		},
		{
			kind:   "GITLAB",
			desc:   "with example.com url and badscheme",
			config: `{"url": "badscheme://github-enterprise.example.com"}`,
			assert: includes(
				"url: Must not validate the schema (not)",
				"url: Does not match pattern '^https?://'",
			),
		},
		{
			kind:   "GITLAB",
			desc:   "with invalid gitURLType",
			config: `{"gitURLType": "git"}`,
			assert: includes(`gitURLType: gitURLType must be one of the following: "http", "ssh"`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid token",
			config: `{"token": ""}`,
			assert: includes(`token: String length must be greater than or equal to 1`),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid certificate",
			config: `{"certificate": ""}`,
			assert: includes("certificate: Does not match pattern '^-----BEGIN CERTIFICATE-----\n'"),
		},
		{
			kind:   "GITLAB",
			desc:   "invalid authorization ttl",
			config: `{"authorization": {"ttl": "foo"}}`,
			assert: includes(`authorization.ttl: time: invalid duration foo`),
		},
		{
			kind:   "GITLAB",
			desc:   "valid authorization ttl 0",
			config: `{"authorization": {"ttl": "0"}}`,
			assert: excludes(`authorization.ttl: time: invalid duration 0`),
		},
		{
			kind: "GITLAB",
			desc: "missing oauth provider",
			config: `
			{
				"url": "https://gitlab.foo.bar",
				"authorization": { "identityProvider": { "type": "oauth" } }
			}
			`,
			assert: includes(`Did not find authentication provider matching "https://gitlab.foo.bar"`),
		},
		{
			kind: "GITLAB",
			desc: "valid oauth provider",
			config: `
			{
				"url": "https://gitlab.foo.bar",
				"authorization": { "identityProvider": { "type": "oauth" } }
			}
			`,
			ps: []schema.AuthProviders{
				{Gitlab: &schema.GitLabAuthProvider{Url: "https://gitlab.foo.bar"}},
			},
			assert: excludes(`Did not find authentication provider matching "https://gitlab.foo.bar"`),
		},
		{
			kind: "GITLAB",
			desc: "missing external provider",
			config: `
			{
				"url": "https://gitlab.foo.bar",
				"authorization": {
					"identityProvider": {
						"type": "external",
						"authProviderID": "foo",
						"authProviderType": "bar",
						"gitlabProvider": "baz"
					}
				}
			}
			`,
			assert: includes(`Did not find authentication provider matching type bar and configID foo`),
		},
		{
			kind: "GITLAB",
			desc: "valid external provider with SAML",
			config: `
			{
				"url": "https://gitlab.foo.bar",
				"authorization": {
					"identityProvider": {
						"type": "external",
						"authProviderID": "foo",
						"authProviderType": "bar",
						"gitlabProvider": "baz"
					}
				}
			}
			`,
			ps: []schema.AuthProviders{
				{
					Saml: &schema.SAMLAuthProvider{
						ConfigID: "foo",
						Type:     "bar",
					},
				},
			},
			assert: excludes(`Did not find authentication provider matching type bar and configID foo`),
		},
		{
			kind: "GITLAB",
			desc: "valid external provider with OIDC",
			config: `
			{
				"url": "https://gitlab.foo.bar",
				"authorization": {
					"identityProvider": {
						"type": "external",
						"authProviderID": "foo",
						"authProviderType": "bar",
						"gitlabProvider": "baz"
					}
				}
			}
			`,
			ps: []schema.AuthProviders{
				{
					Openidconnect: &schema.OpenIDConnectAuthProvider{
						ConfigID: "foo",
						Type:     "bar",
					},
				},
			},
			assert: excludes(`Did not find authentication provider matching type bar and configID foo`),
		},
		{
			kind: "GITLAB",
			desc: "username identity provider",
			config: `
			{
				"url": "https://gitlab.foo.bar",
				"token": "super-secret-token",
				"projectQuery": ["none"],
				"authorization": {
					"identityProvider": {
						"type": "username",
					}
				}
			}
			`,
			assert: equals("<nil>"),
		},
		{
			kind:   "PHABRICATOR",
			desc:   "without repos nor token",
			config: `{}`,
			assert: includes(
				`(root): Must validate at least one schema (anyOf)`,
				`token: token is required`,
			),
		},
		{
			kind:   "PHABRICATOR",
			desc:   "with empty repos",
			config: `{"repos": []}`,
			assert: includes(`repos: Array must have at least 1 items`),
		},
		{
			kind:   "PHABRICATOR",
			desc:   "with repos",
			config: `{"repos": [{"path": "gitolite/my/repo", "callsign": "MUX"}]}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "PHABRICATOR",
			desc:   "invalid token",
			config: `{"token": ""}`,
			assert: includes(`token: String length must be greater than or equal to 1`),
		},
		{
			kind:   "PHABRICATOR",
			desc:   "with token",
			config: `{"token": "a given token"}`,
			assert: equals(`<nil>`),
		},
		{
			kind:   "OTHER",
			desc:   "without url nor repos array",
			config: `{}`,
			assert: includes(`repos: repos is required`),
		},
		{
			kind:   "OTHER",
			desc:   "without URL but with null repos array",
			config: `{"repos": null}`,
			assert: includes(`repos: Invalid type. Expected: array, given: null`),
		},
		{
			kind:   "OTHER",
			desc:   "without URL but with empty repos array",
			config: `{"repos": []}`,
			assert: excludes(`repos: Array must have at least 1 items`),
		},
		{
			kind:   "OTHER",
			desc:   "without URL and empty repo array item",
			config: `{"repos": [""]}`,
			assert: includes(`repos.0: String length must be greater than or equal to 1`),
		},
		{
			kind:   "OTHER",
			desc:   "without URL and invalid repo array item",
			config: `{"repos": ["https://github.com/%%%%malformed"]}`,
			assert: includes(`repos.0: Does not match format 'uri-reference'`),
		},
		{
			kind:   "OTHER",
			desc:   "without URL and invalid scheme in repo array item",
			config: `{"repos": ["badscheme://github.com/my/repo"]}`,
			assert: includes(`repos.0: scheme "badscheme" not one of git, http, https or ssh`),
		},
		{
			kind:   "OTHER",
			desc:   "without URL and valid repos",
			config: `{"repos": ["http://git.hub/repo", "https://git.hub/repo", "git://user@hub.com:3233/repo.git/", "ssh://user@hub.com/repo.git/"]}`,
			assert: equals("<nil>"),
		},
		{
			kind:   "OTHER",
			desc:   "with URL but null repos array",
			config: `{"url": "http://github.com/", "repos": null}`,
			assert: includes(`repos: Invalid type. Expected: array, given: null`),
		},
		{
			kind:   "OTHER",
			desc:   "with URL but empty repos array",
			config: `{"url": "http://github.com/", "repos": []}`,
			assert: excludes(`repos: Array must have at least 1 items`),
		},
		{
			kind:   "OTHER",
			desc:   "with URL and empty repo array item",
			config: `{"url": "http://github.com/", "repos": [""]}`,
			assert: includes(`repos.0: String length must be greater than or equal to 1`),
		},
		{
			kind:   "OTHER",
			desc:   "with URL and invalid repo array item",
			config: `{"url": "https://github.com/", "repos": ["foo/%%%%malformed"]}`,
			assert: includes(`repos.0: Does not match format 'uri-reference'`),
		},
		{
			kind:   "OTHER",
			desc:   "with invalid scheme URL",
			config: `{"url": "badscheme://github.com/", "repos": ["my/repo"]}`,
			assert: includes(`url: Does not match pattern '^(git|ssh|https?)://'`),
		},
		{
			kind:   "OTHER",
			desc:   "with URL and valid repos",
			config: `{"url": "https://github.com/", "repos": ["foo/", "bar", "/baz", "bam.git"]}`,
			assert: equals("<nil>"),
		},
	} {
		tc := tc
		t.Run(tc.kind+"/"+tc.desc, func(t *testing.T) {
			var have []string
			if tc.ps == nil {
				tc.ps = conf.Get().Critical.AuthProviders
			}

			s := NewExternalServicesStore()
			err := s.ValidateConfig(tc.kind, tc.config, tc.ps)
			switch e := err.(type) {
			case nil:
				have = append(have, "<nil>")
			case *multierror.Error:
				for _, err := range e.Errors {
					have = append(have, err.Error())
				}
			default:
				have = append(have, err.Error())
			}

			tc.assert(t, have)
		})
	}
}
