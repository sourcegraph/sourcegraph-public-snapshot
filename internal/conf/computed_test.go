package conf

import (
	"os"
	"strings"
	"testing"
	"time"

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

func setenv(t *testing.T, keyval string) func() {
	t.Helper()

	parts := strings.SplitN(keyval, "=", 2)
	key := parts[0]
	value := parts[1]

	orig, set := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
	if set {
		return func() {
			if err := os.Setenv(key, orig); err != nil {
				t.Fatal(err)
			}
		}
	}
	return func() {
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
