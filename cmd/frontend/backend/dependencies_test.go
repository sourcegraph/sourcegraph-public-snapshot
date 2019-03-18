package backend

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/go-lsp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/inventory"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

func TestDependencies_List(t *testing.T) {
	ctx := testContext()

	xlangDone := mockXLang(func(ctx context.Context, mode string, rootURI lsp.DocumentURI, method string, params, results interface{}) error {
		switch method {
		case "workspace/xdependencies":
			res, ok := results.(*[]lspext.DependencyReference)
			if !ok {
				t.Fatalf("attempted to call workspace/xpackages with invalid return type %T", results)
			}
			if rootURI != "git://r?aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
				t.Fatalf("unexpected rootURI: %q", rootURI)
			}
			switch mode {
			case "go_bg":
				*res = []lspext.DependencyReference{{
					Attributes: map[string]interface{}{
						"name":   "github.com/gorilla/dep",
						"vendor": true,
					},
				}}
			default:
				t.Fatalf("unexpected mode: %q", mode)
			}
		}
		return nil
	})
	defer xlangDone()

	Mocks.Repos.GetInventory = func(context.Context, *types.Repo, api.CommitID) (*inventory.Inventory, error) {
		return &inventory.Inventory{Languages: []*inventory.Lang{{Name: "Go"}}}, nil
	}

	repo := &types.Repo{ID: 1, URI: "r"}
	commitID := api.CommitID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	deps, err := Dependencies.List(ctx, repo, commitID, true)
	if err != nil {
		t.Fatal(err)
	}

	want := []*api.DependencyReference{{
		Language: "go",
		DepData:  map[string]interface{}{"name": "github.com/gorilla/dep", "vendor": true},
		RepoID:   repo.ID,
	}}
	if !reflect.DeepEqual(deps, want) {
		t.Errorf("got %+v, want %+v", deps, want)
	}
}
