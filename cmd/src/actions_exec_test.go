package main

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCodeHostSupported(t *testing.T) {
	t.Run("error from getSourcegraphVersion", func(t *testing.T) {
		want := errors.New("foo")
		if _, have := isCodeHostSupportedForCampaignsImpl("GITLAB", func() (string, error) {
			return "", want
		}); !errors.Is(have, want) {
			t.Errorf("unexpected error: have %+v; want %+v", have, want)
		}
	})

	t.Run("error from sourcegraphVersionCheck", func(t *testing.T) {
		if _, err := isCodeHostSupportedForCampaignsImpl("GITLAB", func() (string, error) {
			return "x.y.z", nil
		}); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("no errors", func(t *testing.T) {
		for name, tc := range map[string]struct {
			want    bool
			kind    string
			version string
		}{
			"GitHub": {
				want:    true,
				kind:    "GITHUB",
				version: "1.2.3",
			},
			"BitBucket": {
				want:    true,
				kind:    "BITBUCKETSERVER",
				version: "1.2.3",
			},
			"GitLab with old semver version": {
				want:    false,
				kind:    "GITLAB",
				version: "1.2.3",
			},
			"GitLab with old dev version": {
				want:    false,
				kind:    "GITLAB",
				version: "68956_2019-07-21_c3a5992",
			},
			"GitLab with new semver version": {
				want:    true,
				kind:    "GITLAB",
				version: "3.18.0",
			},
			"GitLab with new dev version": {
				want:    true,
				kind:    "GITLAB",
				version: "68956_2020-07-21_c3a5992",
			},
			"unknown kind": {
				want:    false,
				kind:    "CODE HOSTS R US",
				version: "3.18.0",
			},
		} {
			t.Run(name, func(t *testing.T) {
				have, err := isCodeHostSupportedForCampaignsImpl(tc.kind, func() (string, error) {
					return tc.version, nil
				})
				if err != nil {
					t.Errorf("unexpected non-nil error: %+v", err)
				}

				if have != tc.want {
					t.Errorf("unexpected support status: have %v; want %v", have, tc.want)
				}
			})
		}
	})
}
