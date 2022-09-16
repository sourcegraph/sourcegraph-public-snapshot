package repos

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSrcExpose(t *testing.T) {
	var body string
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/list-repos" {
			http.Error(w, r.URL.String()+" not found", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	defer s.Close()

	cases := []struct {
		name string
		body string
		want []*types.Repo
		err  string
	}{{
		name: "error",
		body: "boom",
		err:  "failed to decode response from src-expose: boom",
	}, {
		name: "nouri",
		body: `{"Items":[{"name": "foo"}]}`,
		err:  "repo without URI",
	}, {
		name: "empty",
		body: `{"items":[]}`,
		want: []*types.Repo{},
	}, {
		name: "minimal",
		body: `{"Items":[{"uri": "/repos/foo", "clonePath":"/repos/foo/.git"},{"uri":"/repos/bar/baz", "clonePath":"/repos/bar/baz/.git"}]}`,
		want: []*types.Repo{{
			URI:  "/repos/foo",
			Name: "/repos/foo",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: map[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metadata: &extsvc.OtherRepoMetadata{RelativePath: "/repos/foo/.git"},
		}, {
			URI:  "/repos/bar/baz",
			Name: "/repos/bar/baz",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/bar/baz",
			},
			Sources: map[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/bar/baz/.git",
				},
			},
			Metadata: &extsvc.OtherRepoMetadata{RelativePath: "/repos/bar/baz/.git"},
		}},
	}, {
		name: "override",
		body: `{"Items":[{"uri": "/repos/foo", "name": "foo", "description": "hi", "clonePath":"/repos/foo/.git"}]}`,
		want: []*types.Repo{{
			URI:         "/repos/foo",
			Name:        "foo",
			Description: "",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: map[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metadata: &extsvc.OtherRepoMetadata{RelativePath: "/repos/foo/.git"},
		}},
	}, {
		name: "immutable",
		body: `{"Items":[{"uri": "/repos/foo", "clonePath":"/repos/foo/.git", "enabled": false, "externalrepo": {"serviceid": "x", "servicetype": "y", "id": "z"}, "sources": {"x":{"id":"x", "cloneurl":"y"}}}]}`,
		want: []*types.Repo{{
			URI:  "/repos/foo",
			Name: "/repos/foo",
			ExternalRepo: api.ExternalRepoSpec{
				ServiceID:   s.URL,
				ServiceType: extsvc.TypeOther,
				ID:          "/repos/foo",
			},
			Sources: map[string]*types.SourceInfo{
				"extsvc:other:1": {
					ID:       "extsvc:other:1",
					CloneURL: s.URL + "/repos/foo/.git",
				},
			},
			Metadata: &extsvc.OtherRepoMetadata{RelativePath: "/repos/foo/.git"},
		}},
	}}

	ctx := context.Background()
	source, err := NewOtherSource(ctx, &types.ExternalService{
		ID:     1,
		Kind:   extsvc.KindOther,
		Config: extsvc.NewUnencryptedConfig(fmt.Sprintf(`{"url": %q}`, s.URL)),
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body = tc.body

			repos, err := source.srcExpose(context.Background())
			if got := fmt.Sprintf("%v", err); !strings.Contains(got, tc.err) {
				t.Fatalf("got error %v, want %v", got, tc.err)
			}
			if !reflect.DeepEqual(repos, tc.want) {
				t.Fatal("unexpected repos", cmp.Diff(tc.want, repos))
			}
		})
	}
}
