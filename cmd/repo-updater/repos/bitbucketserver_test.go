package repos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func TestBitbucketServerSource_MakeRepo(t *testing.T) {
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
			Url:                         "https://bitbucket.example.com",
			Token:                       "secret",
			InitialRepositoryEnablement: true,
			GitURLType:                  "ssh",
		},
		"path-pattern": {
			Url:                   "https://bitbucket.example.com",
			Token:                 "secret",
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
		},
		"username": {
			Url:                   "https://bitbucket.example.com",
			Username:              "foo",
			Token:                 "secret",
			RepositoryPathPattern: "bb/{projectKey}/{repositorySlug}",
		},
	}

	svc := ExternalService{ID: 1, Kind: "BITBUCKETSERVER"}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := newBitbucketServerSource(&svc, config, nil)
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
				d, err := testutil.Diff(string(actual), string(expect))
				if err != nil {
					t.Fatal(err)
				}
				t.Error(d)
			}
		})
	}
}

func TestBitbucketServerSource_Exclude(t *testing.T) {
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
			Url:   "https://bitbucket.example.com",
			Token: "secret",
		},
		"name": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Name: "SG/python-langserver-fork",
			}, {
				Name: "~KEEGAN/rgp",
			}},
		},
		"id": {
			Url:     "https://bitbucket.example.com",
			Token:   "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{Id: 4}},
		},
		"pattern": {
			Url:   "https://bitbucket.example.com",
			Token: "secret",
			Exclude: []*schema.ExcludedBitbucketServerRepo{{
				Pattern: "SG/python.*",
			}, {
				Pattern: "~KEEGAN/.*",
			}},
		},
		"both": {
			Url:   "https://bitbucket.example.com",
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

	svc := ExternalService{ID: 1, Kind: "BITBUCKETSERVER"}

	for name, config := range cases {
		t.Run(name, func(t *testing.T) {
			s, err := newBitbucketServerSource(&svc, config, nil)
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
				if s.excludes(r) {
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
				d, err := testutil.Diff(string(actual), string(expect))
				if err != nil {
					t.Fatal(err)
				}
				t.Error(d)
			}
		})
	}
}

func TestBitbucketServerSource_LoadChangesets(t *testing.T) {
	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	repo := &Repo{
		Metadata: &bitbucketserver.Repo{
			Slug:    "vegeta",
			Project: &bitbucketserver.Project{Key: "SOUR"},
		},
	}

	testCases := []struct {
		name string
		cs   []*Changeset
		err  string
	}{
		{
			name: "found",
			cs: []*Changeset{
				{Repo: repo, Changeset: &a8n.Changeset{ExternalID: "2"}},
				{Repo: repo, Changeset: &a8n.Changeset{ExternalID: "4"}},
			},
		},
		{
			name: "subset-not-found",
			cs: []*Changeset{
				{Repo: repo, Changeset: &a8n.Changeset{ExternalID: "2"}},
				{Repo: repo, Changeset: &a8n.Changeset{ExternalID: "999"}},
			},
			err: "Bitbucket API HTTP error: code=404 url=\"${INSTANCEURL}/rest/api/1.0/projects/SOUR/repos/vegeta/pull-requests/999\" body=\"{\\\"errors\\\":[{\\\"context\\\":null,\\\"message\\\":\\\"Pull request 999 does not exist in SOUR/vegeta.\\\",\\\"exceptionName\\\":\\\"com.atlassian.bitbucket.pull.NoSuchPullRequestException\\\"}]}\"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_LoadChangesets_" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &ExternalService{
				Kind: "BITBUCKETSERVER",
				Config: marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   instanceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				}),
			}

			bbsSrc, err := NewBitbucketServerSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			err = bbsSrc.LoadChangesets(ctx, tc.cs...)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			meta := make([]*bitbucketserver.PullRequest, 0, len(tc.cs))
			for _, cs := range tc.cs {
				meta = append(meta, cs.Changeset.Metadata.(*bitbucketserver.PullRequest))
			}

			data, err := json.MarshalIndent(meta, " ", " ")
			if err != nil {
				t.Fatal(err)
			}

			path := "testdata/golden/" + tc.name
			if update(tc.name) {
				if err = ioutil.WriteFile(path, data, 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := string(data), string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func TestBitbucketServerSource_CreateChangeset(t *testing.T) {
	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		// The test fixtures and golden files were generated with
		// this config pointed to bitbucket.sgdev.org
		instanceURL = "https://bitbucket.sgdev.org"
	}

	repo := &Repo{
		Metadata: &bitbucketserver.Repo{
			Slug:    "automation-testing",
			Project: &bitbucketserver.Project{Key: "SOUR"},
		},
	}

	testCases := []struct {
		name string
		cs   *Changeset
		err  string
	}{
		{
			name: "abbreviated refs",
			cs: &Changeset{
				Title:     "This is a test PR",
				Body:      "This is the body of a test PR",
				BaseRef:   "master",
				HeadRef:   "test-pr-bbs-9",
				Repo:      repo,
				Changeset: &a8n.Changeset{},
			},
		},
		{
			name: "success",
			cs: &Changeset{
				Title:     "This is a test PR",
				Body:      "This is the body of a test PR",
				BaseRef:   "refs/heads/master",
				HeadRef:   "refs/heads/test-pr-bbs-10",
				Repo:      repo,
				Changeset: &a8n.Changeset{},
			},
		},
		{
			name: "already exists",
			cs: &Changeset{
				Title:     "This is a test PR",
				Body:      "This is the body of a test PR",
				BaseRef:   "refs/heads/master",
				HeadRef:   "refs/heads/always-open-pr-bbs",
				Repo:      repo,
				Changeset: &a8n.Changeset{},
			},
			// CreateChangeset is idempotent so if the PR already exists
			// it is not an error
			err: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "BitbucketServerSource_CreateChangeset_" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &ExternalService{
				Kind: "BITBUCKETSERVER",
				Config: marshalJSON(t, &schema.BitbucketServerConnection{
					Url:   instanceURL,
					Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				}),
			}

			bbsSrc, err := NewBitbucketServerSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			if tc.err == "" {
				tc.err = "<nil>"
			}

			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			err = bbsSrc.CreateChangeset(ctx, tc.cs)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			pr := tc.cs.Changeset.Metadata.(*bitbucketserver.PullRequest)
			data, err := json.MarshalIndent(pr, " ", " ")
			if err != nil {
				t.Fatal(err)
			}

			path := "testdata/golden/" + tc.name
			if update(tc.name) {
				if err = ioutil.WriteFile(path, data, 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := string(data), string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}
