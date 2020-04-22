package extsvc

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestCodeHostOf(t *testing.T) {
	for _, tc := range []struct {
		name      string
		repo      api.RepoName
		codehosts []*CodeHost
		want      *CodeHost
	}{{
		name:      "none",
		repo:      "github.com/foo/bar",
		codehosts: nil,
		want:      nil,
	}, {
		name:      "out",
		repo:      "github.com/foo/bar",
		codehosts: []*CodeHost{GitLabDotCom},
		want:      nil,
	}, {
		name:      "in",
		repo:      "github.com/foo/bar",
		codehosts: PublicCodeHosts,
		want:      GitHubDotCom,
	}, {
		name:      "case-insensitive",
		repo:      "GITHUB.COM/foo/bar",
		codehosts: PublicCodeHosts,
		want:      GitHubDotCom,
	}, {
		name:      "missing-path",
		repo:      "github.com",
		codehosts: PublicCodeHosts,
		want:      nil,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			have := CodeHostOf(tc.repo, tc.codehosts...)
			if have != tc.want {
				t.Errorf(
					"CodeHostOf(%q, %#v): want %#v, have %#v",
					tc.repo,
					tc.codehosts,
					tc.want,
					have,
				)
			}
		})
	}
}
