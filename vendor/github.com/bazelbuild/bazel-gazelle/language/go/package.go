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
	"fmt"
	"log"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/pathtools"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// goPackage contains metadata for a set of .go and .proto files that can be
// used to generate Go rules.
type goPackage struct {
	name, dir, rel        string
	library, binary, test goTarget
	tests                 []goTarget
	proto                 protoTarget
	hasTestdata           bool
	hasMainFunction       bool
	importPath            string
}

// goTarget contains information used to generate an individual Go rule
// (library, binary, or test).
type goTarget struct {
	sources, embedSrcs, imports, cppopts, copts, cxxopts, clinkopts platformStringsBuilder
	cgo, hasInternalTest                                            bool
}

// protoTarget contains information used to generate a go_proto_library rule.
type protoTarget struct {
	name        string
	sources     platformStringsBuilder
	imports     platformStringsBuilder
	hasServices bool
}

// platformStringsBuilder is used to construct rule.PlatformStrings. Bazel
// has some requirements for deps list (a dependency cannot appear in more
// than one select expression; dependencies cannot be duplicated), so we need
// to build these carefully.
type platformStringsBuilder struct {
	strs map[string]platformStringInfo
}

// platformStringInfo contains information about a single string (source,
// import, or option).
type platformStringInfo struct {
	set                 platformStringSet
	osConstraints       map[string]bool
	archConstraints     map[string]bool
	platformConstraints map[rule.PlatformConstraint]bool
}

type platformStringSet int

const (
	genericSet platformStringSet = iota
	osSet
	archSet
	platformSet
)

// Matches a package version, eg. the end segment of 'example.com/foo/v1'
var pkgVersionRe = regexp.MustCompile("^v[0-9]+$")

// addFile adds the file described by "info" to a target in the package "p" if
// the file is buildable.
//
// "cgo" tells whether any ".go" file in the package contains cgo code. This
// affects whether C files are added to targets.
//
// An error is returned if a file is buildable but invalid (for example, a
// test .go file containing cgo code). Files that are not buildable will not
// be added to any target (for example, .txt files).
func (pkg *goPackage) addFile(c *config.Config, er *embedResolver, info fileInfo, cgo bool) error {
	switch {
	case info.ext == unknownExt || !cgo && (info.ext == cExt || info.ext == csExt):
		return nil
	case info.ext == protoExt:
		if pcMode := getProtoMode(c); pcMode == proto.LegacyMode {
			// Only add files in legacy mode. This is used to generate a filegroup
			// that contains all protos. In order modes, we get the .proto files
			// from information emitted by the proto language extension.
			pkg.proto.addFile(info)
		}
	case info.isTest:
		if info.isCgo {
			return fmt.Errorf("%s: use of cgo in test not supported", info.path)
		}
		if getGoConfig(c).testMode == fileTestMode || len(pkg.tests) == 0 {
			pkg.tests = append(pkg.tests, goTarget{})
		}
		// Add the the file to the most recently added test target (in fileTestMode)
		// or the only test target (in defaultMode).
		// In both cases, this will be the last element in the slice.
		test := &pkg.tests[len(pkg.tests)-1]
		test.addFile(c, er, info)
		if !info.isExternalTest {
			test.hasInternalTest = true
		}
	default:
		pkg.hasMainFunction = pkg.hasMainFunction || info.hasMainFunction
		pkg.library.addFile(c, er, info)
	}

	return nil
}

// isCommand returns true if the package name is "main".
func (pkg *goPackage) isCommand() bool {
	return pkg.name == "main" && pkg.hasMainFunction
}

// isBuildable returns true if anything in the package is buildable.
// This is true if the package has Go code that satisfies build constraints
// on any platform or has proto files not in legacy mode.
func (pkg *goPackage) isBuildable(c *config.Config) bool {
	return pkg.firstGoFile() != "" || !pkg.proto.sources.isEmpty()
}

