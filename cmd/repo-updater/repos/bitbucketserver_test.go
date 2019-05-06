package repos

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
)

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
		"username": {
			Url:                   "bitbucket.example.com",
			Username:              "foo",
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
			if update(name) {
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

func TestBitbucketServerExclude(t *testing.T) {
	b, err := ioutil.ReadFile(filepath.Join("testdata", "bitbucketserver-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketserver.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	cases := map[string]*schema.BitbucketServerConnection{
		"none": {
			Url:   "bitbucket.example.com",
			Token: "secret",
		},
		"name": {
			Url:   "bitbucket.example.com",
			Token: "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Name: "SG/python-langserver-fork",
			}, {
				Name: "~KEEGAN/rgp",
			}},
		},
		"id": {
			Url:     "bitbucket.example.com",
			Token:   "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{Id: 4}},
		},
		"pattern": {
			Url:   "bitbucket.example.com",
			Token: "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Pattern: "SG/python.*",
			}, {
				Pattern: "~KEEGAN/.*",
			}},
		},
		"both": {
			Url:   "bitbucket.example.com",
			Token: "secret",
			// We match on the bitbucket server repo name, not the repository path pattern.
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Id: 1,
			}, {
				Name: "~KEEGAN/rgp",
			}, {
				Pattern: ".*-fork",
			}},
		},
	}
	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			conn, err := newBitbucketServerConnection(config, nil)
			if err != nil {
				t.Fatal(err)
			}

			type output struct {
				Include []string
				Exclude []string
			}
			var got output
			for _, r := range repos {
				name := r.Slug
				if r.Project != nil {
					name = r.Project.Key + "/" + name
				}
				if conn.excludes(r) {
					got.Exclude = append(got.Exclude, name)
				} else {
					got.Include = append(got.Include, name)
				}
			}
			actual, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			golden := filepath.Join("testdata", "bitbucketserver-repos-exclude-"+name+".golden")
			if update(name) {
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
