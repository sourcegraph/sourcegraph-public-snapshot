package phabricator_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/google/go-cmp/cmp"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

var update = flag.Bool("update", false, "update testdata")

func TestClient_ListRepos(t *testing.T) {
	cli, save := newClient(t, "ListRepos")
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	for _, tc := range []struct {
		name   string
		ctx    context.Context
		args   phabricator.ListReposArgs
		cursor *phabricator.Cursor
		err    string
	}{
		{
			name:   "repos-listed",
			args:   phabricator.ListReposArgs{Cursor: &phabricator.Cursor{Limit: 5}},
			cursor: &phabricator.Cursor{Limit: 5, After: "5", Order: "oldest"},
		},
		{
			name: "pagination",
			args: phabricator.ListReposArgs{
				Cursor: &phabricator.Cursor{
					Limit: 5,
					After: "5",
					Order: "oldest",
				},
			},
			cursor: &phabricator.Cursor{
				Limit:  5,
				After:  "19",
				Before: "8",
				Order:  "oldest",
			},
		},
		{
			name: "timeout",
			ctx:  timeout,
			err:  `Post "https://secure.phabricator.com/api/diffusion.repository.search": context deadline exceeded`,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			repos, cursor, err := cli.ListRepos(tc.ctx, tc.args)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if have, want := cursor, tc.cursor; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}

			if tc.args == (phabricator.ListReposArgs{}) {
				return
			}

			bs, err := json.MarshalIndent(repos, "", "  ")
			if err != nil {
				t.Fatalf("failed to marshal repos: %s", err)
			}

			path := fmt.Sprintf("testdata/golden/ListRepos-%s.json", tc.name)
			if *update {
				if err = ioutil.WriteFile(path, bs, 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := string(bs), string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func TestClient_GetRawDiff(t *testing.T) {
	cli, save := newClient(t, "GetRawDiff")
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	for _, tc := range []struct {
		name string
		ctx  context.Context
		id   int
		err  string
	}{{
		name: "diff not found",
		id:   0xdeadbeef,
		err:  "ERR_NOT_FOUND: Diff not found.",
	}, {
		name: "diff found",
		id:   20455,
	}, {
		name: "timeout",
		ctx:  timeout,
		err:  `Post "https://secure.phabricator.com/api/differential.getrawdiff": context deadline exceeded`,
	}} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			diff, err := cli.GetRawDiff(tc.ctx, tc.id)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.id == 0 {
				return
			}

			path := "testdata/golden/GetRawDiff-" + strconv.Itoa(tc.id)
			if *update {
				if err = ioutil.WriteFile(path, []byte(diff), 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := diff, string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func TestClient_GetDiffInfo(t *testing.T) {
	cli, save := newClient(t, "GetDiffInfo")
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	for _, tc := range []struct {
		name string
		ctx  context.Context
		id   int
		info *phabricator.DiffInfo
		err  string
	}{{
		name: "diff not found",
		id:   0xdeadbeef,
		err:  "phabricator error: no diff info found for diff 3735928559",
	}, {
		name: "diff info found",
		id:   20455,
		info: &phabricator.DiffInfo{
			AuthorName:  "epriestley",
			AuthorEmail: "git@epriestley.com",
			DateCreated: "1395874084",
			Date:        time.Unix(1395874084, 0).UTC(),
		},
	}, {
		name: "timeout",
		ctx:  timeout,
		err:  `Post "https://secure.phabricator.com/api/differential.querydiffs": context deadline exceeded`,
	}} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			info, err := cli.GetDiffInfo(tc.ctx, tc.id)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if have, want := info, tc.info; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		})
	}
}

func newClient(t testing.TB, name string) (*phabricator.Client, func()) {
	t.Helper()

	cassete := filepath.Join("testdata/vcr/", strings.Replace(name, " ", "-", -1))
	rec, err := httptestutil.NewRecorder(cassete, *update, func(i *cassette.Interaction) error {
		// Remove all tokens
		i.Request.Body = ""
		i.Request.Form = map[string][]string{}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	hc, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	cli, err := phabricator.NewClient(
		ctx,
		"https://secure.phabricator.com",
		os.Getenv("PHABRICATOR_TOKEN"),
		hc,
	)
	if err != nil {
		t.Fatal(err)
	}

	return cli, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	}
}
