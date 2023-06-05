package conf

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthPasswordResetLinkDuration(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		want int
	}{{
		name: "password link expiry has a default value if null",
		sc:   &Unified{},
		want: defaultPasswordLinkExpiry,
	}, {
		name: "password link expiry has a default value if blank",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{AuthPasswordResetLinkExpiry: 0}},
		want: defaultPasswordLinkExpiry,
	}, {
		name: "password link expiry can be customized",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{AuthPasswordResetLinkExpiry: 60}},
		want: 60,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			if got, want := AuthPasswordResetLinkExpiry(), test.want; got != want {
				t.Fatalf("AuthPasswordResetLinkExpiry() = %v, want %v", got, want)
			}
		})
	}
}

func TestGitLongCommandTimeout(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		want time.Duration
	}{{
		name: "Git long command timeout has a default value if null",
		sc:   &Unified{},
		want: defaultGitLongCommandTimeout,
	}, {
		name: "Git long command timeout has a default value if blank",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitLongCommandTimeout: 0}},
		want: defaultGitLongCommandTimeout,
	}, {
		name: "Git long command timeout can be customized",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitLongCommandTimeout: 60}},
		want: time.Duration(60) * time.Second,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			if got, want := GitLongCommandTimeout(), test.want; got != want {
				t.Fatalf("GitLongCommandTimeout() = %v, want %v", got, want)
			}
		})
	}
}

func TestGitMaxCodehostRequestsPerSecond(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		want int
	}{
		{
			name: "not set should return default",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{}},
			want: -1,
		},
		{
			name: "bad value should return default",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitMaxCodehostRequestsPerSecond: intPtr(-100)}},
			want: -1,
		},
		{
			name: "set 0 should return 0",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitMaxCodehostRequestsPerSecond: intPtr(0)}},
			want: 0,
		},
		{
			name: "set non-0 should return non-0",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{GitMaxCodehostRequestsPerSecond: intPtr(100)}},
			want: 100,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			if got, want := GitMaxCodehostRequestsPerSecond(), test.want; got != want {
				t.Fatalf("GitMaxCodehostRequestsPerSecond() = %v, want %v", got, want)
			}
		})
	}
}

func TestGitMaxConcurrentClones(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		want int
	}{
		{
			name: "not set should return default",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{}},
			want: 5,
		},
		{
			name: "bad value should return default",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					GitMaxConcurrentClones: -100,
				},
			},
			want: 5,
		},
		{
			name: "set non-zero should return non-zero",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					GitMaxConcurrentClones: 100,
				},
			},
			want: 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			if got, want := GitMaxConcurrentClones(), test.want; got != want {
				t.Fatalf("GitMaxConcurrentClones() = %v, want %v", got, want)
			}
		})
	}
}

