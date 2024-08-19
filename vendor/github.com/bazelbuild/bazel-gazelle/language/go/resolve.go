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
	"errors"
	"fmt"
	"go/build"
	"log"
	"path"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/pathtools"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func (*goLang) Imports(_ *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	if !isGoLibrary(r.Kind()) || isExtraLibrary(r) {
		return nil
	}
	if importPath := r.AttrString("importpath"); importPath == "" {
		return []resolve.ImportSpec{}
	} else {
		return []resolve.ImportSpec{{
			Lang: goName,
			Imp:  importPath,
		}}
	}
}

func (*goLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	embedStrings := r.AttrStrings("embed")
	if isGoProtoLibrary(r.Kind()) {
		embedStrings = append(embedStrings, r.AttrString("proto"))
	}
	embedLabels := make([]label.Label, 0, len(embedStrings))
	for _, s := range embedStrings {
		l, err := label.Parse(s)
		if err != nil {
			continue
		}
		l = l.Abs(from.Repo, from.Pkg)
		embedLabels = append(embedLabels, l)
	}
	return embedLabels
}

func (gl *goLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, importsRaw interface{}, from label.Label) {
	if importsRaw == nil {
		// may not be set in tests.
		return
	}
	imports := importsRaw.(rule.PlatformStrings)
	r.DelAttr("deps")
	var resolve func(*config.Config, *resolve.RuleIndex, *repo.RemoteCache, string, label.Label) (label.Label, error)
	switch r.Kind() {
	case "go_proto_library":
		resolve = resolveProto
	default:
		resolve = ResolveGo
	}
	deps, errs := imports.Map(func(imp string) (string, error) {
		l, err := resolve(c, ix, rc, imp, from)
		if err == errSkipImport {
			return "", nil
		} else if err != nil {
			return "", err
		}
		for _, embed := range gl.Embeds(r, from) {
			if embed.Equal(l) {
				return "", nil
			}
		}
		l = l.Rel(from.Repo, from.Pkg)
		return l.String(), nil
	})
	for _, err := range errs {
		log.Print(err)
	}
	if !deps.IsEmpty() {
		if r.Kind() == "go_proto_library" {
			// protos may import the same library multiple times by different names,
			// so we need to de-duplicate them. Protos are not platform-specific,
			// so it's safe to just flatten them.
			r.SetAttr("deps", deps.Flat())
		} else {
			r.SetAttr("deps", deps)
		}
	}
}

var (
	errSkipImport = errors.New("std or self import")
	errNotFound   = errors.New("rule not found")
)

// ResolveGo resolves a Go import path to a Bazel label, possibly using the
// given rule index and remote cache. Some special cases may be applied to
// known proto import paths, depending on the current proto mode.
//
// This may be used directly by other language extensions related to Go
// (gomock). Gazelle calls Language.Resolve instead.
func ResolveGo(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, imp string, from label.Label) (label.Label, error) {
	gc := getGoConfig(c)
	if build.IsLocalImport(imp) {
		cleanRel := path.Clean(path.Join(from.Pkg, imp))
		if build.IsLocalImport(cleanRel) {
			return label.NoLabel, fmt.Errorf("relative import path %q from %q points outside of repository", imp, from.Pkg)
		}
		imp = path.Join(gc.prefix, cleanRel)
	}

	if IsStandard(imp) {
		return label.NoLabel, errSkipImport
	}

	if l, ok := resolve.FindRuleWithOverride(c, resolve.ImportSpec{Lang: "go", Imp: imp}, "go"); ok {
		return l, nil
	}

	if l, err := resolveWithIndexGo(c, ix, imp, from); err == nil || err == errSkipImport {
		return l, err
	} else if err != errNotFound {
		return label.NoLabel, err
	}

	// Special cases for rules_go and bazel_gazelle.
	// These have names that don't following conventions and they're
	// typeically declared with http_archive, not go_repository, so Gazelle
	// won't recognize them.
	if !c.Bzlmod {
		if pathtools.HasPrefix(imp, "github.com/bazelbuild/rules_go") {
			pkg := pathtools.TrimPrefix(imp, "github.com/bazelbuild/rules_go")
			return label.New("io_bazel_rules_go", pkg, "go_default_library"), nil
		} else if pathtools.HasPrefix(imp, "github.com/bazelbuild/bazel-gazelle") {
			pkg := pathtools.TrimPrefix(imp, "github.com/bazelbuild/bazel-gazelle")
			return label.New("bazel_gazelle", pkg, "go_default_library"), nil
		}
	}

	if !c.IndexLibraries {
		// packages in current repo were not indexed, relying on prefix to decide what may have been in
		// current repo
		if pathtools.HasPrefix(imp, gc.prefix) {
			pkg := path.Join(gc.prefixRel, pathtools.TrimPrefix(imp, gc.prefix))
			libName := libNameByConvention(gc.goNamingConvention, imp, "")
			return label.New("", pkg, libName), nil
		}
	}

	if gc.depMode == vendorMode {
		return resolveVendored(gc, imp)
	}
	var resolveFn func(string) (string, string, error)
	if gc.depMode == staticMode {
		resolveFn = rc.RootStatic
	} else if gc.moduleMode || pathWithoutSemver(imp) != "" {
		resolveFn = rc.Mod
	} else {
		resolveFn = rc.Root
	}
	return resolveToExternalLabel(c, resolveFn, imp)
}