// firstGoFile returns the name of a .go file if the package contains at least
// one .go file, or "" otherwise.
func (pkg *goPackage) firstGoFile() string {
	goSrcs := []platformStringsBuilder{
		pkg.library.sources,
		pkg.binary.sources,
	}
	for _, test := range pkg.tests {
		goSrcs = append(goSrcs, test.sources)
	}

	for _, sb := range goSrcs {
		if sb.strs != nil {
			for s := range sb.strs {
				if strings.HasSuffix(s, ".go") {
					return s
				}
			}
		}
	}
	return ""
}

func (pkg *goPackage) haveCgo() bool {
	if pkg.library.cgo || pkg.binary.cgo {
		return true
	}
	for _, t := range pkg.tests {
		if t.cgo {
			return true
		}
	}
	return false
}

func (pkg *goPackage) inferImportPath(c *config.Config) error {
	if pkg.importPath != "" {
		log.Panic("importPath already set")
	}
	gc := getGoConfig(c)
	if !gc.prefixSet {
		return fmt.Errorf("%s: go prefix is not set, so importpath can't be determined for rules. Set a prefix with a '# gazelle:prefix' comment or with -go_prefix on the command line", pkg.dir)
	}
	pkg.importPath = InferImportPath(c, pkg.rel)
	return nil
}

// libNameFromImportPath returns a a suitable go_library name based on the import path.
// Major version suffixes (eg. "v1") are dropped.
func libNameFromImportPath(dir string) string {
	i := strings.LastIndexAny(dir, "/\\")
	if i < 0 {
		return dir
	}
	name := dir[i+1:]
	if pkgVersionRe.MatchString(name) {
		dir := dir[:i]
		i = strings.LastIndexAny(dir, "/\\")
		if i >= 0 {
			name = dir[i+1:]
		}
	}
	return strings.ReplaceAll(name, ".", "_")
}

// libNameByConvention returns a suitable name for a go_library using the given
// naming convention, the import path, and the package name.
func libNameByConvention(nc namingConvention, imp string, pkgName string) string {
	if nc == goDefaultLibraryNamingConvention {
		return defaultLibName
	}
	name := libNameFromImportPath(imp)
	isCommand := pkgName == "main"
	if name == "" {
		if isCommand {
			name = "lib"
		} else {
			name = pkgName
		}
	} else if isCommand {
		name += "_lib"
	}
	return name
}

// testNameByConvention returns a suitable name for a go_test using the given
// naming convention and the import path.
func testNameByConvention(nc namingConvention, imp string) string {
	if nc == goDefaultLibraryNamingConvention {
		return defaultTestName
	}
	libName := libNameFromImportPath(imp)
	if libName == "" {
		libName = "lib"
	}
	return libName + "_test"
}

// testNameFromSingleSource returns a suitable name for a go_test using the
// single Go source file name.
func testNameFromSingleSource(src string) string {
	if i := strings.LastIndexByte(src, '.'); i >= 0 {
		src = src[0:i]
	}
	libName := libNameFromImportPath(src)
	if libName == "" {
		return ""
	}
	if strings.HasSuffix(libName, "_test") {
		return libName
	}
	return libName + "_test"
}

// binName returns a suitable name for a go_binary.
func binName(rel, prefix, repoRoot string) string {
	return pathtools.RelBaseName(rel, prefix, repoRoot)
}

func InferImportPath(c *config.Config, rel string) string {
	gc := getGoConfig(c)
	if rel == gc.prefixRel {
		return gc.prefix
	} else {
		fromPrefixRel := strings.TrimPrefix(rel, gc.prefixRel+"/")
		return path.Join(gc.prefix, fromPrefixRel)
	}
}

func goProtoPackageName(pkg proto.Package) string {
	if value, ok := pkg.Options["go_package"]; ok {
		if strings.LastIndexByte(value, '/') == -1 {
			return value
		} else {
			if i := strings.LastIndexByte(value, ';'); i != -1 {
				return value[i+1:]
			} else {
				return path.Base(value)
			}
		}
	}
	return strings.Replace(pkg.Name, ".", "_", -1)
}

