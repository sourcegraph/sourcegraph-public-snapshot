/* Copyright 2017 The Bazel Authors. All rights reserved.

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
	"log"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/rule"
	bzl "github.com/bazelbuild/buildtools/build"
)

func (*goLang) Fix(c *config.Config, f *rule.File) {
	migrateLibraryEmbed(c, f)
	migrateGrpcCompilers(c, f)
	flattenSrcs(c, f)
	squashCgoLibrary(c, f)
	squashXtest(c, f)
	removeLegacyProto(c, f)
	removeLegacyGazelle(c, f)
	migrateNamingConvention(c, f)
}

// migrateNamingConvention renames rules according to go_naming_convention
// directives.
func migrateNamingConvention(c *config.Config, f *rule.File) {
	// Determine old and new names for go_library and go_test.
	nc := getGoConfig(c).goNamingConvention
	importPath := InferImportPath(c, f.Pkg)
	if importPath == "" {
		return
	}
	var pkgName string // unknown unless there's a binary
	if fileContainsGoBinary(c, f) {
		pkgName = "main"
	}
	libName := libNameByConvention(nc, importPath, pkgName)
	testName := testNameByConvention(nc, importPath)
	var migrateLibName, migrateTestName string
	switch nc {
	case goDefaultLibraryNamingConvention:
		migrateLibName = libNameByConvention(importNamingConvention, importPath, pkgName)
		migrateTestName = testNameByConvention(importNamingConvention, importPath)
	case importNamingConvention, importAliasNamingConvention:
		migrateLibName = defaultLibName
		migrateTestName = defaultTestName
	default:
		return
	}

	// Check whether the new names are in use. If there are rules with both old
	// and new names, there will be a conflict.
	var haveLib, haveMigrateLib, haveTest, haveMigrateTest bool
	for _, r := range f.Rules {
		switch {
		case r.Name() == libName:
			haveLib = true
		case r.Kind() == "go_library" && r.Name() == migrateLibName && r.AttrString("importpath") == importPath:
			haveMigrateLib = true
		case r.Name() == testName:
			haveTest = true
		case r.Kind() == "go_test" && r.Name() == migrateTestName && strListAttrContains(r, "embed", ":"+migrateLibName):
			haveMigrateTest = true
		}
	}
	if haveLib && haveMigrateLib {
		log.Printf("%[1]s: Tried to rename %[2]s to %[3]s, but %[3]s already exists.", f.Path, migrateLibName, libName)
	}
	if haveTest && haveMigrateTest {
		log.Printf("%[1]s: Tried to rename %[2]s to %[3]s, but %[3]s already exists.", f.Path, migrateTestName, testName)
	}
	shouldMigrateLib := haveMigrateLib && !haveLib
	shouldMigrateTest := haveMigrateTest && !haveTest

	// Rename the targets and stuff in the same file that refers to them.
	for _, r := range f.Rules {
		// TODO(jayconrod): support map_kind directive.
		// We'll need to move the metaresolver from resolve.RuleIndex to config.Config so we can access it from here.
		switch r.Kind() {
		case "go_binary":
			if haveMigrateLib && shouldMigrateLib {
				replaceInStrListAttr(r, "embed", ":"+migrateLibName, ":"+libName)
			}
		case "go_library":
			if r.Name() == migrateLibName && shouldMigrateLib {
				r.SetName(libName)
			}
		case "go_test":
			if r.Name() == migrateTestName && shouldMigrateTest {
				r.SetName(testName)
			}
			if shouldMigrateLib {
				replaceInStrListAttr(r, "embed", ":"+migrateLibName, ":"+libName)
			}
		}
	}
}

// fileContainsGoBinary returns whether the file has a go_binary rule.
func fileContainsGoBinary(c *config.Config, f *rule.File) bool {
	if f == nil {
		return false
	}
	for _, r := range f.Rules {
		kind := r.Kind()
		if kind == "go_binary" {
			return true
		}

		if mappedKind, ok := c.KindMap["go_binary"]; ok {
			if mappedKind.KindName == kind {
				return true
			}
		}
	}
	return false
}

func replaceInStrListAttr(r *rule.Rule, attr, old, new string) {
	items := r.AttrStrings(attr)
	changed := false
	for i := range items {
		if items[i] == old {
			changed = true
			items[i] = new
		}
	}
	if changed {
		r.SetAttr(attr, items)
	}
}

func strListAttrContains(r *rule.Rule, attr, s string) bool {
	items := r.AttrStrings(attr)
	for _, item := range items {
		if item == s {
			return true
		}
	}
	return false
}

// migrateLibraryEmbed converts "library" attributes to "embed" attributes,
// preserving comments. This only applies to Go rules, and only if there is
// no keep comment on "library" and no existing "embed" attribute.
func migrateLibraryEmbed(c *config.Config, f *rule.File) {
	for _, r := range f.Rules {
		if !isGoRule(r.Kind()) {
			continue
		}
		libExpr := r.Attr("library")
		if libExpr == nil || rule.ShouldKeep(libExpr) || r.Attr("embed") != nil {
			continue
		}
		r.DelAttr("library")
		r.SetAttr("embed", &bzl.ListExpr{List: []bzl.Expr{libExpr}})
	}
}

// migrateGrpcCompilers converts "go_grpc_library" rules into "go_proto_library"
// rules with a "compilers" attribute.
func migrateGrpcCompilers(c *config.Config, f *rule.File) {
	for _, r := range f.Rules {
		if r.Kind() != "go_grpc_library" || r.ShouldKeep() || r.Attr("compilers") != nil {
			continue
		}
		r.SetKind("go_proto_library")
		r.SetAttr("compilers", []string{grpcCompilerLabel})
	}
}

// squashCgoLibrary removes cgo_library rules with the default name and
// merges their attributes with go_library with the default name. If no
// go_library rule exists, a new one will be created.
//
// Note that the library attribute is disregarded, so cgo_library and
// go_library attributes will be squashed even if the cgo_library was unlinked.
// MergeFile will remove unused values and attributes later.
func squashCgoLibrary(c *config.Config, f *rule.File) {
	// Find the default cgo_library and go_library rules.
	var cgoLibrary, goLibrary *rule.Rule
	for _, r := range f.Rules {
		if r.Kind() == "cgo_library" && r.Name() == "cgo_default_library" && !r.ShouldKeep() {
			if cgoLibrary != nil {
				log.Printf("%s: when fixing existing file, multiple cgo_library rules with default name found", f.Path)
				continue
			}
			cgoLibrary = r
			continue
		}
		if r.Kind() == "go_library" && r.Name() == defaultLibName {
			if goLibrary != nil {
				log.Printf("%s: when fixing existing file, multiple go_library rules with default name referencing cgo_library found", f.Path)
			}
			goLibrary = r
			continue
		}
	}

	if cgoLibrary == nil {
		return
	}
	if !c.ShouldFix {
		log.Printf("%s: cgo_library is deprecated. Run 'gazelle fix' to squash with go_library.", f.Path)
		return
	}

	if goLibrary == nil {
		cgoLibrary.SetKind("go_library")
		cgoLibrary.SetName(defaultLibName)
		cgoLibrary.SetAttr("cgo", true)
		return
	}

	if err := rule.SquashRules(cgoLibrary, goLibrary, f.Path); err != nil {
		log.Print(err)
		return
	}
	goLibrary.DelAttr("embed")
	goLibrary.SetAttr("cgo", true)
	cgoLibrary.Delete()
}

// squashXtest removes go_test rules with the default external name and merges
// their attributes with a go_test rule with the default internal name. If
// no internal go_test rule exists, a new one will be created (effectively
// renaming the old rule).
func squashXtest(c *config.Config, f *rule.File) {
	// Search for internal and external tests.
	var itest, xtest *rule.Rule
	for _, r := range f.Rules {
		if r.Kind() != "go_test" {
			continue
		}
		if r.Name() == defaultTestName {
			itest = r
		} else if r.Name() == "go_default_xtest" {
			xtest = r
		}
	}

	if xtest == nil || xtest.ShouldKeep() || (itest != nil && itest.ShouldKeep()) {
		return
	}
	if !c.ShouldFix {
		if itest == nil {
			log.Printf("%s: go_default_xtest is no longer necessary. Run 'gazelle fix' to rename to go_default_test.", f.Path)
		} else {
			log.Printf("%s: go_default_xtest is no longer necessary. Run 'gazelle fix' to squash with go_default_test.", f.Path)
		}
		return
	}

	// If there was no internal test, we can just rename the external test.
	if itest == nil {
		xtest.SetName(defaultTestName)
		return
	}

	// Attempt to squash.
	if err := rule.SquashRules(xtest, itest, f.Path); err != nil {
		log.Print(err)
		return
	}
	xtest.Delete()
}

// flattenSrcs transforms srcs attributes structured as concatenations of
// lists and selects (generated from PlatformStrings; see
// extractPlatformStringsExprs for matching details) into a sorted,
// de-duplicated list. Comments are accumulated and de-duplicated across
// duplicate expressions.
func flattenSrcs(c *config.Config, f *rule.File) {
	for _, r := range f.Rules {
		if !isGoRule(r.Kind()) {
			continue
		}
		oldSrcs := r.Attr("srcs")
		if oldSrcs == nil {
			continue
		}
		flatSrcs := rule.FlattenExpr(oldSrcs)
		if flatSrcs != oldSrcs {
			r.SetAttr("srcs", flatSrcs)
		}
	}
}

// removeLegacyProto removes uses of the old proto rules. It deletes loads
// from go_proto_library.bzl. It deletes proto filegroups. It removes
// go_proto_library attributes which are no longer recognized. New rules
// are generated in place of the deleted rules, but attributes and comments
// are not migrated.
func removeLegacyProto(c *config.Config, f *rule.File) {
	// Don't fix if the proto mode was set to something other than the default.
	if pcMode := getProtoMode(c); pcMode != proto.DefaultMode {
		return
	}

	// Scan for definitions to delete.
	var protoLoads []*rule.Load
	for _, l := range f.Loads {
		if l.Name() == "@io_bazel_rules_go//proto:go_proto_library.bzl" {
			protoLoads = append(protoLoads, l)
		}
	}
	var protoFilegroups, protoRules []*rule.Rule
	for _, r := range f.Rules {
		if r.Kind() == "filegroup" && r.Name() == legacyProtoFilegroupName {
			protoFilegroups = append(protoFilegroups, r)
		}
		if r.Kind() == "go_proto_library" {
			protoRules = append(protoRules, r)
		}
	}
	if len(protoLoads)+len(protoFilegroups) == 0 {
		return
	}
	if !c.ShouldFix {
		log.Printf("%s: go_proto_library.bzl is deprecated. Run 'gazelle fix' to replace old rules.", f.Path)
		return
	}

	// Delete legacy proto loads and filegroups. Only delete go_proto_library
	// rules if we deleted a load.
	for _, l := range protoLoads {
		l.Delete()
	}
	for _, r := range protoFilegroups {
		r.Delete()
	}
	if len(protoLoads) > 0 {
		for _, r := range protoRules {
			r.Delete()
		}
	}
}

// removeLegacyGazelle removes loads of the "gazelle" macro from
// @io_bazel_rules_go//go:def.bzl. The definition has moved to
// @bazel_gazelle//:def.bzl, and the old one will be deleted soon.
func removeLegacyGazelle(c *config.Config, f *rule.File) {
	for _, l := range f.Loads {
		if l.Name() == "@io_bazel_rules_go//go:def.bzl" && l.Has("gazelle") {
			l.Remove("gazelle")
			if l.IsEmpty() {
				l.Delete()
			}
		}
	}
}

func isGoRule(kind string) bool {
	return kind == "go_library" ||
		kind == "go_binary" ||
		kind == "go_test" ||
		kind == "go_proto_library" ||
		kind == "go_grpc_library"
}