// IsStandard returns whether a package is in the standard library.
func IsStandard(imp string) bool {
	return stdPackages[imp]
}

func resolveWithIndexGo(c *config.Config, ix *resolve.RuleIndex, imp string, from label.Label) (label.Label, error) {
	matches := ix.FindRulesByImportWithConfig(c, resolve.ImportSpec{Lang: "go", Imp: imp}, "go")
	var bestMatch resolve.FindResult
	var bestMatchIsVendored bool
	var bestMatchVendorRoot string
	var bestMatchEmbedsProtos bool
	var matchError error
	goRepositoryMode := getGoConfig(c).goRepositoryMode

	for _, m := range matches {
		// Apply vendoring logic for Go libraries. A library in a vendor directory
		// is only visible in the parent tree. Vendored libraries supercede
		// non-vendored libraries, and libraries closer to from.Pkg supercede
		// those further up the tree.
		//
		// Also, in external repositories, prefer go_proto_library targets to checked-in .go files
		// pregenerated from .proto files over go_proto_library targets. Ideally, the two should be
		// in sync. If not, users can choose between the two by using the go_generate_proto
		// directive.
		isVendored := false
		vendorRoot := ""
		parts := strings.Split(m.Label.Pkg, "/")
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == "vendor" {
				isVendored = true
				vendorRoot = strings.Join(parts[:i], "/")
				break
			}
		}
		if isVendored && !label.New(m.Label.Repo, vendorRoot, "").Contains(from) {
			// vendor directory not visible
			continue
		}

		embedsProtos := false
		for _, embed := range m.Embeds {
			if strings.HasSuffix(embed.Name, goProtoSuffix) {
				embedsProtos = true
			}
		}

		if bestMatch.Label.Equal(label.NoLabel) ||
			(isVendored && (!bestMatchIsVendored || len(vendorRoot) > len(bestMatchVendorRoot))) ||
			(goRepositoryMode && !bestMatchEmbedsProtos && embedsProtos) {
			// Current match is better
			bestMatch = m
			bestMatchIsVendored = isVendored
			bestMatchVendorRoot = vendorRoot
			bestMatchEmbedsProtos = embedsProtos
			matchError = nil
		} else if (!isVendored && bestMatchIsVendored) ||
			(isVendored && len(vendorRoot) < len(bestMatchVendorRoot)) ||
			(goRepositoryMode && bestMatchEmbedsProtos && !embedsProtos) {
			// Current match is worse
		} else {
			// Match is ambiguous
			// TODO: consider listing all the ambiguous rules here.
			matchError = fmt.Errorf("rule %s imports %q which matches multiple rules: %s and %s. # gazelle:resolve may be used to disambiguate", from, imp, bestMatch.Label, m.Label)
		}
	}
	if matchError != nil {
		return label.NoLabel, matchError
	}
	if bestMatch.Label.Equal(label.NoLabel) {
		return label.NoLabel, errNotFound
	}
	if bestMatch.IsSelfImport(from) {
		return label.NoLabel, errSkipImport
	}
	return bestMatch.Label, nil
}