func goProtoImportPath(c *config.Config, pkg proto.Package, rel string) string {
	if value, ok := pkg.Options["go_package"]; ok {
		if strings.LastIndexByte(value, '/') == -1 {
			return InferImportPath(c, rel)
		} else if i := strings.LastIndexByte(value, ';'); i != -1 {
			return value[:i]
		} else {
			return value
		}
	}
	return InferImportPath(c, rel)
}

func (t *goTarget) addFile(c *config.Config, er *embedResolver, info fileInfo) {
	t.cgo = t.cgo || info.isCgo
	add := getPlatformStringsAddFunction(c, info, nil)
	add(&t.sources, info.name)
	add(&t.imports, info.imports...)
	if er != nil {
		for _, embed := range info.embeds {
			embedSrcs, err := er.resolve(embed)
			if err != nil {
				log.Print(err)
				continue
			}
			add(&t.embedSrcs, embedSrcs...)
		}
	}
	for _, cppopts := range info.cppopts {
		optAdd := add
		if !cppopts.empty() {
			optAdd = getPlatformStringsAddFunction(c, info, cppopts)
		}
		optAdd(&t.cppopts, cppopts.opts)
	}
	for _, copts := range info.copts {
		optAdd := add
		if !copts.empty() {
			optAdd = getPlatformStringsAddFunction(c, info, copts)
		}
		optAdd(&t.copts, copts.opts)
	}
	for _, cxxopts := range info.cxxopts {
		optAdd := add
		if !cxxopts.empty() {
			optAdd = getPlatformStringsAddFunction(c, info, cxxopts)
		}
		optAdd(&t.cxxopts, cxxopts.opts)
	}
	for _, clinkopts := range info.clinkopts {
		optAdd := add
		if !clinkopts.empty() {
			optAdd = getPlatformStringsAddFunction(c, info, clinkopts)
		}
		optAdd(&t.clinkopts, clinkopts.opts)
	}
}

func protoTargetFromProtoPackage(name string, pkg proto.Package) protoTarget {
	target := protoTarget{name: name}
	for f := range pkg.Files {
		target.sources.addGenericString(f)
	}
	for i := range pkg.Imports {
		target.imports.addGenericString(i)
	}
	target.hasServices = pkg.HasServices
	return target
}

func (t *protoTarget) addFile(info fileInfo) {
	t.sources.addGenericString(info.name)
	for _, imp := range info.imports {
		t.imports.addGenericString(imp)
	}
	t.hasServices = t.hasServices || info.hasServices
}

