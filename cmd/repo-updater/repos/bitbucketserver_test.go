package repos

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
)

var update = flag.Bool("update", false, "update golden files")

func TestBitbucketServerRepoInfo(t *testing.T) {
	b, err := ioutil.ReadFile(filepath.Join("testdata", "bitbucketserver-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketserver.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	cases := map[string]*schema.BitbucketServerConnection{
		"simple": {
			Url:   "bitbucket.example.com",
			Token: "secret",
		},
		"ssh": {
			Url:                         "bitbucket.example.com",
			Token:                       "secret",
			InitialRepositoryEnablement: true,
			GitURLType:                  "ssh",
		},
		"path-pattern": {
			Url:                   "bitbucket.example.com",
			Token:                 "secret",
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
		},
	}
	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			var got []*protocol.RepoInfo
			for _, r := range repos {
				got = append(got, bitbucketServerRepoInfo(config, r))
			}
			actual, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			golden := filepath.Join("testdata", "bitbucketserver-repos-"+name+".golden")
			if *update {
				err := ioutil.WriteFile(golden, actual, 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			expect, err := ioutil.ReadFile(golden)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(actual, expect) {
				d, err := diff(actual, expect)
				if err != nil {
					t.Fatal(err)
				}
				t.Error(d)
			}
		})
	}
}

func diff(b1, b2 []byte) (string, error) {
	f1, err := ioutil.TempFile("", "repos_test")
	if err != nil {
		return "", err
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "repos_test")
	if err != nil {
		return "", err
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	_, _ = f1.Write(b1)
	_, _ = f2.Write(b2)

	data, err := exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		err = nil
	}
	return string(data), err
}