func resolveToExternalLabel(c *config.Config, resolveFn func(string) (string, string, error), imp string) (label.Label, error) {
	prefix, repo, err := resolveFn(imp)
	if err != nil {
		return label.NoLabel, err
	} else if prefix == "" && repo == "" {
		return label.NoLabel, errSkipImport
	}

	var pkg string
	if pathtools.HasPrefix(imp, prefix) {
		pkg = pathtools.TrimPrefix(imp, prefix)
	} else if impWithoutSemver := pathWithoutSemver(imp); pathtools.HasPrefix(impWithoutSemver, prefix) {
		// We may have used minimal module compatibility to resolve a path
		// without a semantic import version suffix to a repository that has one.
		pkg = pathtools.TrimPrefix(impWithoutSemver, prefix)
	}

	// Determine what naming convention is used by the repository.
	// If there is no known repository, it's probably declared in an http_archive
	// somewhere like go_rules_dependencies, so use the old naming convention,
	// unless the user has explicitly told us otherwise.
	// If the repository uses the import_alias convention (default for
	// go_repository), use the convention from the current directory unless the
	// user has told us otherwise.
	gc := getGoConfig(c)
	nc := gc.repoNamingConvention[repo]
	if nc == unknownNamingConvention {
		if gc.goNamingConventionExternal != unknownNamingConvention {
			nc = gc.goNamingConventionExternal
		} else {
			nc = goDefaultLibraryNamingConvention
		}
	} else if nc == importAliasNamingConvention {
		if gc.goNamingConventionExternal != unknownNamingConvention {
			nc = gc.goNamingConventionExternal
		} else {
			nc = gc.goNamingConvention
		}
	}

	name := libNameByConvention(nc, imp, "")
	return label.New(repo, pkg, name), nil
}

func resolveVendored(gc *goConfig, imp string) (label.Label, error) {
	name := libNameByConvention(gc.goNamingConvention, imp, "")
	return label.New("", path.Join("vendor", imp), name), nil
}

func resolveProto(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, imp string, from label.Label) (label.Label, error) {
	if wellKnownProtos[imp] {
		return label.NoLabel, errSkipImport
	}

	if l, ok := resolve.FindRuleWithOverride(c, resolve.ImportSpec{Lang: "proto", Imp: imp}, "go"); ok {
		return l, nil
	}

	if l, err := resolveWithIndexProto(c, ix, imp, from); err == nil || err == errSkipImport {
		return l, err
	} else if err != errNotFound {
		return label.NoLabel, err
	}

	// As a fallback, guess the label based on the proto file name. We assume
	// all proto files in a directory belong to the same package, and the
	// package name matches the directory base name. We also assume that protos
	// in the vendor directory must refer to something else in vendor.
	rel := path.Dir(imp)
	if rel == "." {
		rel = ""
	}
	if from.Pkg == "vendor" || strings.HasPrefix(from.Pkg, "vendor/") {
		rel = path.Join("vendor", rel)
	}
	libName := libNameByConvention(getGoConfig(c).goNamingConvention, imp, "")
	return label.New("", rel, libName), nil
}

// wellKnownProtos is the set of proto sets for which we don't need to add
// an explicit dependency in go_proto_library.
// TODO(jayconrod): generate from
// @io_bazel_rules_go//proto/wkt:WELL_KNOWN_TYPE_PACKAGES
var wellKnownProtos = map[string]bool{
	"google/protobuf/any.proto":             true,
	"google/protobuf/api.proto":             true,
	"google/protobuf/compiler/plugin.proto": true,
	"google/protobuf/descriptor.proto":      true,
	"google/protobuf/duration.proto":        true,
	"google/protobuf/empty.proto":           true,
	"google/protobuf/field_mask.proto":      true,
	"google/protobuf/source_context.proto":  true,
	"google/protobuf/struct.proto":          true,
	"google/protobuf/timestamp.proto":       true,
	"google/protobuf/type.proto":            true,
	"google/protobuf/wrappers.proto":        true,
}

func resolveWithIndexProto(c *config.Config, ix *resolve.RuleIndex, imp string, from label.Label) (label.Label, error) {
	matches := ix.FindRulesByImportWithConfig(c, resolve.ImportSpec{Lang: "proto", Imp: imp}, "go")
	if len(matches) == 0 {
		return label.NoLabel, errNotFound
	}
	if len(matches) > 1 {
		return label.NoLabel, fmt.Errorf("multiple rules (%s and %s) may be imported with %q from %s", matches[0].Label, matches[1].Label, imp, from)
	}
	if matches[0].IsSelfImport(from) {
		return label.NoLabel, errSkipImport
	}
	return matches[0].Label, nil
}

func isGoLibrary(kind string) bool {
	return kind == "go_library" || isGoProtoLibrary(kind)
}

func isGoProtoLibrary(kind string) bool {
	return kind == "go_proto_library" || kind == "go_grpc_library"
}

// isExtraLibrary returns true if this rule is one of a handful of proto
// libraries generated by maybeGenerateExtraLib. It should not be indexed for
// dependency resolution.
func isExtraLibrary(r *rule.Rule) bool {
	if !strings.HasSuffix(r.Name(), "_gen") {
		return false
	}
	switch r.AttrString("importpath") {
	case "github.com/golang/protobuf/descriptor",
		"github.com/golang/protobuf/protoc-gen-go/generator",
		"github.com/golang/protobuf/ptypes":
		return true
	default:
		return false
	}
}