// getPlatformStringsAddFunction returns a function used to add strings to
// a *platformStringsBuilder under the same set of constraints. This is a
// performance optimization to avoid evaluating constraints repeatedly.
func getPlatformStringsAddFunction(c *config.Config, info fileInfo, cgoTags *cgoTagsAndOpts) func(sb *platformStringsBuilder, ss ...string) {
	isOSSpecific, isArchSpecific := isOSArchSpecific(info, cgoTags)
	v := getGoConfig(c).rulesGoVersion
	constraintPrefix := "@" + getGoConfig(c).rulesGoRepoName + "//go/platform:"

	switch {
	case !isOSSpecific && !isArchSpecific:
		if checkConstraints(c, "", "", info.goos, info.goarch, info.tags, cgoTags) {
			return func(sb *platformStringsBuilder, ss ...string) {
				for _, s := range ss {
					sb.addGenericString(s)
				}
			}
		}

	case isOSSpecific && !isArchSpecific:
		var osMatch []string
		for _, os := range rule.KnownOSs {
			if rulesGoSupportsOS(v, os) &&
				checkConstraints(c, os, "", info.goos, info.goarch, info.tags, cgoTags) {
				osMatch = append(osMatch, os)
			}
		}
		if len(osMatch) > 0 {
			return func(sb *platformStringsBuilder, ss ...string) {
				for _, s := range ss {
					sb.addOSString(s, osMatch, constraintPrefix)
				}
			}
		}

	case !isOSSpecific && isArchSpecific:
		var archMatch []string
		for _, arch := range rule.KnownArchs {
			if rulesGoSupportsArch(v, arch) &&
				checkConstraints(c, "", arch, info.goos, info.goarch, info.tags, cgoTags) {
				archMatch = append(archMatch, arch)
			}
		}
		if len(archMatch) > 0 {
			return func(sb *platformStringsBuilder, ss ...string) {
				for _, s := range ss {
					sb.addArchString(s, archMatch, constraintPrefix)
				}
			}
		}

	default:
		var platformMatch []rule.Platform
		for _, platform := range rule.KnownPlatforms {
			if rulesGoSupportsPlatform(v, platform) &&
				checkConstraints(c, platform.OS, platform.Arch, info.goos, info.goarch, info.tags, cgoTags) {
				platformMatch = append(platformMatch, platform)
			}
		}
		if len(platformMatch) > 0 {
			return func(sb *platformStringsBuilder, ss ...string) {
				for _, s := range ss {
					sb.addPlatformString(s, platformMatch, constraintPrefix)
				}
			}
		}
	}

	return func(_ *platformStringsBuilder, _ ...string) {}
}

func (sb *platformStringsBuilder) isEmpty() bool {
	return sb.strs == nil
}

func (sb *platformStringsBuilder) hasGo() bool {
	for s := range sb.strs {
		if strings.HasSuffix(s, ".go") {
			return true
		}
	}
	return false
}

func (sb *platformStringsBuilder) addGenericString(s string) {
	if sb.strs == nil {
		sb.strs = make(map[string]platformStringInfo)
	}
	sb.strs[s] = platformStringInfo{set: genericSet}
}

func (sb *platformStringsBuilder) addOSString(s string, oss []string, constraintPrefix string) {
	if sb.strs == nil {
		sb.strs = make(map[string]platformStringInfo)
	}
	si, ok := sb.strs[s]
	if !ok {
		si.set = osSet
		si.osConstraints = make(map[string]bool)
	}
	switch si.set {
	case genericSet:
		return
	case osSet:
		for _, os := range oss {
			si.osConstraints[constraintPrefix+os] = true
		}
	default:
		si.convertToPlatforms(constraintPrefix)
		for _, os := range oss {
			for _, arch := range rule.KnownOSArchs[os] {
				si.platformConstraints[rule.PlatformConstraint{
					Platform:         rule.Platform{OS: os, Arch: arch},
					ConstraintPrefix: constraintPrefix,
				}] = true
			}
		}
	}
	sb.strs[s] = si
}

func (sb *platformStringsBuilder) addArchString(s string, archs []string, constraintPrefix string) {
	if sb.strs == nil {
		sb.strs = make(map[string]platformStringInfo)
	}
	si, ok := sb.strs[s]
	if !ok {
		si.set = archSet
		si.archConstraints = make(map[string]bool)
	}
	switch si.set {
	case genericSet:
		return
	case archSet:
		for _, arch := range archs {
			si.archConstraints[constraintPrefix+arch] = true
		}
	default:
		si.convertToPlatforms(constraintPrefix)
		for _, arch := range archs {
			for _, os := range rule.KnownArchOSs[arch] {
				si.platformConstraints[rule.PlatformConstraint{
					Platform:         rule.Platform{OS: os, Arch: arch},
					ConstraintPrefix: constraintPrefix,
				}] = true
			}
		}
	}
	sb.strs[s] = si
}

