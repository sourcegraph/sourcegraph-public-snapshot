package extsvc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExtractRateLimitConfig(t *testing.T) {
	for _, tc := range []struct {
		name        string
		config      string
		kind        string
		displayName string
		want        RateLimitConfig
	}{
		{
			name:        "GitLab default",
			config:      `{"url": "https://example.com/"}`,
			kind:        KindGitLab,
			displayName: "GitLab 1",
			want: RateLimitConfig{
				BaseURL:     "https://example.com/",
				DisplayName: "GitLab 1",
				Limit:       10.0,
				IsDefault:   true,
			},
		},
		{
			name:        "GitHub default",
			config:      `{"url": "https://example.com/"}`,
			kind:        KindGitHub,
			displayName: "GitHub 1",
			want: RateLimitConfig{
				BaseURL:     "https://example.com/",
				DisplayName: "GitHub 1",
				Limit:       1.3888888888888888,
				IsDefault:   true,
			},
		},
		{
			name:        "Bitbucket Server default",
			config:      `{"url": "https://example.com/"}`,
			kind:        KindBitbucketServer,
			displayName: "BitbucketServer 1",
			want: RateLimitConfig{
				BaseURL:     "https://example.com/",
				DisplayName: "BitbucketServer 1",
				Limit:       8.0,
				IsDefault:   true,
			},
		},
		{
			name:        "Bitbucket Cloud default",
			config:      `{"url": "https://example.com/"}`,
			kind:        KindBitbucketCloud,
			displayName: "BitbucketCloud 1",
			want: RateLimitConfig{
				BaseURL:     "https://example.com/",
				DisplayName: "BitbucketCloud 1",
				Limit:       2.0,
				IsDefault:   true,
			},
		},
		{
			name:        "GitLab non-default",
			config:      `{"url": "https://example.com/", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:        KindGitLab,
			displayName: "GitLab 1",
			want: RateLimitConfig{
				BaseURL:     "https://example.com/",
				DisplayName: "GitLab 1",
				Limit:       1.0,
				IsDefault:   false,
			},
		},
		{
			name:        "GitHub default",
			config:      `{"url": "https://example.com/", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:        KindGitHub,
			displayName: "GitHub 1",
			want: RateLimitConfig{
				BaseURL:     "https://example.com/",
				DisplayName: "GitHub 1",
				Limit:       1.0,
				IsDefault:   false,
			},
		},
		{
			name:        "Bitbucket Server default",
			config:      `{"url": "https://example.com/", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:        KindBitbucketServer,
			displayName: "BitbucketServer 1",
			want: RateLimitConfig{
				BaseURL:     "https://example.com/",
				DisplayName: "BitbucketServer 1",
				Limit:       1.0,
				IsDefault:   false,
			},
		},
		{
			name:        "Bitbucket Cloud default",
			config:      `{"url": "https://example.com/", "rateLimit": {"enabled": true, "requestsPerHour": 3600}}`,
			kind:        KindBitbucketCloud,
			displayName: "BitbucketCloud 1",
			want: RateLimitConfig{
				BaseURL:     "https://example.com/",
				DisplayName: "BitbucketCloud 1",
				Limit:       1.0,
				IsDefault:   false,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rlc, err := ExtractRateLimitConfig(tc.config, tc.kind, tc.displayName)
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
