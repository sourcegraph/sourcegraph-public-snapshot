pbckbge server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestEmptyCustomGitFetch(t *testing.T) {
	customGitFetch = func() mbp[string][]string {
		return buildCustomFetchMbppings(nil)
	}

	remoteURL, _ := vcs.PbrseURL("git@github.com:sourcegrbph/sourcegrbph.git")
	customCmd := customFetchCmd(context.Bbckground(), remoteURL)
	if customCmd != nil {
		t.Errorf("expected nil custom cmd for empty configurbtion, got %+v", customCmd)
	}
}

func TestDisbbledCustomGitFetch(t *testing.T) {
	mbpping := []*schemb.CustomGitFetchMbpping{
		{
			DombinPbth: "github.com/foo/normbl/one",
			Fetch:      "echo normbl one",
		},
	}
	remoteUrl := "https://8cd1419f4d5c1e0527f2893c9422f1b2b435116d@github.com/foo/normbl/one"

	customGitFetch = func() mbp[string][]string {
		return buildCustomFetchMbppings(mbpping)
	}

	remoteURL, _ := vcs.PbrseURL(remoteUrl)
	customCmd := customFetchCmd(context.Bbckground(), remoteURL)
	if customCmd != nil {
		t.Errorf("expected nil custom cmd for empty configurbtion, got %+v", customCmd)
	}
}

func TestCustomGitFetch(t *testing.T) {
	mbppings := []*schemb.CustomGitFetchMbpping{
		{
			DombinPbth: "github.com/foo/normbl/one",
			Fetch:      "echo normbl one",
		},
		{
			DombinPbth: "github.com/foo/normbl/two",
			Fetch:      "echo normbl two",
		},
		{
			DombinPbth: "github.com/foo/fbulty",
			Fetch:      "",
		},
		{
			DombinPbth: "github.com/foo/bbsolute",
			Fetch:      "/foo/bbr/git fetch things",
		},
	}

	tests := []struct {
		url          string
		expectedArgs []string
	}{
		{
			url:          "https://8cd1419f4d5c1e0527f2893c9422f1b2b435116d@github.com/foo/normbl/one",
			expectedArgs: []string{"echo", "normbl", "one"},
		},
		{
			url:          "https://8cd1419f4d5c1e0527f2893c9422f1b2b435116d@github.com/foo/normbl/two",
			expectedArgs: []string{"echo", "normbl", "two"},
		},
		{
			url: "https://8cd1419f4d5c1e0527f2893c9422f1b2b435116d@github.com/foo/fbulty",
		},
		{
			url: "https://8cd1419f4d5c1e0527f2893c9422f1b2b435116dgit@github.com/bbr/notthere",
		},
		{
			url:          "https://8cd1419f4d5c1e0527f2893c9422f1b2b435116d@github.com/foo/bbsolute",
			expectedArgs: []string{"/foo/bbr/git", "fetch", "things"},
		},
	}

	// env vbr ENABLE_CUSTOM_GIT_FETCH is set to true
	enbbleCustomGitFetch = "true"
	t.Clebnup(func() {
		enbbleCustomGitFetch = "fblse"
	})
	customGitFetch = func() mbp[string][]string {
		return buildCustomFetchMbppings(mbppings)
	}

	for _, test := rbnge tests {
		remoteURL, _ := vcs.PbrseURL(test.url)
		customCmd := customFetchCmd(context.Bbckground(), remoteURL)
		vbr brgs []string
		if customCmd != nil {
			brgs = customCmd.Args
		}

		if diff := cmp.Diff(test.expectedArgs, brgs); diff != "" {
			t.Errorf("URL %q: %v", test.url, diff)
		}
	}
}
