package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/ctxvfs"
	gobuildserver "github.com/sourcegraph/enterprise/cmd/xlang-go/internal/server"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	lsext "github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/sourcegraph/xlang"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
	"github.com/sourcegraph/sourcegraph/xlang/proxy"
	"github.com/sourcegraph/sourcegraph/xlang/uri"
)

func init() {
	// Use in-process Go language server for tests.
	proxy.ServersByMode = map[string]func() (jsonrpc2.ObjectStream, error){
		"go": func() (jsonrpc2.ObjectStream, error) {
			// Run in-process for easy development (no recompiles, etc.).
			a, b := proxy.InMemoryPeerConns()
			jsonrpc2.NewConn(context.Background(), a, jsonrpc2.AsyncHandler(gobuildserver.NewHandler()))
			return b, nil
		},
	}
}

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
				cleanup := useMapFS(test.fs)
				defer cleanup()
			}
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
			}

			ctx := context.Background()
			p := proxy.New()
			if test.rootURI == "" {
				t.Fatal("no rootPath set in test fixture")
			}

			addr, done := startProxy(t, p)
			defer done()
			c := dialProxy(t, addr, nil)

			root, err := uri.Parse(string(test.rootURI))
			if err != nil {
				t.Fatal(err)
			}

			// Prepare the connection.
			if err := c.Call(ctx, "initialize", lspext.ClientProxyInitializeParams{
				InitializeParams:      lsp.InitializeParams{RootURI: test.rootURI},
				InitializationOptions: lspext.ClientProxyInitializationOptions{Mode: test.mode},
			}, nil); err != nil {
				t.Fatal("initialize:", err)
			}

			lspTests(t, ctx, c, root, test.wantHover, test.wantDefinition, test.wantXDefinition, test.wantReferences, test.wantSymbols, test.wantXDependencies, test.wantXReferences, test.wantXPackages)
		})
	}
}

func startProxy(t testing.TB, p *proxy.Proxy) (addr string, done func()) {
	proxy.LogServerStats = false
	bindAddr := ":0"
	if os.Getenv("CI") != "" {
		// CircleCI has issues with IPv6 (e.g., "dial tcp [::]:39984:
		// connect: network is unreachable").
		bindAddr = "127.0.0.1:0"
	}
	l, err := net.Listen("tcp", bindAddr)
	if err != nil {
		t.Fatal("Listen:", err)
	}
	go p.Serve(context.Background(), l)
	return l.Addr().String(), func() {
		l.Close()
		if err := p.Close(context.Background()); err != nil && err.Error() != "jsonrpc2: connection is closed" {
			t.Fatal("proxy.Close:", err)
		}
	}
}

