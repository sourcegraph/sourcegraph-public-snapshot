/* Copyright 2018 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package golang

import (
	"fmt"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

var goKinds = map[string]rule.KindInfo{
	"alias": {
		NonEmptyAttrs:  map[string]bool{"actual": true},
		MergeableAttrs: map[string]bool{"actual": true},
	},
	"filegroup": {
		NonEmptyAttrs:  map[string]bool{"srcs": true},
		MergeableAttrs: map[string]bool{"srcs": true},
	},
	"go_binary": {
		MatchAny: true,
		NonEmptyAttrs: map[string]bool{
			"deps":  true,
			"embed": true,
			"srcs":  true,
		},
		SubstituteAttrs: map[string]bool{"embed": true},
		MergeableAttrs: map[string]bool{
			"cgo":       true,
			"clinkopts": true,
			"cppopts":   true,
			"copts":     true,
			"cxxopts":   true,
			"embed":     true,
			"embedsrcs": true,
			"srcs":      true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
	"go_library": {
		MatchAttrs: []string{"importpath"},
		NonEmptyAttrs: map[string]bool{
			"deps":  true,
			"embed": true,
			"srcs":  true,
		},
		SubstituteAttrs: map[string]bool{
			"embed": true,
		},
		MergeableAttrs: map[string]bool{
			"cgo":        true,
			"clinkopts":  true,
			"cppopts":    true,
			"copts":      true,
			"cxxopts":    true,
			"embed":      true,
			"embedsrcs":  true,
			"importmap":  true,
			"importpath": true,
			"srcs":       true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
	"go_proto_library": {
		MatchAttrs: []string{"importpath"},
		NonEmptyAttrs: map[string]bool{
			"deps":  true,
			"embed": true,
			"proto": true,
			"srcs":  true,
		},
		SubstituteAttrs: map[string]bool{"proto": true},
		MergeableAttrs: map[string]bool{
			"srcs":       true,
			"importpath": true,
			"importmap":  true,
			"cgo":        true,
			"clinkopts":  true,
			"cppopts":    true,
			"copts":      true,
			"cxxopts":    true,
			"embed":      true,
			"proto":      true,
			"compilers":  true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
	"go_repository": {
		MatchAttrs: []string{"importpath"},
		NonEmptyAttrs: map[string]bool{
			"importpath": true,
		},
		MergeableAttrs: map[string]bool{
			"commit":       true,
			"build_tags":   true,
			"importpath":   true,
			"remote":       true,
			"replace":      true,
			"sha256":       true,
			"strip_prefix": true,
			"sum":          true,
			"tag":          true,
			"type":         true,
			"urls":         true,
			"vcs":          true,
			"version":      true,
		},
	},
	"go_test": {
		NonEmptyAttrs: map[string]bool{
			"deps":  true,
			"embed": true,
			"srcs":  true,
		},
		MergeableAttrs: map[string]bool{
			"cgo":       true,
			"clinkopts": true,
			"cppopts":   true,
			"copts":     true,
			"cxxopts":   true,
			"embed":     true,
			"embedsrcs": true,
			"srcs":      true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
	// HACK(#834): remove when bazelbuild/rules_go#2374 is resolved.
	"go_tool_library": {
		MatchAttrs: []string{"importpath"},
		NonEmptyAttrs: map[string]bool{
			"deps":  true,
			"embed": true,
			"srcs":  true,
		},
		SubstituteAttrs: map[string]bool{
			"embed": true,
		},
		MergeableAttrs: map[string]bool{
			"cgo":        true,
			"clinkopts":  true,
			"cppopts":    true,
			"copts":      true,
			"cxxopts":    true,
			"embed":      true,
			"importmap":  true,
			"importpath": true,
			"srcs":       true,
		},
		ResolveAttrs: map[string]bool{"deps": true},
	},
}

func (*goLang) Kinds() map[string]rule.KindInfo { return goKinds }

func (*goLang) Loads() []rule.LoadInfo {
	panic("ApparentLoads should be called instead")
}

func (*goLang) ApparentLoads(moduleToApparentName func(string) string) []rule.LoadInfo {
	return apparentLoads(moduleToApparentName)
}

func apparentLoads(moduleToApparentName func(string) string) []rule.LoadInfo {
	rulesGo := moduleToApparentName("rules_go")
	if rulesGo == "" {
		rulesGo = "io_bazel_rules_go"
	}
	gazelle := moduleToApparentName("gazelle")
	if gazelle == "" {
		gazelle = "bazel_gazelle"
	}

	return []rule.LoadInfo{
		{
			Name: fmt.Sprintf("@%s//go:def.bzl", rulesGo),
			Symbols: []string{
				"cgo_library",
				"go_binary",
				"go_library",
				"go_prefix",
				"go_repository",
				"go_test",
				"go_tool_library",
			},
		}, {
			Name: fmt.Sprintf("@%s//proto:def.bzl", rulesGo),
			Symbols: []string{
				"go_grpc_library",
				"go_proto_library",
			},
		}, {
			Name: fmt.Sprintf("@%s//:deps.bzl", gazelle),
			Symbols: []string{
				"go_repository",
			},
			After: []string{
				"go_rules_dependencies",
				"go_register_toolchains",
				"gazelle_dependencies",
			},
		},
	}
}

var goLoadsForTesting = apparentLoads(func(string) string { return "" })
