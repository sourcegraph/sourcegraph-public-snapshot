package server

import (
	"context"
	"testing"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"github.com/sourcegraph/sourcegraph/xlang/lspext"
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
			Name:           "customer-cmd-path",
			RootURI:        "git://github.com/alice/code",
			WantImportPath: "alice.org/code",
			WantGoPath:     gopath,
			FS: map[string]string{
				"kode/cmd/alice/alice.go": `package http // import "alice.org/code/kode/cmd/alice"`,
			},
		},
		{
			Name:           "cfg-too-long",
			RootURI:        "git://github.com/alice/code",
			WantImportPath: "good.org/code",
			WantGoPath:     gopath,
			FS: map[string]string{
				// Not picked up because it is one level too deep.
				"a/b/c/d/alice.go": `package http // import "alice.org/code/a/b/c/d"`,

				// But users can manually specify it via the config:
				".sourcegraph/config.json": `{"go": {"RootImportPath": "good.org/code"}}`,
			},
		},
		{
			Name:           "cfg-invalid-detection",
			RootURI:        "git://github.com/alice/code",
			WantImportPath: "alice.org/code",
			WantGoPath:     gopath,
			FS: map[string]string{
				// Pretend this code is *not* their actual code, they just have
				// someone else's code copied into their project (e.g. very poor
				// vendoring solution). They do *not* want us to use this
				// canonical import path.
				"kode/cmd/alice/alice.go": `package http // import "bad.org/code/kode/cmd/alice"`,

				// They can override it via the config:
				".sourcegraph/config.json": `{"go": {"RootImportPath": "alice.org/code"}}`,
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
				".vscode/settings.json": `{
// this JSON document has comments!
  "go.gopath": "${workspaceRoot}/third_party:${workspaceRoot}/code",
}`,
			},
		},

		{
			Name:           "monorepo_envrc",
			RootURI:        "git://github.com/janet/monorepo",
			WantImportPath: "",
			WantGoPath:     "/workspace/third_party:/workspace/third_party2:/workspace/code:/workspace/code2:/workspace/included/intentionally:/",
			FS: map[string]string{
				".envrc": `junk
unparsable
export GOPATH=${PWD}/third_party:$(PWD)/third_party2
GOPATH_add code:code2
GOPATH_add /absolute
` + "export GOPATH=   \"`pwd`included/intentionally\"" + `
123\more/junk
`,
			},
		},

		{
			Name:           "monorepo_sourcegraph_config",
			RootURI:        "git://github.com/kim/monorepo",
			WantImportPath: "",
			WantGoPath:     "/workspace/third_party:/workspace/code:/",
			FS: map[string]string{
				".sourcegraph/config.json": `{"go": {"GOPATH": ["/third_party", "code"]}}`,
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

// mapFS lets us easily instantiate a VFS with a map[string]string
// (which is less noisy than map[string][]byte in test fixtures).
func mapFS(m map[string]string) ctxvfs.FileSystem {
	m2 := make(map[string][]byte, len(m))
	for k, v := range m {
		m2[k] = []byte(v)
	}
	return ctxvfs.Map(m2)
}
