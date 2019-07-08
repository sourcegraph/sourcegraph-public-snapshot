package repos

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestBitbucketCloudSource_MakeRepo(t *testing.T) {
	b, err := ioutil.ReadFile(filepath.Join("testdata", "bitbucketcloud-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*bitbucketcloud.Repo
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	cases := map[string]*schema.BitbucketCloudConnection{
		"simple": {
			Url:         "bitbucket.org",
			Username:    "alice",
			AppPassword: "secret",
		},
		"ssh": {
			Url:                         "bitbucket.org",
			Username:                    "alice",
			AppPassword:                 "secret",
			GitURLType:                  "ssh",
			InitialRepositoryEnablement: true,
		},
		"path-pattern": {
			Url:                   "bitbucket.org",
			Username:              "alice",
			AppPassword:           "secret",
			RepositoryPathPattern: "bb/{nameWithOwner}",
		},
	}

	svc := ExternalService{ID: 1, Kind: "BITBUCKETCLOUD"}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := newBitbucketCloudSource(&svc, config, nil)
			if err != nil {
				t.Fatal(err)
			}

			var got []*Repo
			for _, r := range repos {
				got = append(got, s.makeRepo(r))
			}
			actual, err := json.MarshalIndent(got, "", "  ")
			if err != nil {
				t.Fatal(err)
			}

			golden := filepath.Join("testdata", "bitbucketcloud-repos-"+name+".golden")
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
