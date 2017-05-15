package gobuildserver

import (
	"context"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/xlang/lspext"
)

func TestDetermineEnvironment(t *testing.T) {
	type tcase struct {
		Name           string
		RootPath       string
		WantImportPath string
		WantGoPath     string
		FS             map[string]string
	}

	cases := []tcase{
		{
			Name:           "glide",
			RootPath:       "git://github.com/alice/pkg",
			WantImportPath: "alice.org/pkg",
			WantGoPath:     gopath,
			FS: map[string]string{
				"glide.yaml": "package: alice.org/pkg",
			},
		},
		{
			Name:           "canonical",
			RootPath:       "git://github.com/alice/pkg",
			WantImportPath: "alice.org/pkg",
			WantGoPath:     gopath,
			FS: map[string]string{
				"doc.go": `package pkg // import "alice.org/pkg"`,
			},
		},
		{
			Name:           "nested-canonical",
			RootPath:       "git://github.com/alice/pkg",
			WantImportPath: "alice.org/pkg",
			WantGoPath:     gopath,
			FS: map[string]string{
				"http/doc.go": `package http // import "alice.org/pkg/http"`,
			},
		},
		{
			Name:           "fallback",
			RootPath:       "git://github.com/alice/pkg",
			WantImportPath: "github.com/alice/pkg",
			WantGoPath:     gopath,
			FS: map[string]string{
				"doc.go": `package pkg`,
			},
		},

		{
			Name:           "monorepo",
			RootPath:       "git://github.com/alice/monorepo",
			WantImportPath: "",
			WantGoPath:     "/workspace/third_party:/workspace/code:/",
			FS: map[string]string{
				".vscode/settings.json": `{"go.gopath":"${workspaceRoot}/third_party:${workspaceRoot}/code"}`,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			params := lspext.InitializeParams{
				OriginalRootPath: tc.RootPath,
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
