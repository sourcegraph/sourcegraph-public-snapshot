package backend

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReposSubset(t *testing.T) {
	var indexed map[string][]string
	index := &Indexers{
		Map: prefixMap([]string{"foo", "bar", "baz.fully.qualified:80"}),
		Indexed: func(ctx context.Context, k string) map[string]struct{} {
			set := map[string]struct{}{}
			if indexed == nil {
				return set
			}
			for _, s := range indexed[k] {
				set[s] = struct{}{}
			}
			return set
		},
	}

	cases := []struct {
		name     string
		hostname string
		indexed  map[string][]string
		repos    []string
		want     []string
		errS     string
	}{{
		name:     "bad hostname",
		hostname: "bam",
		errS:     "hostname \"bam\" not found",
	}, {
		name:     "all",
		hostname: "foo",
		repos:    []string{"foo-1", "foo-2", "foo-3"},
		want:     []string{"foo-1", "foo-2", "foo-3"},
	}, {
		name:     "none",
		hostname: "bar",
		repos:    []string{"foo-1", "foo-2", "foo-3"},
		want:     []string{},
	}, {
		name:     "subset",
		hostname: "foo",
		repos:    []string{"foo-2", "bar-1", "foo-1", "foo-3"},
		want:     []string{"foo-2", "foo-1", "foo-3"},
	}, {
		name:     "qualified",
		hostname: "baz.fully.qualified",
		repos:    []string{"baz.fully.qualified:80-1", "baz.fully.qualified:80-2", "foo-1"},
		want:     []string{"baz.fully.qualified:80-1", "baz.fully.qualified:80-2"},
	}, {
		name:     "unqualified",
		hostname: "baz",
		repos:    []string{"baz.fully.qualified:80-1", "baz.fully.qualified:80-2", "foo-1"},
		want:     []string{"baz.fully.qualified:80-1", "baz.fully.qualified:80-2"},
	}, {
		name:     "drop",
		hostname: "foo",
		indexed: map[string][]string{
			"foo": {"foo-1", "foo-drop", "bar-drop", "bar-keep"},
			"bar": {"foo-1", "bar-drop"},
		},
		repos: []string{"foo-1", "foo-2", "foo-3", "bar-drop", "bar-keep"},
		want:  []string{"foo-1", "foo-2", "foo-3", "bar-keep"},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			indexed = tc.indexed
			got, err := index.ReposSubset(ctx, tc.hostname, index.Indexed(ctx, tc.hostname), tc.repos)
			if tc.errS != "" {
				got := fmt.Sprintf("%v", err)
				if !strings.Contains(got, tc.errS) {
					t.Fatalf("error %q does not contain %q", got, tc.errS)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("reposSubset mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestFindEndpoint(t *testing.T) {
	cases := []struct {
		name      string
		hostname  string
		endpoints []string
		want      string
		errS      string
	}{{
		name:      "empty",
		hostname:  "",
		endpoints: []string{"foo.internal", "bar.internal"},
		errS:      "hostname \"\" not found",
	}, {
		name:      "empty endpoints",
		hostname:  "foo",
		endpoints: []string{},
		errS:      "hostname \"foo\" not found",
	}, {
		name:      "bad prefix",
		hostname:  "foo",
		endpoints: []string{"foobar", "barfoo"},
		errS:      "hostname \"foo\" not found",
	}, {
		name:      "bad port",
		hostname:  "foo",
		endpoints: []string{"foo:80", "foo.internal"},
		errS:      "hostname \"foo\" matches multiple",
	}, {
		name:      "multiple",
		hostname:  "foo",
		endpoints: []string{"foo.internal", "foo.external"},
		errS:      "hostname \"foo\" matches multiple",
	}, {
		name:      "exact multiple",
		hostname:  "foo",
		endpoints: []string{"foo", "foo.internal"},
		errS:      "hostname \"foo\" matches multiple",
	}, {
		name:      "exact",
		hostname:  "foo",
		endpoints: []string{"foo", "bar"},
		want:      "foo",
	}, {
		name:      "prefix",
		hostname:  "foo",
		endpoints: []string{"foo.internal", "bar.internal"},
		want:      "foo.internal",
	}, {
		name:      "prefix with bad",
		hostname:  "foo",
		endpoints: []string{"foo.internal", "foobar.internal"},
		want:      "foo.internal",
	}, {
		name:      "port prefix",
		hostname:  "foo",
		endpoints: []string{"foo.internal:80", "bar.internal:80"},
		want:      "foo.internal:80",
	}, {
		name:      "port exact",
		hostname:  "foo.internal",
		endpoints: []string{"foo.internal:80", "bar.internal:80"},
		want:      "foo.internal:80",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eps := map[string]struct{}{}
			for _, ep := range tc.endpoints {
				eps[ep] = struct{}{}
			}

			got, err := findEndpoint(eps, tc.hostname)
			if tc.errS != "" {
				got := fmt.Sprintf("%v", err)
				if !strings.Contains(got, tc.errS) {
					t.Fatalf("error %q does not contain %q", got, tc.errS)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if tc.want != got {
				t.Errorf("findEndpoint got %q, want %q", got, tc.want)
			}
		})
	}
}

// prefixMap assigns keys to values if the value is a prefix of key.
type prefixMap []string

func (m prefixMap) Endpoints() (map[string]struct{}, error) {
	eps := map[string]struct{}{}
	for _, v := range m {
		eps[v] = struct{}{}
	}
	return eps, nil
}

func (m prefixMap) GetMany(keys ...string) ([]string, error) {
	vs := make([]string, len(keys))
	for i, k := range keys {
		for _, v := range m {
			if strings.HasPrefix(k, v) {
				vs[i] = v
			}
		}
	}
	return vs, nil
}