func TestAuthLockout(t *testing.T) {
	defer Mock(nil)

	tests := []struct {
		name string
		mock *schema.AuthLockout
		want *schema.AuthLockout
	}{
		{
			name: "missing entire config",
			mock: nil,
			want: &schema.AuthLockout{
				ConsecutivePeriod:      3600,
				FailedAttemptThreshold: 5,
				LockoutPeriod:          1800,
			},
		},
		{
			name: "missing all fields",
			mock: &schema.AuthLockout{},
			want: &schema.AuthLockout{
				ConsecutivePeriod:      3600,
				FailedAttemptThreshold: 5,
				LockoutPeriod:          1800,
			},
		},
		{
			name: "missing some fields",
			mock: &schema.AuthLockout{
				ConsecutivePeriod: 7200,
			},
			want: &schema.AuthLockout{
				ConsecutivePeriod:      7200,
				FailedAttemptThreshold: 5,
				LockoutPeriod:          1800,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(&Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthLockout: test.mock,
				},
			})

			got := AuthLockout()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestGitHubAppConfig(t *testing.T) {
	tests := []struct {
		name    string
		sc      *Unified
		want    GitHubAppConfiguration
		wantErr bool
	}{
		{
			name: "not set should return default",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{}},
			want: GitHubAppConfiguration{},
		},
		{
			name: "bad value should return error",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					GitHubApp: &schema.GitHubApp{PrivateKey: "f00b4r"},
				},
			},
			wantErr: true,
		},
		{
			name: "configured should return configured",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					GitHubApp: &schema.GitHubApp{
						PrivateKey:   `LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlCUEFJQkFBSkJBUEpIaWprdG1UMUlLYUd0YTVFZXAzQVo5Q2VPZUw4alBESUZUN3dRZ0tabXQzRUZxRGhCCk93bitRVUhKdUs5Zm92UkROSmVWTDJvWTVCT0l6NHJ3L0cwQ0F3RUFBUUpCQU1BK0o5Mks0d2NQVllsbWMrM28KcHU5NmlKTkNwMmp5Nm5hK1pEQlQzK0VvSUo1VFJGdnN3R2kvTHUzZThYUWwxTDNTM21ub0xPSlZNcTF0bUxOMgpIY0VDSVFEK3daeS83RlYxUEFtdmlXeWlYVklETzJnNWJOaUJlbmdKQ3hFa3Nia1VtUUloQVBOMlZaczN6UFFwCk1EVG9vTlJXcnl0RW1URERkamdiOFpzTldYL1JPRGIxQWlCZWNKblNVQ05TQllLMXJ5VTFmNURTbitoQU9ZaDkKWDFBMlVnTDE3bWhsS1FJaEFPK2JMNmRDWktpTGZORWxmVnRkTUtxQnFjNlBIK01heFU2VzlkVlFvR1dkQWlFQQptdGZ5cE9zYTFiS2hFTDg0blovaXZFYkJyaVJHalAya3lERHYzUlg0V0JrPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`,
						ClientID:     "1",
						ClientSecret: "hush",
						Slug:         "slugs-are-cool",
						AppID:        "99",
					},
				},
			},
			want: GitHubAppConfiguration{
				ClientID:     "1",
				ClientSecret: "hush",
				Slug:         "slugs-are-cool",
				AppID:        "99",
				PrivateKey: []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBPAIBAAJBAPJHijktmT1IKaGta5Eep3AZ9CeOeL8jPDIFT7wQgKZmt3EFqDhB
Own+QUHJuK9fovRDNJeVL2oY5BOIz4rw/G0CAwEAAQJBAMA+J92K4wcPVYlmc+3o
pu96iJNCp2jy6na+ZDBT3+EoIJ5TRFvswGi/Lu3e8XQl1L3S3mnoLOJVMq1tmLN2
HcECIQD+wZy/7FV1PAmviWyiXVIDO2g5bNiBengJCxEksbkUmQIhAPN2VZs3zPQp
MDTooNRWrytEmTDDdjgb8ZsNWX/RODb1AiBecJnSUCNSBYK1ryU1f5DSn+hAOYh9
X1A2UgL17mhlKQIhAO+bL6dCZKiLfNElfVtdMKqBqc6PH+MaxU6W9dVQoGWdAiEA
mtfypOsa1bKhEL84nZ/ivEbBriRGjP2kyDDv3RX4WBk=
-----END RSA PRIVATE KEY-----
`),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			have, err := GitHubAppConfig()
			if err != nil && !test.wantErr {
				t.Fatalf("unexpected err: %s", err)
			}
			if err == nil && test.wantErr {
				t.Fatal("want err but got none")
			}
			if diff := cmp.Diff(have, test.want); diff != "" {
				t.Fatalf("GitHubAppConfig() wrong: %s", diff)
			}
		})
	}
}

func TestIsAccessRequestEnabled(t *testing.T) {
	falseVal, trueVal := false, true
	tests := []struct {
		name string
		sc   *Unified
		want bool
	}{
		{
			name: "not set should return default true",
			sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{}},
			want: true,
		},
		{
			name: "parent object set should return default true",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthAccessRequest: &schema.AuthAccessRequest{},
				},
			},
			want: true,
		},
		{
			name: "explicitly set enabled=true should return true",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthAccessRequest: &schema.AuthAccessRequest{Enabled: &trueVal},
				},
			},
			want: true,
		},
		{
			name: "explicitly set enabled=false should return false",
			sc: &Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthAccessRequest: &schema.AuthAccessRequest{
						Enabled: &falseVal,
					},
				},
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Mock(test.sc)
			have := IsAccessRequestEnabled()
			assert.Equal(t, test.want, have)
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
