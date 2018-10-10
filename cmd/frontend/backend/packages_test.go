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

func TestPackages_List(t *testing.T) {
	ctx := testContext()

	xlangDone := mockXLang(func(ctx context.Context, mode string, rootURI lsp.DocumentURI, method string, params, results interface{}) error {
		switch method {
		case "workspace/xpackages":
			res, ok := results.(*[]lspext.PackageInformation)
			if !ok {
				t.Fatalf("attempted to call workspace/xpackages with invalid return type %T", results)
			}
			if rootURI != "git://r?aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
				t.Fatalf("unexpected rootURI: %q", rootURI)
			}
			switch mode {
			case "typescript":
				*res = []lspext.PackageInformation{{
					Package: map[string]interface{}{
						"name":    "tspkg",
						"version": "2.2.2",
					},
					Dependencies: []lspext.DependencyReference{},
				}}
			case "python":
				*res = []lspext.PackageInformation{{
					Package: map[string]interface{}{
						"name":    "pypkg",
						"version": "3.3.3",
					},
					Dependencies: []lspext.DependencyReference{},
				}}
			default:
				t.Fatalf("unexpected mode: %q", mode)
			}
		}
		return nil
	})
	defer xlangDone()

	Mocks.Repos.GetInventory = func(context.Context, *types.Repo, api.CommitID) (*inventory.Inventory, error) {
		return &inventory.Inventory{Languages: []*inventory.Lang{{Name: "TypeScript"}}}, nil
	}

	repo := &types.Repo{ID: 1, URI: "r"}
	commitID := api.CommitID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	pkgs, err := Packages.List(ctx, repo, commitID)
	if err != nil {
		t.Fatal(err)
	}

	want := []*api.PackageInfo{{
		RepoID: repo.ID,
		Lang:   "typescript",
		Pkg: map[string]interface{}{
			"name":    "tspkg",
			"version": "2.2.2",
		},
		Dependencies: []lspext.DependencyReference{},
	}}
	if !reflect.DeepEqual(pkgs, want) {
		t.Errorf("got %+v, want %+v", pkgs, want)
	}
}
