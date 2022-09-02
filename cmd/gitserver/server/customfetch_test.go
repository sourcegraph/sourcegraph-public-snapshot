package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestEmptyCustomGitFetch(t *testing.T) {
	remoteURL, _ := vcs.ParseURL("git@github.com:sourcegraph/sourcegraph.git")
	customCmd := customFetchCmd(context.Background(), remoteURL)
	if customCmd != nil {
		t.Errorf("expected nil custom cmd for empty configuration, got %+v", customCmd)
	}
}

func TestCustomGitFetch(t *testing.T) {
	// mock value for test
	customGitFetchCmdConf = filepath.Join(t.TempDir(), "customGitConfig.json")

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
		{
			DomainPath: "github.com/foo/absolute",
			Fetch:      "/foo/bar/git fetch things",
		},
	}

	c, err := json.Marshal(mappings)
	if err != nil {
		t.Fatal(err)
	}

	// write config file to disk
	err = ioutil.WriteFile(customGitFetchCmdConf, c, 0644)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.Remove(customGitFetchCmdConf)
	})

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
		{
			url:          "https://8cd1419f4d5c1e0527f2893c9422f1a2a435116d@github.com/foo/absolute",
			expectedArgs: []string{"/foo/bar/git", "fetch", "things"},
		},
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
