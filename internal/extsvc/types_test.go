package extsvc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExtractToken(t *testing.T) {
	for _, tc := range []struct {
		config string
		kind   string
		want   string
	}{
		{
			config: `{"token": "deadbeef"}`,
			kind:   KindGitLab,
			want:   "deadbeef",
		},
		{
			config: `{"token": "deadbeef"}`,
			kind:   KindGitHub,
			want:   "deadbeef",
		},
		{
			config: `{"token": "deadbeef"}`,
			kind:   KindBitbucketServer,
			want:   "deadbeef",
		},
		{
			config: `{"token": "deadbeef"}`,
			kind:   KindPhabricator,
			want:   "deadbeef",
		},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			have, err := ExtractToken(tc.config, tc.kind)
			if err != nil {
				t.Fatal(err)
			}
			if have != tc.want {
				t.Errorf("Want %q, have %q", tc.want, have)
			}
		})
	}

	t.Run("fails for unsupported kind", func(t *testing.T) {
		_, err := ExtractToken(`{}`, KindGitolite)
		if err == nil {
			t.Fatal("expected an error for unsupported kind")
		}
	})
}

func TestExtractRateLimitConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
		kind   string
		want   rate.Limit
	}{
		{
			name:   "GitLab default",
			config: `{"url": "https://example.com/"}`,
			kind:   KindGitLab,
			want:   10.0,
		},
		{
			name:   "GitHub default",
			config: `{"url": "https://example.com/"}`,
			kind:   KindGitHub,
			want:   1.3888888888888888,
		},
		{
			name:   "Bitbucket Server default",
			config: `{"url": "https://example.com/"}`,
			kind:   KindBitbucketServer,
			want:   8.0,
		},
		{
			name:   "Bitbucket Cloud default",
			config: `{"url": "https://example.com/"}`,
			kind:   KindBitbucketCloud,
			want:   2.0,
		},
		{
			name:   "GitLab non-default",
			config: `{"url": "https://example.com/", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:   KindGitLab,
			want:   1.0,
		},
		{
			name:   "GitHub non-default",
			config: `{"url": "https://example.com/", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:   KindGitHub,
			want:   1.0,
		},
		{
			name:   "Bitbucket Server non-default",
			config: `{"url": "https://example.com/", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:   KindBitbucketServer,
			want:   1.0,
		},
		{
			name:   "Bitbucket Cloud non-default",
			config: `{"url": "https://example.com/", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:   KindBitbucketCloud,
			want:   1.0,
		},
		{
			name:   "NPM default",
			config: `{"registry": "https://registry.npmjs.org"}`,
			kind:   KindNpmPackages,
			want:   3000.0 / 3600.0,
		},
		{
			name:   "NPM non-default",
			config: `{"registry": "https://registry.npmjs.org", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:   KindNpmPackages,
			want:   1.0,
		},
		{
			name:   "Go mod default",
			config: `{"urls": ["https://example.com"]}`,
			kind:   KindGoModules,
			want:   57600.0 / 3600.0,
		},
		{
			name:   "Go mod non-default",
			config: `{"urls": ["https://example.com"], "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:   KindNpmPackages,
			want:   1.0,
		},
		{
			name:   "No trailing slash",
			config: `{"url": "https://example.com", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:   KindBitbucketCloud,
			want:   1.0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rlc, err := ExtractRateLimit(tc.config, tc.kind)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, rlc); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestEncodeURN(t *testing.T) {
	tests := []struct {
		desc    string
		kind    string
		id      int64
		wantURN string
	}{
		{
			desc:    "An empty kind and ID",
			kind:    "",
			id:      0,
			wantURN: "extsvc::0",
		},
		{
			desc:    "A valid kind and ID",
			kind:    "github.com",
			id:      1,
			wantURN: "extsvc:github.com:1",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			urn := URN(test.kind, test.id)
			if urn != test.wantURN {
				t.Fatalf("got urn %q, want %q", urn, test.wantURN)
			}
		})
	}
}

func TestDecodeURN(t *testing.T) {
	tests := []struct {
		desc     string
		urn      string
		wantKind string
		wantID   int64
	}{
		{
			desc:     "An empty string",
			urn:      "",
			wantKind: "",
			wantID:   0,
		},
		{
			desc:     "An incomplete URN",
			urn:      "extsvc:",
			wantKind: "",
			wantID:   0,
		},
		{
			desc:     "A valid complete URN",
			urn:      "extsvc:github.com:1",
			wantKind: "github.com",
			wantID:   1,
		},
		{
			desc:     "A valid URN with no kind",
			urn:      "extsvc::1",
			wantKind: "",
			wantID:   1,
		},
		{
			desc:     "A URN with floating-point ID",
			urn:      "extsvc:github.com:1.0",
			wantKind: "",
			wantID:   0,
		},
		{
			desc:     "A URN with string ID",
			urn:      "extsvc:github.com:fake",
			wantKind: "",
			wantID:   0,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			kind, id := DecodeURN(test.urn)
			if kind != test.wantKind {
				t.Errorf("got kind %q, want %q", kind, test.wantKind)
			}
			if id != test.wantID {
				t.Errorf("got id %d, want %d", id, test.wantID)
			}
		})
	}
}

func TestUniqueCodeHostIdentifier(t *testing.T) {
	for _, tc := range []struct {
		config string
		kind   string
		want   string
	}{
		{
			kind:   KindGitLab,
			config: `{"url": "https://example.com"}`,
			want:   "https://example.com/",
		},
		{
			kind:   KindGitHub,
			config: `{"url": "https://github.com"}`,
			want:   "https://github.com/",
		},
		{
			kind:   KindGitHub,
			config: `{"url": "https://github.example.com/"}`,
			want:   "https://github.example.com/",
		},
		{
			kind: KindAWSCodeCommit,
			config: `{
				"region": "eu-west-1",
				"accessKeyID": "accesskey",
				"secretAccessKey": "secretaccesskey",
				"gitCredentials": {
					"username": "my-user",
					"password": "my-password"
				}
			}`,
			want: "eu-west-1:accesskey",
		},
		{
			kind:   KindGerrit,
			config: `{"url": "https://example.com"}`,
			want:   "https://example.com/",
		},
		{
			kind:   KindBitbucketServer,
			config: `{"url": "https://bitbucket.sgdev.org/"}`,
			want:   "https://bitbucket.sgdev.org/",
		},

		{
			kind:   KindBitbucketCloud,
			config: `{"url": "https://bitbucket.org/"}`,
			want:   "https://bitbucket.org/",
		},

		{
			kind:   KindGitolite,
			config: `{"host": "ssh://git@gitolite.example.com:2222/"}`,
			want:   "ssh://git@gitolite.example.com:2222/",
		},
		{
			kind:   KindGitolite,
			config: `{"host": "git@gitolite.example.com"}`,
			want:   "git@gitolite.example.com/",
		},
		{
			kind:   KindPerforce,
			config: `{"p4.port": "ssl:111.222.333.444:1666"}`,
			want:   "ssl:111.222.333.444:1666",
		},
		{
			kind:   KindPhabricator,
			config: `{"url": "https://phabricator.sgdev.org/"}`,
			want:   "https://phabricator.sgdev.org/",
		},
		{
			kind:   KindOther,
			config: `{"url": "ssh://user@host.xz:2333/"}`,
			want:   "ssh://user@host.xz:2333/",
		},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			have, err := UniqueCodeHostIdentifier(tc.kind, tc.config)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, have); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestWebhookURL(t *testing.T) {
	const externalServiceID = 42
	const externalURL = "https://sourcegraph.com"

	t.Run("unknown kind", func(t *testing.T) {
		_, err := WebhookURL(KindOther, externalServiceID, nil, externalURL)
		assert.NotNil(t, err)
	})

	t.Run("basic kinds", func(t *testing.T) {
		for kind, want := range map[string]string{
			KindGitHub:          externalURL + "/.api/github-webhooks?externalServiceID=42",
			KindBitbucketServer: externalURL + "/.api/bitbucket-server-webhooks?externalServiceID=42",
			KindGitLab:          externalURL + "/.api/gitlab-webhooks?externalServiceID=42",
		} {
			t.Run(kind, func(t *testing.T) {
				// Note the use of a nil configuration here: these kinds do not
				// depend on the configuration being passed in or valid.
				have, err := WebhookURL(kind, externalServiceID, nil, externalURL)
				assert.Nil(t, err)
				assert.Equal(t, want, have)
			})
		}
	})

	t.Run("Bitbucket Cloud", func(t *testing.T) {
		t.Run("invalid configurations", func(t *testing.T) {
			for name, cfg := range map[string]any{
				"nil":               nil,
				"GitHub connection": &schema.GitHubConnection{},
			} {
				t.Run(name, func(t *testing.T) {
					_, err := WebhookURL(KindBitbucketCloud, externalServiceID, cfg, externalURL)
					assert.NotNil(t, err)
				})
			}
		})

		t.Run("valid configuration", func(t *testing.T) {
			have, err := WebhookURL(
				KindBitbucketCloud, externalServiceID,
				&schema.BitbucketCloudConnection{
					WebhookSecret: "foo bar",
				},
				externalURL,
			)
			assert.Nil(t, err)
			assert.Equal(t, externalURL+"/.api/bitbucket-cloud-webhooks?externalServiceID=42&secret=foo+bar", have)
		})
	})
}