func dialProxy(t testing.TB, addr string, recvDiags chan<- lsp.PublishDiagnosticsParams) *jsonrpc2.Conn {
	h := &xlang.ClientHandler{
		RecvDiagnostics: func(uri lsp.DocumentURI, diags []lsp.Diagnostic) {
			if recvDiags == nil {
				var buf bytes.Buffer
				for _, d := range diags {
					fmt.Fprintf(&buf, "\t:%d:%d: %s\n", d.Range.Start.Line+1, d.Range.Start.Character+1, d.Message)
				}
				t.Logf("diagnostics: %s\n%s", uri, buf.String())
			} else {
				recvDiags <- lsp.PublishDiagnosticsParams{URI: uri, Diagnostics: diags}
			}
		},
		RecvPartialResult: func(id lsp.ID, patch interface{}) {
			p, _ := json.Marshal(patch)
			t.Logf("partialResult: %s %s", id.String(), string(p))
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	c, err := xlang.DialProxy(ctx, addr, h)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// TestProxy_connections tests that connections are created, reused
// and resumed as appropriate.
func TestProxy_connections(t *testing.T) {
	ctx := context.Background()

	cleanup := useMapFS(map[string]string{"f": "x"})
	defer cleanup()

	// Store data sent/received for checking.
	var (
		mu     sync.Mutex
		reqs   []testRequest // store received reqs
		addReq = func(req *jsonrpc2.Request) {
			mu.Lock()
			defer mu.Unlock()
			reqs = append(reqs, testRequest{req.Method, req.Params})
		}
		waitForReqs = func() error {
			for t0 := time.Now(); time.Since(t0) < 250*time.Millisecond; time.Sleep(10 * time.Millisecond) {
				mu.Lock()
				if reqs != nil {
					mu.Unlock()
					return nil
				}
				mu.Unlock()
			}
			return errors.New("timed out waiting for test server to receive req")
		}
		getAndClearReqs = func() []testRequest {
			mu.Lock()
			defer mu.Unlock()
			v := reqs
			reqs = nil
			return v
		}
		wantReqs = func(want []testRequest) error {
			if os.Getenv("CI") != "" {
				time.Sleep(3 * time.Second)
			}
			if err := waitForReqs(); err != nil {
				return err
			}
			got := getAndClearReqs()
			join := func(reqs []testRequest) (s string) {
				for i, r := range reqs {
					if i != 0 {
						s += "\n"
					}
					s += r.String()
				}
				return
			}
			sort.Sort(testRequests(got))
			sort.Sort(testRequests(want))
			if !testRequestsEqual(got, want) {
				return fmt.Errorf("got reqs != want reqs\n\nGOT REQS:\n%s\n\nWANT REQS:\n%s", join(got), join(want))
			}
			return nil
		}
	)

	// Start test build/lang server S1.
	calledConnectToTestServer := 0 // track the times we need to open a new server connection
	proxy.ServersByMode["test"] = func() (jsonrpc2.ObjectStream, error) {
		mu.Lock()
		calledConnectToTestServer++
		mu.Unlock()
		a, b := proxy.InMemoryPeerConns()
		jsonrpc2.NewConn(context.Background(), a, jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
			addReq(req)
			return nil, nil
		})))
		return b, nil
	}
	defer func() {
		delete(proxy.ServersByMode, "test")
	}()

	proxy := proxy.New()
	addr, done := startProxy(t, proxy)
	defer done()

	// We always send the same capabilities, put in variable to avoid
	// repetition.
	caps := lsp.ClientCapabilities{
		XFilesProvider:   true,
		XContentProvider: true,
		XCacheProvider:   true,
	}

	// Start the test client C1.
	c1 := dialProxy(t, addr, nil)

	// C1 connects to the proxy.
	initParams := lspext.ClientProxyInitializeParams{
		InitializeParams: lsp.InitializeParams{
			RootURI:      "test://test?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Capabilities: caps,
		},
		InitializationOptions: lspext.ClientProxyInitializationOptions{Mode: "test"},
	}
	if err := c1.Call(ctx, "initialize", initParams, nil); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond) // we're testing for a negative, so this is not as flaky as it seems; if a request is received later, it'll cause a test failure the next time we call wantReqs
	want := []testRequest{
		{"initialize", lspext.InitializeParams{
			InitializeParams: lsp.InitializeParams{
				RootPath:              "/",
				RootURI:               "file:///",
				Capabilities:          caps,
				InitializationOptions: json.RawMessage("null"),
			},
			OriginalRootURI: "test://test?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Mode:            "test",
		}},
	}
	if err := wantReqs(want); err != nil {
		t.Fatal("after C1 initialize request:", err)
	}

	// Now C1 sends an actual request. The proxy should open a
	// connection to S1, initialize it, and send the request.
	if err := c1.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#myfile"},
		Position:     lsp.Position{Line: 1, Character: 2},
	}, nil); err != nil {
		t.Fatal(err)
	}
	want = []testRequest{
		{"textDocument/definition", lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///myfile"},
			Position:     lsp.Position{Line: 1, Character: 2},
		}},
	}
	if err := wantReqs(want); err != nil {
		t.Fatal("after C1 textDocument/definition request:", err)
	}
	if want := 1; calledConnectToTestServer != want {
		t.Errorf("got %d server connections, want %d (the server should have been connected to)", calledConnectToTestServer, want)
	}

	// C1 sends another request. The server is already initialized, so
	// just the single request needs to get sent.
	if err := c1.Call(ctx, "textDocument/hover", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#myfile2"},
		Position:     lsp.Position{Line: 3, Character: 4},
	}, nil); err != nil {
		t.Fatal(err)
	}
	want = []testRequest{
		{"textDocument/hover", lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///myfile2"},
			Position:     lsp.Position{Line: 3, Character: 4}}},
	}
	if err := wantReqs(want); err != nil {
		t.Fatal("after C1 textDocument/hover request:", err)
	}

	// Kill the server to simulate either an idle shutdown by the
	// proxy, or an unexpected failure on the server.
	if err := proxy.ShutdownServers(context.Background()); err != nil {
		t.Fatal(err)
	}
	if want := 1; calledConnectToTestServer != want {
		t.Errorf("got %d server connections, want %d (after shutting down the server, did not expect proxy to reconnect until the next client request arrived that should be routed to the server)", calledConnectToTestServer, want)
	}

	// C1 does not know the server was killed. When it sends the next
	// request, the proxy should transparently spin up a new server
	// and reinitialize it appropriately.
	if err := c1.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#myfile3"},
		Position:     lsp.Position{Line: 5, Character: 6},
	}, nil); err != nil {
		t.Fatal(err)
	}
	want = []testRequest{
		{"shutdown", nil},
		{"exit", nil},
		{"initialize", lspext.InitializeParams{
			InitializeParams: lsp.InitializeParams{
				RootPath:              "/",
				RootURI:               "file:///",
				Capabilities:          caps,
				InitializationOptions: json.RawMessage("null"),
			},
			OriginalRootURI: "test://test?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			Mode:            "test",
		}},
		{"textDocument/definition", lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: "file:///myfile3"},
			Position:     lsp.Position{Line: 5, Character: 6}}},
	}
	if err := wantReqs(want); err != nil {
		t.Fatal("after C1's post-server-shutdown textDocument/definition request:", err)
	}
	if want := 2; calledConnectToTestServer != want {
		t.Errorf("got %d server connections, want %d (the server should have been reconnected to and reinitialized)", calledConnectToTestServer, want)
	}
}

