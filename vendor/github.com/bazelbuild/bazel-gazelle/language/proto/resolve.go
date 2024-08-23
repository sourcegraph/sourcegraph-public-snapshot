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

package proto

import (
	"errors"
	"fmt"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/pathtools"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

func (*protoLang) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	rel := f.Pkg
	srcs := r.AttrStrings("srcs")
	imports := make([]resolve.ImportSpec, len(srcs))
	pc := GetProtoConfig(c)
	prefix := rel
	if stripImportPrefix := r.AttrString("strip_import_prefix"); stripImportPrefix != "" {
		// If strip_import_prefix starts with a /, it's interpreted as being
		// relative to the repository root. Otherwise, it's interpreted as being
		// relative to the package directory.
		//
		// So for the file //a/b:c/d.proto, if strip_import_prefix = "/a",
		// the proto should be imported as "b/c/d.proto".
		// If strip_import_prefix = "c", the proto should be imported as "d.proto".
		//
		// The package-relativeform only seems useful if there is one Bazel package
		// covering protos in subdirectories. Gazelle does not generate build files
		// like that, but we might still index proto_library rules like that,
		// so we support it here.
		if strings.HasPrefix(stripImportPrefix, "/") {
			prefix = pathtools.TrimPrefix(rel, stripImportPrefix[len("/"):])
		} else {
			prefix = pathtools.TrimPrefix(rel, path.Join(rel, pc.StripImportPrefix))
		}
		if rel == prefix {
			// Stripped prefix is not a prefix of rel, so the rule won't be buildable.
			// Don't index it.
			return nil
		}
	}
	if importPrefix := r.AttrString("import_prefix"); importPrefix != "" {
		prefix = path.Join(importPrefix, prefix)
	}
	for i, src := range srcs {
		imports[i] = resolve.ImportSpec{Lang: "proto", Imp: path.Join(prefix, src)}
	}
	return imports
}

func (*protoLang) Embeds(r *rule.Rule, from label.Label) []label.Label {
	return nil
}

func (*protoLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, importsRaw interface{}, from label.Label) {
	if importsRaw == nil {
		// may not be set in tests.
		return
	}
	imports := importsRaw.([]string)
	r.DelAttr("deps")
	depSet := make(map[string]bool)
	for _, imp := range imports {
		l, err := resolveProto(c, ix, r, imp, from)
		if err == errSkipImport {
			continue
		} else if err != nil {
			log.Print(err)
		} else {
			l = l.Rel(from.Repo, from.Pkg)
			depSet[l.String()] = true
		}
	}
	if len(depSet) > 0 {
		deps := make([]string, 0, len(depSet))
		for dep := range depSet {
			deps = append(deps, dep)
		}
		sort.Strings(deps)
		r.SetAttr("deps", deps)
	}
}

var (
	errSkipImport = errors.New("std import")
	errNotFound   = errors.New("not found")
)

func resolveProto(c *config.Config, ix *resolve.RuleIndex, r *rule.Rule, imp string, from label.Label) (label.Label, error) {
	pc := GetProtoConfig(c)
	if !strings.HasSuffix(imp, ".proto") {
		return label.NoLabel, fmt.Errorf("can't import non-proto: %q", imp)
	}

	if l, ok := resolve.FindRuleWithOverride(c, resolve.ImportSpec{Imp: imp, Lang: "proto"}, "proto"); ok {
		return l, nil
	}

	if l, ok := knownImports[imp]; ok && pc.Mode.ShouldUseKnownImports() {
		if l.Equal(from) {
			return label.NoLabel, errSkipImport
		} else {
			return l, nil
		}
	}

	if l, err := resolveWithIndex(c, ix, imp, from); err == nil || err == errSkipImport {
		return l, err
	} else if err != errNotFound {
		return label.NoLabel, err
	}

	rel := path.Dir(imp)
	if rel == "." {
		rel = ""
	}
	name := RuleName(rel)
	return label.New("", rel, name), nil
}

func resolveWithIndex(c *config.Config, ix *resolve.RuleIndex, imp string, from label.Label) (label.Label, error) {
	matches := ix.FindRulesByImportWithConfig(c, resolve.ImportSpec{Lang: "proto", Imp: imp}, "proto")
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

// CrossResolve provides dependency resolution logic for the go language extension.
func (*protoLang) CrossResolve(c *config.Config, ix *resolve.RuleIndex, imp resolve.ImportSpec, lang string) []resolve.FindResult {
	if lang != "go" {
		return nil
	}
	pc := GetProtoConfig(c)
	if imp.Lang == "proto" && pc.Mode.ShouldUseKnownImports() {
		if l, ok := knownProtoImports[imp.Imp]; ok {
			return []resolve.FindResult{{Label: l}}
		}
	}
	if imp.Lang == "go" && pc.Mode.ShouldUseKnownImports() {
		// These are commonly used libraries that depend on Well Known Types.
		// They depend on the generated versions of these protos to avoid conflicts.
		// However, since protoc-gen-go depends on these libraries, we generate
		// its rules in disable_global mode (to avoid cyclic dependency), so the
		// "go_default_library" versions of these libraries depend on the
		// pre-generated versions of the proto libraries.
		switch imp.Imp {
		case "github.com/golang/protobuf/proto":
			return []resolve.FindResult{{Label: label.New("com_github_golang_protobuf", "proto", "go_default_library")}}
		case "github.com/golang/protobuf/jsonpb":
			return []resolve.FindResult{{Label: label.New("com_github_golang_protobuf", "jsonpb", "go_default_library_gen")}}
		case "github.com/golang/protobuf/descriptor":
			return []resolve.FindResult{{Label: label.New("com_github_golang_protobuf", "descriptor", "go_default_library_gen")}}
		case "github.com/golang/protobuf/ptypes":
			return []resolve.FindResult{{Label: label.New("com_github_golang_protobuf", "ptypes", "go_default_library_gen")}}
		case "github.com/golang/protobuf/protoc-gen-go/generator":
			return []resolve.FindResult{{Label: label.New("com_github_golang_protobuf", "protoc-gen-go/generator", "go_default_library_gen")}}
		case "google.golang.org/grpc":
			return []resolve.FindResult{{Label: label.New("org_golang_google_grpc", "", "go_default_library")}}
		}
		if l, ok := knownGoProtoImports[imp.Imp]; ok {
			return []resolve.FindResult{{Label: l}}
		}
	}
	return nil
}
