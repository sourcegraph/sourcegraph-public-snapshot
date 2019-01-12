package db

import (
	"fmt"
	"testing"
)

func TestExternalServices_ValidateConfig(t *testing.T) {
	for _, tc := range []struct {
		kind   string
		desc   string
		ext    externalServices
		config string
		err    string
	}{
		{
			kind:   "OTHER",
			desc:   "without url nor repos array",
			config: `{}`,
			err:    `required "repos" property is empty`,
		},
		{
			kind:   "OTHER",
			desc:   "without URL but with null repos array",
			config: `{"repos": null}`,
			err:    `required "repos" property is empty`,
		},
		{
			kind:   "OTHER",
			desc:   "without URL but with empty repos array",
			config: `{"repos": []}`,
			err:    `required "repos" property is empty`,
		},
		{
			kind:   "OTHER",
			desc:   "without URL and empty repo array item",
			config: `{"repos": [""]}`,
			err:    `invalid empty repos[0]`,
		},
		{
			kind:   "OTHER",
			desc:   "without URL and invalid repo array item",
			config: `{"repos": ["https://github.com/%%%%malformed"]}`,
			err:    `failed to parse repos[0]="https://github.com/%%%%malformed" with url="": parse https://github.com/%%%%malformed: invalid URL escape "%%%"`,
		},
		{
			kind:   "OTHER",
			desc:   "without URL and invalid scheme in repo array item",
			config: `{"repos": ["badscheme://github.com/my/repo"]}`,
			err:    `failed to parse repos[0]="badscheme://github.com/my/repo" with url="". scheme "badscheme" not one of git, http, https or ssh`,
		},
		{
			kind:   "OTHER",
			desc:   "without URL and valid repos",
			config: `{"repos": ["http://git.hub/repo", "https://git.hub/repo", "git://user@hub.com:3233/repo.git/", "ssh://user@hub.com/repo.git/"]}`,
			err:    "<nil>",
		},
		{
			kind:   "OTHER",
			desc:   "with URL but null repos array",
			config: `{"url": "http://github.com/", "repos": null}`,
			err:    `required "repos" property is empty`,
		},
		{
			kind:   "OTHER",
			desc:   "with URL but empty repos array",
			config: `{"url": "http://github.com/", "repos": []}`,
			err:    `required "repos" property is empty`,
		},
		{
			kind:   "OTHER",
			desc:   "with URL and empty repo array item",
			config: `{"url": "http://github.com/", "repos": [""]}`,
			err:    `invalid empty repos[0]`,
		},
		{
			kind:   "OTHER",
			desc:   "with URL and invalid repo array item",
			config: `{"url": "https://github.com/", "repos": ["foo/%%%%malformed"]}`,
			err:    `failed to parse repos[0]="foo/%%%%malformed" with url="https://github.com/": parse foo/%%%%malformed: invalid URL escape "%%%"`,
		},
		{
			kind:   "OTHER",
			desc:   "with invalid scheme URL",
			config: `{"url": "badscheme://github.com/", "repos": ["my/repo"]}`,
			err:    `failed to parse repos[0]="my/repo" with url="badscheme://github.com/". scheme "badscheme" not one of git, http, https or ssh`,
		},
		{
			kind:   "OTHER",
			desc:   "with URL and valid repos",
			config: `{"url": "https://github.com/", "repos": ["foo/", "bar", "/baz", "bam.git"]}`,
			err:    "<nil>",
		},
	} {
		tc := tc
		t.Run(tc.kind+"/"+tc.desc, func(t *testing.T) {
			t.Parallel()

			err := tc.ext.validateConfig(tc.kind, tc.config)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("validateConfig(%q, %s):\nhave: %s\nwant: %s", tc.kind, tc.config, have, want)
			}
		})
	}
}
