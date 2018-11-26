package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	lsext "github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/go-lsp/lspext"
	"github.com/sourcegraph/jsonrpc2"
	gobuildserver "github.com/sourcegraph/sourcegraph/enterprise/cmd/xlang-go/internal/server"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
)

func TestProxy(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tests := map[string]struct {
		rootURI           lsp.DocumentURI
		mode              string
		fs                map[string]string
		wantHover         map[string]string
		wantDefinition    map[string]string
		wantXDefinition   map[string]string
		wantReferences    map[string][]string
		wantSymbols       map[string][]string
		wantXDependencies string
		wantXReferences   map[*lsext.WorkspaceReferencesParams][]string
		wantXPackages     []string
		depFS             map[string]map[string]string // dep clone URL -> map VFS
	}{
		"go basic": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": "package p; func A() { A() }",
				"b.go": "package p; func B() { A() }",
			},
			wantHover: map[string]string{
				"a.go:1:9":  "package p",
				"a.go:1:17": "func A()",
				"a.go:1:23": "func A()",
				"b.go:1:17": "func B()",
				"b.go:1:23": "func A()",
			},
			wantDefinition: map[string]string{
				"a.go:1:17": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17",
				"a.go:1:23": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17",
				"b.go:1:17": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b.go:1:17",
				"b.go:1:23": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17",
			},
			wantXDefinition: map[string]string{
				"a.go:1:17": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17 id:test/pkg/-/A name:A package:test/pkg packageName:p recv: vendor:false",
				"a.go:1:23": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17 id:test/pkg/-/A name:A package:test/pkg packageName:p recv: vendor:false",
				"b.go:1:17": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b.go:1:17 id:test/pkg/-/B name:B package:test/pkg packageName:p recv: vendor:false",
				"b.go:1:23": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17 id:test/pkg/-/A name:A package:test/pkg packageName:p recv: vendor:false",
			},
			wantReferences: map[string][]string{
				"a.go:1:17": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:23",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b.go:1:23",
				},
				"a.go:1:23": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:23",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b.go:1:23",
				},
				"b.go:1:17": []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b.go:1:17"},
				"b.go:1:23": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:23",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b.go:1:23",
				},
			},
			wantSymbols: map[string][]string{
				"":            []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:function:A:0:16", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b.go:function:B:0:16"},
				"A":           []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:function:A:0:16"},
				"B":           []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b.go:function:B:0:16"},
				"is:exported": []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:function:A:0:16", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b.go:function:B:0:16"},
			},
			wantXPackages: []string{"test/pkg"},
		},
		"go detailed": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": "package p; type T struct { F string }",
			},
			// "a.go:1:28": "(T).F string", // TODO(sqs): see golang/hover.go; this is the output we want
			wantHover: map[string]string{},
			wantSymbols: map[string][]string{
				"":            []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:class:T:0:16", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:field:T.F:0:27"},
				"T":           []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:class:T:0:16", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:field:T.F:0:27"},
				"F":           []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:field:T.F:0:27"},
				"is:exported": []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:class:T:0:16", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:field:T.F:0:27"},
			},
		},
		"exported defs unexported type": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": "package p; type t struct { F string }",
			},
			wantSymbols: map[string][]string{
				"is:exported": []string{},
			},
		},
		"go xtest": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go":      "package p; var A int",
				"a_test.go": `package p_test; import "test/pkg"; var X = p.A`,
			},
			wantHover: map[string]string{
				"a.go:1:16":      "A int",
				"a_test.go:1:40": "X int",
				"a_test.go:1:46": "A int",
			},
		},
		"go subdirectory in repo": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d",
			mode:    "go",
			fs: map[string]string{
				"a.go":    "package d; func A() { A() }",
				"d2/b.go": `package d2; import "test/pkg/d"; func B() { d.A(); B() }`,
			},
			wantHover: map[string]string{
				"a.go:1:17":    "func A()",
				"a.go:1:23":    "func A()",
				"d2/b.go:1:39": "func B()",
				"d2/b.go:1:47": "func A()",
				"d2/b.go:1:52": "func B()",
			},
			wantDefinition: map[string]string{
				"a.go:1:17":    "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/a.go:1:17",
				"a.go:1:23":    "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/a.go:1:17",
				"d2/b.go:1:39": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:39",
				"d2/b.go:1:47": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/a.go:1:17",
				"d2/b.go:1:52": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:39",
			},
			wantXDefinition: map[string]string{
				"a.go:1:17":    "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/a.go:1:17 id:test/pkg/d/-/A name:A package:test/pkg/d packageName:d recv: vendor:false",
				"a.go:1:23":    "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/a.go:1:17 id:test/pkg/d/-/A name:A package:test/pkg/d packageName:d recv: vendor:false",
				"d2/b.go:1:39": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:39 id:test/pkg/d/d2/-/B name:B package:test/pkg/d/d2 packageName:d2 recv: vendor:false",
				"d2/b.go:1:47": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/a.go:1:17 id:test/pkg/d/-/A name:A package:test/pkg/d packageName:d recv: vendor:false",
				"d2/b.go:1:52": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:39 id:test/pkg/d/d2/-/B name:B package:test/pkg/d/d2 packageName:d2 recv: vendor:false",
			},
			wantSymbols: map[string][]string{
				"":            []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/a.go:function:A:0:16", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:function:B:0:38"},
				"is:exported": []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/a.go:function:A:0:16", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:function:B:0:38"},
			},
			wantXReferences: map[*lsext.WorkspaceReferencesParams][]string{
				// Non-matching name query.
				{Query: lsext.SymbolDescriptor{"name": "nope"}}: []string{},

				// Matching against invalid field name.
				{Query: lsext.SymbolDescriptor{"nope": "A"}}: []string{},

				// Matching against an invalid dirs hint.
				{Query: lsext.SymbolDescriptor{"package": "test/pkg/d"}, Hints: map[string]interface{}{"dirs": []string{"file:///src/test/pkg/d/d3"}}}: []string{},

				// Matching against a dirs hint with multiple dirs.
				{Query: lsext.SymbolDescriptor{"package": "test/pkg/d"}, Hints: map[string]interface{}{"dirs": []string{"file:///d2", "file:///invalid"}}}: []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:20-1:32 -> id:test/pkg/d name: package:test/pkg/d packageName:d recv: vendor:false",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:47-1:48 -> id:test/pkg/d/-/A name:A package:test/pkg/d packageName:d recv: vendor:false",
				},

				// Matching against a dirs hint.
				{Query: lsext.SymbolDescriptor{"package": "test/pkg/d"}, Hints: map[string]interface{}{"dirs": []string{"file:///d2"}}}: []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:20-1:32 -> id:test/pkg/d name: package:test/pkg/d packageName:d recv: vendor:false",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:47-1:48 -> id:test/pkg/d/-/A name:A package:test/pkg/d packageName:d recv: vendor:false",
				},

				// Matching against single field.
				{Query: lsext.SymbolDescriptor{"package": "test/pkg/d"}}: []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:20-1:32 -> id:test/pkg/d name: package:test/pkg/d packageName:d recv: vendor:false",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:47-1:48 -> id:test/pkg/d/-/A name:A package:test/pkg/d packageName:d recv: vendor:false",
				},

				// Matching against no fields.
				{Query: lsext.SymbolDescriptor{}}: []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:20-1:32 -> id:test/pkg/d name: package:test/pkg/d packageName:d recv: vendor:false",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:47-1:48 -> id:test/pkg/d/-/A name:A package:test/pkg/d packageName:d recv: vendor:false",
				},
				{
					Query: lsext.SymbolDescriptor{
						"name":        "",
						"package":     "test/pkg/d",
						"packageName": "d",
						"recv":        "",
						"vendor":      false,
					},
				}: []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:20-1:32 -> id:test/pkg/d name: package:test/pkg/d packageName:d recv: vendor:false"},
				{
					Query: lsext.SymbolDescriptor{
						"name":        "A",
						"package":     "test/pkg/d",
						"packageName": "d",
						"recv":        "",
						"vendor":      false,
					},
				}: []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#d/d2/b.go:1:47-1:48 -> id:test/pkg/d/-/A name:A package:test/pkg/d packageName:d recv: vendor:false"},
			},
			wantXPackages: []string{"test/pkg/d", "test/pkg/d/d2"},
		},
		"go multiple packages in dir": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": "package p; func A() { A() }",
				"main.go": `// +build ignore

package main; import "test/pkg"; func B() { p.A(); B() }`,
			},
			wantHover: map[string]string{
				"a.go:1:17": "func A()",
				"a.go:1:23": "func A()",
				// Not parsing build-tag-ignored files:
				//
				// "main.go:3:39": "func B()", // func B()
				// "main.go:3:47": "func A()", // p.A()
				// "main.go:3:52": "func B()", // B()
			},
			wantDefinition: map[string]string{
				"a.go:1:17": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17",
				"a.go:1:23": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17",
				// Not parsing build-tag-ignored files:
				//
				// "main.go:3:39": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#main.go:3:39", // B() -> func B()
				// "main.go:3:47": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17",    // p.A() -> a.go func A()
				// "main.go:3:52": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#main.go:3:39", // B() -> func B()
			},
			wantXDefinition: map[string]string{
				"a.go:1:17": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17 id:test/pkg/-/A name:A package:test/pkg packageName:p recv: vendor:false",
				"a.go:1:23": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:17 id:test/pkg/-/A name:A package:test/pkg packageName:p recv: vendor:false",
			},
			wantSymbols: map[string][]string{
				"":            []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:function:A:0:16"},
				"is:exported": []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:function:A:0:16"},
			},
			wantXPackages: []string{"test/pkg"},
		},
		"goroot": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": `package p; import "fmt"; var _ = fmt.Println; var x int`,
			},
			wantHover: map[string]string{
				"a.go:1:40": "func Println(a ...interface{}) (n int, err error)",
				// "a.go:1:53": "type int int",
			},
			wantDefinition: map[string]string{
				"a.go:1:40": "git://github.com/golang/go?go1.7.1#src/fmt/print.go:1:19",
				// "a.go:1:53": "git://github.com/golang/go?go1.7.1#src/builtin/builtin.go:TODO:TODO", // TODO(sqs): support builtins
			},
			wantXDefinition: map[string]string{
				"a.go:1:40": "git://github.com/golang/go?go1.7.1#src/fmt/print.go:1:19 id:fmt/-/Println name:Println package:fmt packageName:fmt recv: vendor:false",
			},
			depFS: map[string]map[string]string{
				"https://github.com/golang/go?go1.7.1": {
					"src/fmt/print.go":       "package fmt; func Println(a ...interface{}) (n int, err error) { return }",
					"src/builtin/builtin.go": "package builtin; type int int",
				},
			},
			wantSymbols: map[string][]string{
				"":            []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:variable:x:0:50"},
				"is:exported": []string{},
			},
		},
		"gopath": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a/a.go": `package a; func A() {}`,
				"b/b.go": `package b; import "test/pkg/a"; var _ = a.A`,
			},
			wantHover: map[string]string{
				"a/a.go:1:17": "func A()",
				// "b/b.go:1:20": "package", // TODO(sqs): make import paths hoverable
				"b/b.go:1:43": "func A()",
			},
			wantDefinition: map[string]string{
				"a/a.go:1:17": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:1:17",
				// "b/b.go:1:20": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a", // TODO(sqs): make import paths hoverable
				"b/b.go:1:43": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:1:17",
			},
			wantXDefinition: map[string]string{
				"a/a.go:1:17": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:1:17 id:test/pkg/a/-/A name:A package:test/pkg/a packageName:a recv: vendor:false",
				"b/b.go:1:43": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:1:17 id:test/pkg/a/-/A name:A package:test/pkg/a packageName:a recv: vendor:false",
			},
			wantReferences: map[string][]string{
				"a/a.go:1:17": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:1:17",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b/b.go:1:43",
				},
				"b/b.go:1:43": []string{ // calling "references" on call site should return same result as on decl
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:1:17",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b/b.go:1:43",
				},
			},
			wantSymbols: map[string][]string{
				"":            []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:function:A:0:16"},
				"is:exported": []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:function:A:0:16"},
			},
		},
		"go vendored dep": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go":                              `package a; import "github.com/v/vendored"; var _ = vendored.V`,
				"vendor/github.com/v/vendored/v.go": "package vendored; func V() {}",
			},
			wantHover: map[string]string{
				"a.go:1:61": "func V()",
			},
			wantDefinition: map[string]string{
				"a.go:1:61": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/v/vendored/v.go:1:24",
			},
			wantXDefinition: map[string]string{
				"a.go:1:61": "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/v/vendored/v.go:1:24 id:test/pkg/vendor/github.com/v/vendored/-/V name:V package:test/pkg/vendor/github.com/v/vendored packageName:vendored recv: vendor:true",
			},
			wantReferences: map[string][]string{
				"vendor/github.com/v/vendored/v.go:1:24": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/v/vendored/v.go:1:24",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:61",
				},
			},
			wantSymbols: map[string][]string{
				"":            []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/v/vendored/v.go:function:V:0:23"},
				"is:exported": []string{},
			},
			wantXPackages: []string{"test/pkg", "test/pkg/vendor/github.com/v/vendored"},
		},
		"go vendor symbols with same name": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"z.go":                          `package pkg; func x() bool { return true }`,
				"vendor/github.com/a/pkg2/x.go": `package pkg2; func x() bool { return true }`,
				"vendor/github.com/x/pkg3/x.go": `package pkg3; func x() bool { return true }`,
			},
			wantSymbols: map[string][]string{
				"": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#z.go:function:x:0:18",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/a/pkg2/x.go:function:x:0:19",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/x/pkg3/x.go:function:x:0:19",
				},
				"x": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#z.go:function:x:0:18",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/a/pkg2/x.go:function:x:0:19",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/x/pkg3/x.go:function:x:0:19",
				},
				"pkg2.x": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#z.go:function:x:0:18",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/a/pkg2/x.go:function:x:0:19",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/x/pkg3/x.go:function:x:0:19",
				},
				"pkg3.x": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#z.go:function:x:0:18",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/x/pkg3/x.go:function:x:0:19",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#vendor/github.com/a/pkg2/x.go:function:x:0:19",
				},
				"is:exported": []string{},
			},
		},
		"go external dep": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": `package a; import "github.com/d/dep"; var _ = dep.D; var _ = dep.D`,
			},
			wantHover: map[string]string{
				"a.go:1:51": "func D()",
			},
			wantDefinition: map[string]string{
				"a.go:1:51": "git://github.com/d/dep?HEAD#d.go:1:19",
			},
			wantXDefinition: map[string]string{
				"a.go:1:51": "git://github.com/d/dep?HEAD#d.go:1:19 id:github.com/d/dep/-/D name:D package:github.com/d/dep packageName:dep recv: vendor:false",
			},
			wantReferences: map[string][]string{
				"a.go:1:51": []string{
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:51",
					"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a.go:1:66",
					// Do not include "refs" from the dependency
					// package itself; only return results in the
					// workspace.
				},
			},
			depFS: map[string]map[string]string{
				"https://github.com/d/dep?HEAD": {
					"d.go": "package dep; func D() {}; var _ = D",
				},
			},
		},
		"external dep with vendor": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": `package p; import "github.com/d/dep"; var _ = dep.D().F`,
			},
			wantDefinition: map[string]string{
				"a.go:1:55": "git://github.com/d/dep?HEAD#vendor/vendp/vp.go:1:32",
			},
			wantXDefinition: map[string]string{
				"a.go:1:55": "git://github.com/d/dep?HEAD#vendor/vendp/vp.go:1:32 id:github.com/d/dep/vendor/vendp/-/V/F name:F package:github.com/d/dep/vendor/vendp packageName:vendp recv:V vendor:true",
			},
			depFS: map[string]map[string]string{
				"https://github.com/d/dep?HEAD": map[string]string{
					"d.go":               `package dep; import "vendp"; func D() (v vendp.V) { return }`,
					"vendor/vendp/vp.go": "package vendp; type V struct { F int }",
				},
			},
		},
		"go external dep at subtree": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": `package a; import "github.com/d/dep/subp"; var _ = subp.D`,
			},
			wantHover: map[string]string{
				"a.go:1:57": "func D()",
			},
			wantDefinition: map[string]string{
				"a.go:1:57": "git://github.com/d/dep?HEAD#subp/d.go:1:20",
			},
			wantXDefinition: map[string]string{
				"a.go:1:57": "git://github.com/d/dep?HEAD#subp/d.go:1:20 id:github.com/d/dep/subp/-/D name:D package:github.com/d/dep/subp packageName:subp recv: vendor:false",
			},
			depFS: map[string]map[string]string{
				"https://github.com/d/dep?HEAD": {
					"subp/d.go": "package subp; func D() {}",
				},
			},
		},
		"go nested external dep": { // a depends on dep1, dep1 depends on dep2
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": `package a; import "github.com/d/dep1"; var _ = dep1.D1().D2`,
			},
			wantHover: map[string]string{
				"a.go:1:53": "func D1() D2",
				"a.go:1:59": "D2 int",
			},
			wantDefinition: map[string]string{
				"a.go:1:53": "git://github.com/d/dep1?HEAD#d1.go:1:48", // func D1
				"a.go:1:58": "git://github.com/d/dep2?HEAD#d2.go:1:32", // field D2
			},
			wantXDefinition: map[string]string{
				"a.go:1:53": "git://github.com/d/dep1?HEAD#d1.go:1:48 id:github.com/d/dep1/-/D1 name:D1 package:github.com/d/dep1 packageName:dep1 recv: vendor:false",
				"a.go:1:58": "git://github.com/d/dep2?HEAD#d2.go:1:32 id:github.com/d/dep2/-/D2/D2 name:D2 package:github.com/d/dep2 packageName:dep2 recv:D2 vendor:false",
			},
			depFS: map[string]map[string]string{
				"https://github.com/d/dep1?HEAD": {
					"d1.go": `package dep1; import "github.com/d/dep2"; func D1() dep2.D2 { return dep2.D2{} }`,
				},
				"https://github.com/d/dep2?HEAD": {
					"d2.go": "package dep2; type D2 struct { D2 int }",
				},
			},
		},
		"go external dep at vanity import path": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a.go": `package a; import "golang.org/x/text"; var _ = text.F`,
			},
			wantHover: map[string]string{
				"a.go:1:53": "func F()",
			},
			wantDefinition: map[string]string{
				"a.go:1:53": "git://github.com/golang/text?HEAD#dummy.go:1:20",
			},
			wantXDefinition: map[string]string{
				"a.go:1:53": "git://github.com/golang/text?HEAD#dummy.go:1:20 id:golang.org/x/text/-/F name:F package:golang.org/x/text packageName:text recv: vendor:false",
			},
			depFS: map[string]map[string]string{
				// We override the Git cloning of this repo to use
				// in-memory dummy data, but we still need to hit the
				// network to resolve the Go custom import path
				// (because that's not mocked yet).
				"https://github.com/golang/text?HEAD": {
					"dummy.go": "package text; func F() {}",
				},
			},
		},

		// This covers repos like github.com/kubernetes/kubernetes,
		// which have doc.go files in subpackages with canonical
		// import path comments of "//
		// k8s.io/kubernetes/SUBPACKAGE...". If we don't set up the
		// workspace at /src/k8s.io/kubernetes, then cross-package
		// definitions will fail, and we will erroneously fetch a
		// separate (HEAD) copy of the entire kubernetes repo at the
		// k8s.io/kubernetes/... root.
		"go packages with canonical import path different from its repo": {
			rootURI: "git://test/foo?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"a/a.go": `package a // import "other/foo/a"

import "other/foo/b"

var A = b.B`,
				"b/b.go": `package b // import "other/foo/b"

var (
	B = 123
	bb = B
)`,
			},
			wantDefinition: map[string]string{
				"a/a.go:5:5":  "git://test/foo?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:5:5", // "var A"
				"a/a.go:5:11": "git://test/foo?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b/b.go:4:2", // "b.B"
				"b/b.go:4:2":  "git://test/foo?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b/b.go:4:2", // "B = 123"
				"b/b.go:5:7":  "git://test/foo?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b/b.go:4:2", // "bb = B"
			},
			wantXDefinition: map[string]string{
				"a/a.go:5:5":  "git://test/foo?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#a/a.go:5:5 id:other/foo/a/-/A name:A package:other/foo/a packageName:a recv: vendor:false",
				"a/a.go:5:11": "git://test/foo?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b/b.go:4:2 id:other/foo/b/-/B name:B package:other/foo/b packageName:b recv: vendor:false",
				"b/b.go:4:2":  "git://test/foo?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b/b.go:4:2 id:other/foo/b/-/B name:B package:other/foo/b packageName:b recv: vendor:false",
				"b/b.go:5:7":  "git://test/foo?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#b/b.go:4:2 id:other/foo/b/-/B name:B package:other/foo/b packageName:b recv: vendor:false",
			},
		},

		"go symbols": {
			rootURI: "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			mode:    "go",
			fs: map[string]string{
				"abc.go": `package a

type XYZ struct {}

func (x XYZ) ABC() {}
`,
				"bcd.go": `package a

type YZA struct {}

func (y YZA) BCD() {}
`,
				"xyz.go": `package a

func yza() {}
`,
			},
			wantSymbols: map[string][]string{
				"":            []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#abc.go:class:XYZ:2:5", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#bcd.go:class:YZA:2:5", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#xyz.go:function:yza:2:5", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#abc.go:method:XYZ.ABC:4:13", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#bcd.go:method:YZA.BCD:4:13"},
				"xyz":         []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#abc.go:class:XYZ:2:5", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#abc.go:method:XYZ.ABC:4:13", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#xyz.go:function:yza:2:5"},
				"yza":         []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#bcd.go:class:YZA:2:5", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#xyz.go:function:yza:2:5", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#bcd.go:method:YZA.BCD:4:13"},
				"abc":         []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#abc.go:method:XYZ.ABC:4:13", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#abc.go:class:XYZ:2:5"},
				"bcd":         []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#bcd.go:method:YZA.BCD:4:13", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#bcd.go:class:YZA:2:5"},
				"is:exported": []string{"git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#abc.go:class:XYZ:2:5", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#bcd.go:class:YZA:2:5", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#abc.go:method:XYZ.ABC:4:13", "git://test/pkg?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#bcd.go:method:YZA.BCD:4:13"},
			},
		},
	}
	for label, test := range tests {
		t.Run(label, func(t *testing.T) {
			// Mock repo and dep fetching to use test fixtures.
			{
				orig := gobuildserver.NewDepRepoVFS
				gobuildserver.NewDepRepoVFS = func(ctx context.Context, cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
					id := cloneURL.String() + "?" + rev
					if fs, ok := test.depFS[id]; ok {
						return mapFS(fs), nil
					}
					return nil, fmt.Errorf("no file system found for dep at %s rev %q", cloneURL, rev)
				}
				defer func() {
					gobuildserver.NewDepRepoVFS = orig
				}()

				origRemoteFS := gobuildserver.RemoteFS
				gobuildserver.RemoteFS = func(ctx context.Context, initializeParams lspext.InitializeParams) (ctxvfs.FileSystem, error) {
					return mapFS(test.fs), nil
				}

				defer func() {
					gobuildserver.RemoteFS = origRemoteFS
				}()
			}

			ctx := context.Background()
			if test.rootURI == "" {
				t.Fatal("no rootPath set in test fixture")
			}

			root, err := gituri.Parse(string(test.rootURI))
			if err != nil {
				t.Fatal(err)
			}

			c, done := connectionToNewBuildServer(string(test.rootURI), t)
			defer done()

			// Prepare the connection.
			if err := c.Call(ctx, "initialize", lspext.InitializeParams{
				InitializeParams: lsp.InitializeParams{RootURI: "file:///"},
				OriginalRootURI:  test.rootURI,
			}, nil); err != nil {
				t.Fatal("initialize:", err)
			}

			lspTests(t, ctx, c, root, test.wantHover, test.wantDefinition, test.wantXDefinition, test.wantReferences, test.wantSymbols, test.wantXDependencies, test.wantXReferences, test.wantXPackages)
		})
	}
}

// InMemoryPeerConns is a convenience helper that returns a pair of
// io.ReadWriteClosers that are each other's peer.
//
// It can be used, for example, to run an in-memory JSON-RPC handler
// that speaks to an in-memory client, without needin to open a Unix
// or TCP connection.
//
// Copied from xlang/proxy/servers.go, which will get deleted soon.
func InMemoryPeerConns() (jsonrpc2.ObjectStream, jsonrpc2.ObjectStream) {
	sr, cw := io.Pipe()
	cr, sw := io.Pipe()
	return jsonrpc2.NewBufferedStream(&pipeReadWriteCloser{sr, sw}, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.NewBufferedStream(&pipeReadWriteCloser{cr, cw}, jsonrpc2.VSCodeObjectCodec{})
}

type pipeReadWriteCloser struct {
	*io.PipeReader
	*io.PipeWriter
}

func (c *pipeReadWriteCloser) Close() error {
	err1 := c.PipeReader.Close()
	err2 := c.PipeWriter.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func connectionToNewBuildServer(root string, t testing.TB) (*jsonrpc2.Conn, func()) {
	rootURI, err := gituri.Parse(root)
	if err != nil {
		t.Fatal(err)
	}
	// Run in-process for easy development (no recompiles, etc.).
	a, b := InMemoryPeerConns()

	convertURIs := func(m *json.RawMessage, f func(root gituri.URI, uriStr string) (*gituri.URI, error)) error {
		var obj interface{}
		if err := json.Unmarshal(*m, &obj); err != nil {
			return err
		}
		var walkErr error
		lspext.WalkURIFields(obj, nil, func(uriStr lsp.DocumentURI) lsp.DocumentURI {
			newURI, err := f(*rootURI, string(uriStr))
			if err != nil {
				walkErr = err
				return ""
			}
			return lsp.DocumentURI(newURI.String())
		})
		if walkErr != nil {
			return walkErr
		}
		r, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		*m = json.RawMessage(r)
		return nil
	}

	onSend := func(req *jsonrpc2.Request, res *jsonrpc2.Response) {
		if res == nil {
			err := convertURIs(req.Params, RelWorkspaceURI)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	onRecv := func(req *jsonrpc2.Request, res *jsonrpc2.Response) {
		if res != nil && res.Result != nil {
			err := convertURIs(res.Result, AbsWorkspaceURI)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	jsonrpc2.NewConn(context.Background(), a, jsonrpc2.AsyncHandler(gobuildserver.NewHandler()))

	conn := jsonrpc2.NewConn(context.Background(), b, NoopHandler{}, jsonrpc2.OnRecv(onRecv), jsonrpc2.OnSend(onSend))
	done := func() {
		a.Close()
		b.Close()
	}
	return conn, done
}

type NoopHandler struct{}

func (NoopHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {}

// RelWorkspaceURI and AbsWorkspaceURI were copied from xlang/proxy/uris.go,
// which will get deleted soon.

// RelWorkspaceURI maps absolute URIs like
// "git://github.com/facebook/react.git?master#dir/file.txt" to
// workspace-relative file URIs like "file:///dir/file.txt". The
// result is a path within the workspace's virtual file system that
// will contain the original path's contents.
//
// If uriStr isn't underneath root, the parsed original uriStr is
// returned. This occurs when a cross-workspace resource is referenced
// (e.g., a client performs a cross-workspace go-to-definition and
// then notifies the server that the client opened the destination
// resource).
func RelWorkspaceURI(root gituri.URI, uriStr string) (*gituri.URI, error) {
	u, err := gituri.Parse(uriStr)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(uriStr, root.String()) {
		// The uriStr resource lives in a different workspace.
		return u, nil
	}
	if p := path.Clean(u.FilePath()); strings.HasPrefix(p, "/") || strings.HasPrefix(p, "..") {
		return nil, fmt.Errorf("invalid file path in URI %q in LSP proxy client request (must not begin with '/', '..', or contain '.' or '..' components)", uriStr)
	} else if u.FilePath() != "" && p != u.FilePath() {
		return nil, fmt.Errorf("invalid file path in URI %q (raw file path %q != cleaned file path %q)", uriStr, u.FilePath(), p)
	}

	// Support when root is rooted at a subdir.
	if rootPath := root.FilePath(); rootPath != "" {
		rootPath = strings.TrimSuffix(rootPath, string(os.PathSeparator))
		if !strings.HasPrefix(u.FilePath(), rootPath+string(os.PathSeparator)) {
			return u, nil
		}
		u = u.WithFilePath(strings.TrimPrefix(u.FilePath(), rootPath+string(os.PathSeparator)))
	}

	return &gituri.URI{URL: url.URL{Scheme: "file", Path: "/" + u.FilePath()}}, nil
}

// AbsWorkspaceURI is the inverse of relWorkspaceURI. It maps
// workspace-relative URIs like "file:///dir/file.txt" to their
// absolute URIs like
// "git://github.com/facebook/react.git?master#dir/file.txt".
func AbsWorkspaceURI(root gituri.URI, uriStr string) (*gituri.URI, error) {
	uri, err := gituri.Parse(uriStr)
	if err != nil {
		return nil, err
	}
	if uri.Scheme == "file" {
		return root.WithFilePath(root.ResolveFilePath(uri.Path)), nil
	}
	return uri, nil
	// Another possibility is a "git://" URI that the build/lang
	// server knew enough to produce on its own (e.g., to refer to
	// git://github.com/golang/go for a Go stdlib definition). No need
	// to rewrite those.
}