func (sb *platformStringsBuilder) addPlatformString(s string, platforms []rule.Platform, constraintPrefix string) {
	if sb.strs == nil {
		sb.strs = make(map[string]platformStringInfo)
	}
	si, ok := sb.strs[s]
	if !ok {
		si.set = platformSet
		si.platformConstraints = make(map[rule.PlatformConstraint]bool)
	}
	switch si.set {
	case genericSet:
		return
	default:
		si.convertToPlatforms(constraintPrefix)
		for _, p := range platforms {
			pConstraint := rule.PlatformConstraint{Platform: rule.Platform{OS: p.OS, Arch: p.Arch}, ConstraintPrefix: constraintPrefix}
			si.platformConstraints[pConstraint] = true
		}
	}
	sb.strs[s] = si
}

func (sb *platformStringsBuilder) build() rule.PlatformStrings {
	var ps rule.PlatformStrings
	for s, si := range sb.strs {
		switch si.set {
		case genericSet:
			ps.Generic = append(ps.Generic, s)
		case osSet:
			if ps.OS == nil {
				ps.OS = make(map[string][]string)
			}
			for os := range si.osConstraints {
				ps.OS[os] = append(ps.OS[os], s)
			}
		case archSet:
			if ps.Arch == nil {
				ps.Arch = make(map[string][]string)
			}
			for arch := range si.archConstraints {
				ps.Arch[arch] = append(ps.Arch[arch], s)
			}
		case platformSet:
			if ps.Platform == nil {
				ps.Platform = make(map[rule.PlatformConstraint][]string)
			}
			for p := range si.platformConstraints {
				ps.Platform[p] = append(ps.Platform[p], s)
			}
		}
	}
	sort.Strings(ps.Generic)
	if ps.OS != nil {
		for _, ss := range ps.OS {
			sort.Strings(ss)
		}
	}
	if ps.Arch != nil {
		for _, ss := range ps.Arch {
			sort.Strings(ss)
		}
	}
	if ps.Platform != nil {
		for _, ss := range ps.Platform {
			sort.Strings(ss)
		}
	}
	return ps
}

func (sb *platformStringsBuilder) buildFlat() []string {
	strs := make([]string, 0, len(sb.strs))
	for s := range sb.strs {
		strs = append(strs, s)
	}
	sort.Strings(strs)
	return strs
}

func (si *platformStringInfo) convertToPlatforms(constraintPrefix string) {
	switch si.set {
	case genericSet:
		log.Panic("cannot convert generic string to platformConstraints")
	case platformSet:
		return
	case osSet:
		si.set = platformSet
		si.platformConstraints = make(map[rule.PlatformConstraint]bool)
		for osConstraint := range si.osConstraints {
			os := strings.TrimPrefix(osConstraint, constraintPrefix)
			for _, arch := range rule.KnownOSArchs[os] {
				si.platformConstraints[rule.PlatformConstraint{
					Platform:         rule.Platform{OS: os, Arch: arch},
					ConstraintPrefix: constraintPrefix,
				}] = true
			}
		}
		si.osConstraints = nil
	case archSet:
		si.set = platformSet
		si.platformConstraints = make(map[rule.PlatformConstraint]bool)
		for archConstraint := range si.archConstraints {
			arch := strings.TrimPrefix(archConstraint, constraintPrefix)
			for _, os := range rule.KnownArchOSs[arch] {
				si.platformConstraints[rule.PlatformConstraint{
					Platform:         rule.Platform{OS: os, Arch: arch},
					ConstraintPrefix: constraintPrefix,
				}] = true
			}
		}
		si.archConstraints = nil
	}
}

var semverRex = regexp.MustCompile(`^.*?(/v\d+)(?:/.*)?$`)

// pathWithoutSemver removes a semantic version suffix from path.
// For example, if path is "example.com/foo/v2/bar", pathWithoutSemver
// will return "example.com/foo/bar". If there is no semantic version suffix,
// "" will be returned.
func pathWithoutSemver(path string) string {
	m := semverRex.FindStringSubmatchIndex(path)
	if m == nil {
		return ""
	}
	v := path[m[2]+2 : m[3]]
	if v[0] == '0' || v == "1" {
		return ""
	}
	return path[:m[2]] + path[m[3]:]
}
