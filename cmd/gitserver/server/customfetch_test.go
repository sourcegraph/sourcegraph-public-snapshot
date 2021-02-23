package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestEmptyCustomGitFetch(t *testing.T) {
	customGitFetch = func() interface{} {
		return buildCustomFetchMappings(nil)
	}

	remoteURL, _ := vcs.ParseURL("git@github.com:sourcegraph/sourcegraph.git")
	customCmd := customFetchCmd(context.Background(), remoteURL)
	if customCmd != nil {
		t.Errorf("expected nil custom cmd for empty configuration, got %+v", customCmd)
	}
}

func TestCustomGitFetch(t *testing.T) {
	mappings := []*schema.CustomGitFetchMapping{
		{
			DomainPath: "github.com/foo/normal/one",
			Fetch:      "echo normal one",
		},
		{
			DomainPath: "github.com/foo/normal/two",
			Fetch:      "echo normal two",
		},
		{
			DomainPath: "github.com/foo/faulty",
			Fetch:      "",
		},
	}

	tests := []struct {
		url          string
		expectedArgs []string
	}{
		{
			url:          "https://8cd1419f4d5c1e0527f2893c9422f1a2a435116d@github.com/foo/normal/one",
			expectedArgs: []string{"echo", "normal", "one"},
		},
		{
			url:          "https://8cd1419f4d5c1e0527f2893c9422f1a2a435116d@github.com/foo/normal/two",
			expectedArgs: []string{"echo", "normal", "two"},
		},
		{
			url: "https://8cd1419f4d5c1e0527f2893c9422f1a2a435116d@github.com/foo/faulty",
		},
		{
			url: "https://8cd1419f4d5c1e0527f2893c9422f1a2a435116dgit@github.com/bar/notthere",
		},
	}

	customGitFetch = func() interface{} {
		return buildCustomFetchMappings(mappings)
	}

	for _, test := range tests {
		remoteURL, _ := vcs.ParseURL(test.url)
		customCmd := customFetchCmd(context.Background(), remoteURL)
		var args []string
		if customCmd != nil {
			args = customCmd.Args
		}

		if diff := cmp.Diff(test.expectedArgs, args); diff != "" {
			t.Errorf("URL %q: %v", test.url, diff)
		}
	}
}
