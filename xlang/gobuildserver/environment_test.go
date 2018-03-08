package gobuildserver

import (
	"context"
	"testing"

	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

func TestDetermineEnvironment(t *testing.T) {
	type tcase struct {
		Name           string
		RootURI        lsp.DocumentURI
		WantImportPath string
		WantGoPath     string
		FS             map[string]string
	}

	cases := []tcase{
		{
			Name:           "glide",
			RootURI:        "git://github.com/alice/pkg",
			WantImportPath: "alice.org/pkg",
			WantGoPath:     gopath,
			FS: map[string]string{
				"glide.yaml": "package: alice.org/pkg",
			},
		},
		{
			Name:           "canonical",
			RootURI:        "git://github.com/alice/pkg",
			WantImportPath: "alice.org/pkg",
			WantGoPath:     gopath,
			FS: map[string]string{
				"doc.go": `package pkg // import "alice.org/pkg"`,
			},
		},
		{
			Name:           "nested-canonical",
			RootURI:        "git://github.com/alice/pkg",
			WantImportPath: "alice.org/pkg",
			WantGoPath:     gopath,
			FS: map[string]string{
				"http/doc.go": `package http // import "alice.org/pkg/http"`,
			},
		},
		{
			Name:           "fallback",
			RootURI:        "git://github.com/alice/pkg",
			WantImportPath: "github.com/alice/pkg",
			WantGoPath:     gopath,
			FS: map[string]string{
				"doc.go": `package pkg`,
			},
		},

		{
			Name:           "monorepo",
			RootURI:        "git://github.com/alice/monorepo",
			WantImportPath: "",
			WantGoPath:     "/workspace/third_party:/workspace/code:/",
			FS: map[string]string{
				".vscode/settings.json": `{"go.gopath":"${workspaceRoot}/third_party:${workspaceRoot}/code"}`,
			},
		},

		{
			Name:           "monorepo_envrc",
			RootURI:        "git://github.com/janet/monorepo",
			WantImportPath: "",
			WantGoPath:     "/workspace/third_party:/workspace/code:/workspace/included/intentionally:/",
			FS: map[string]string{
				".envrc": `junk
unparsable
export GOPATH=${PWD}/third_party
GOPATH_add code
GOPATH_add /absolute
` + "export GOPATH=   \"`pwd`included/intentionally\"" + `
123\more/junk
`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			params := lspext.InitializeParams{
				OriginalRootURI: tc.RootURI,
			}
			got, err := determineEnvironment(context.Background(), mapFS(tc.FS), params)
			if err != nil {
				t.Fatal("unexpected error", err)
			}
			if got.RootImportPath != tc.WantImportPath {
				t.Fatalf("got %q, want %q", got.RootImportPath, tc.WantImportPath)
			}
			if got.BuildContext.GOPATH != tc.WantGoPath {
				t.Fatalf("got %q, want %q", got.BuildContext.GOPATH, tc.WantGoPath)
			}
		})
	}
}