// TestProxy_propagation tests that diagnostics and log messages are
// propagated from the build/lang server through the proxy to the
// client.
func TestProxy_propagation(t *testing.T) {
	ctx := context.Background()

	cleanup := useMapFS(map[string]string{"f": "x"})
	defer cleanup()

	p := proxy.New()
	addr, done := startProxy(t, p)
	defer done()

	// Start test build/lang server that sends diagnostics about any
	// file that we call textDocument/definition on.
	proxy.ServersByMode["test"] = func() (jsonrpc2.ObjectStream, error) {
		a, b := proxy.InMemoryPeerConns()
		jsonrpc2.NewConn(context.Background(), a, jsonrpc2.AsyncHandler(jsonrpc2.HandlerWithError(func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (interface{}, error) {
			if req.Method == "textDocument/definition" {
				var params lsp.TextDocumentPositionParams
				if err := json.Unmarshal(*req.Params, &params); err != nil {
					return nil, err
				}
				if err := conn.Notify(ctx, "textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{
					URI: params.TextDocument.URI,
					Diagnostics: []lsp.Diagnostic{
						{
							Range:   lsp.Range{Start: lsp.Position{Line: 1, Character: 1}, End: lsp.Position{Line: 1, Character: 1}},
							Message: "m",
						},
					},
				}); err != nil {
					return nil, err
				}
				return []lsp.Location{}, nil
			}
			return nil, nil
		})))
		return b, nil
	}
	defer func() {
		delete(proxy.ServersByMode, "test")
	}()

	recvDiags := make(chan lsp.PublishDiagnosticsParams)
	c := dialProxy(t, addr, recvDiags)

	// Connect to the proxy.
	initParams := lspext.ClientProxyInitializeParams{
		InitializeParams:      lsp.InitializeParams{RootURI: "test://test?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"},
		InitializationOptions: lspext.ClientProxyInitializationOptions{Mode: "test"},
	}
	if err := c.Call(ctx, "initialize", initParams, nil); err != nil {
		t.Fatal(err)
	}

	// Call something that triggers the server to return diagnostics.
	if err := c.Call(ctx, "textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "test://test?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#myfile"},
		Position:     lsp.Position{Line: 1, Character: 2},
	}, nil); err != nil {
		t.Fatal(err)
	}

	// Check that we got the diagnostics.
	select {
	case diags := <-recvDiags:
		want := lsp.PublishDiagnosticsParams{
			URI: "test://test?deadbeefdeadbeefdeadbeefdeadbeefdeadbeef#myfile",
			Diagnostics: []lsp.Diagnostic{
				{
					Range:   lsp.Range{Start: lsp.Position{Line: 1, Character: 1}, End: lsp.Position{Line: 1, Character: 1}},
					Message: "m",
				},
			},
		}
		if !reflect.DeepEqual(diags, want) {
			t.Errorf("got diags\n%+v\n\nwant diags\n%+v", diags, want)
		}

	case <-time.After(time.Second):
		t.Fatal("want diagnostics, got nothing before timeout")
	}
}
