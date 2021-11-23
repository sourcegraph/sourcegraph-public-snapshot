package repos

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestPagureSource_ListRepos(t *testing.T) {
	assertAllReposListed := func(want []string) types.ReposAssertion {
		return func(t testing.TB, rs types.Repos) {
			t.Helper()

			have := rs.Names()
			sort.Strings(have)
			sort.Strings(want)

			if !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		}
	}

	testCases := []struct {
		name   string
		assert types.ReposAssertion
		conf   *schema.PagureConnection
		err    string
	}{
		{
			name: "pattern",
			assert: assertAllReposListed([]string{
				"src.fedoraproject.org/rpms/tmux",
			}),
			conf: &schema.PagureConnection{
				Url:     "https://src.fedoraproject.org",
				Pattern: "tmux",
			},
			err: "<nil>",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "PAGURE/" + tc.name
		t.Run(tc.name, func(t *testing.T) {
			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &types.ExternalService{
				Kind:   extsvc.KindPagure,
				Config: marshalJSON(t, tc.conf),
			}

			src, err := NewPagureSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			repos, err := listAll(context.Background(), src)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repos)
			}
		})
	}
}
