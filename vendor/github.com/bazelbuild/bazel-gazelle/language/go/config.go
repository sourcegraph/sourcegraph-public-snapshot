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
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	gzflag "github.com/bazelbuild/bazel-gazelle/flag"
	"github.com/bazelbuild/bazel-gazelle/internal/module"
	"github.com/bazelbuild/bazel-gazelle/internal/version"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/rule"
	bzl "github.com/bazelbuild/buildtools/build"
	"golang.org/x/mod/modfile"
)

var minimumRulesGoVersion = version.Version{0, 29, 0}

// goConfig contains configuration values related to Go rules.
type goConfig struct {
	// The name under which the rules_go repository can be referenced from the
	// repository in which Gazelle is running.
	rulesGoRepoName string

	// rulesGoVersion is the version of io_bazel_rules_go being used. Determined
	// by reading go/def.bzl. May be unset if the version can't be read.
	rulesGoVersion version.Version

	// genericTags is a set of tags that Gazelle considers to be true. Set with
	// -build_tags or # gazelle:build_tags. Some tags, like gc, are always on.
	genericTags map[string]bool

	// prefix is a prefix of an import path, used to generate importpath
	// attributes. Set with -go_prefix or # gazelle:prefix.
	prefix string

	// prefixRel is the package name of the directory where the prefix was set
	// ("" for the root directory).
	prefixRel string

	// prefixSet indicates whether the prefix was set explicitly. It is an error
	// to infer an importpath for a rule without setting the prefix.
	prefixSet bool

	// importMapPrefix is a prefix of a package path, used to generate importmap
	// attributes. Set with # gazelle:importmap_prefix.
	importMapPrefix string

	// importMapPrefixRel is the package name of the directory where importMapPrefix
	// was set ("" for the root directory).
	importMapPrefixRel string

	// depMode determines how imports that are not standard, indexed, or local
	// (under the current prefix) should be resolved.
	depMode dependencyMode

	// goGenerateProto indicates whether to generate go_proto_library
	goGenerateProto bool

	// goNamingConvention controls the name of generated targets
	goNamingConvention namingConvention

	// goNamingConventionExternal controls the default naming convention for
	// imports in external repositories with unknown naming conventions.
	goNamingConventionExternal namingConvention

	// goProtoCompilers is the protocol buffers compiler(s) to use for go code.
	goProtoCompilers []string

	// goProtoCompilersSet indicates whether goProtoCompiler was set explicitly.
	goProtoCompilersSet bool

	// goGrpcCompilers is the gRPC compiler(s) to use for go code.
	goGrpcCompilers []string

	// goGrpcCompilersSet indicates whether goGrpcCompiler was set explicitly.
	goGrpcCompilersSet bool

	// goRepositoryMode is true if Gazelle was invoked by a go_repository rule.
	// In this mode, we won't go out to the network to resolve external deps.
	goRepositoryMode bool

	// By default, internal packages are only visible to its siblings.
	// goVisibility adds a list of packages the internal packages should be
	// visible to
	goVisibility []string

	// moduleMode is true if the current directory is intended to be built
	// as part of a module. Minimal module compatibility won't be supported
	// if this is true in the root directory. External dependencies may be
	// resolved differently (also depending on goRepositoryMode).
	moduleMode bool

	// map between external repo names and their `build_naming_convention`
	// attribute.
	repoNamingConvention map[string]namingConvention

	// submodules is a list of modules which have the current module's path
	// as a prefix of their own path. This affects visibility attributes
	// in internal packages.
	submodules []moduleRepo

	// testMode determines how go_test targets are generated.
	testMode testMode

	// buildDirectives, buildExternalAttr, buildExtraArgsAttr,
	// buildFileGenerationAttr, buildFileNamesAttr, buildFileProtoModeAttr and
	// buildTagsAttr are attributes for go_repository rules, set on the command
	// line.
	buildDirectivesAttr, buildExternalAttr, buildExtraArgsAttr, buildFileGenerationAttr, buildFileNamesAttr, buildFileProtoModeAttr, buildTagsAttr string
}

// testMode determines how go_test rules are generated.
type testMode int

const (
	// defaultTestMode generates a go_test for the primary package in a directory.
	defaultTestMode = iota

	// fileTestMode generates a go_test for each Go test file.
	fileTestMode
)

var (
	defaultGoProtoCompilers = []string{"@io_bazel_rules_go//proto:go_proto"}
	defaultGoGrpcCompilers  = []string{"@io_bazel_rules_go//proto:go_grpc"}
)

func (m testMode) String() string {
	switch m {
	case defaultTestMode:
		return "default"
	case fileTestMode:
		return "file"
	default:
		return "unknown"
	}
}

func testModeFromString(s string) (testMode, error) {
	switch s {
	case "default":
		return defaultTestMode, nil
	case "file":
		return fileTestMode, nil
	default:
		return 0, fmt.Errorf("unrecognized go_test mode: %q", s)
	}
}

func newGoConfig() *goConfig {
	gc := &goConfig{
		goProtoCompilers: defaultGoProtoCompilers,
		goGrpcCompilers:  defaultGoGrpcCompilers,
		goGenerateProto:  true,
	}
	gc.preprocessTags()
	return gc
}

func getGoConfig(c *config.Config) *goConfig {
	return c.Exts[goName].(*goConfig)
}

func (gc *goConfig) clone() *goConfig {
	gcCopy := *gc
	gcCopy.genericTags = make(map[string]bool)
	for k, v := range gc.genericTags {
		gcCopy.genericTags[k] = v
	}
	gcCopy.goProtoCompilers = gc.goProtoCompilers[:len(gc.goProtoCompilers):len(gc.goProtoCompilers)]
	gcCopy.goGrpcCompilers = gc.goGrpcCompilers[:len(gc.goGrpcCompilers):len(gc.goGrpcCompilers)]
	gcCopy.submodules = gc.submodules[:len(gc.submodules):len(gc.submodules)]
	return &gcCopy
}

// preprocessTags adds some tags which are on by default before they are
// used to match files.
func (gc *goConfig) preprocessTags() {
	if gc.genericTags == nil {
		gc.genericTags = make(map[string]bool)
	}
	gc.genericTags["gc"] = true
}

// setBuildTags sets genericTags by parsing as a comma separated list. An
// error will be returned for tags that wouldn't be recognized by "go build".
// preprocessTags should be called before this.
func (gc *goConfig) setBuildTags(tags string) error {
	if tags == "" {
		return nil
	}
	for _, t := range strings.Split(tags, ",") {
		if strings.HasPrefix(t, "!") {
			return fmt.Errorf("build tags can't be negated: %s", t)
		}
		gc.genericTags[t] = true
	}
	return nil
}

func getProtoMode(c *config.Config) proto.Mode {
	if gc := getGoConfig(c); !gc.goGenerateProto {
		return proto.DisableMode
	} else if pc := proto.GetProtoConfig(c); pc != nil {
		return pc.Mode
	} else {
		return proto.DisableGlobalMode
	}
}

// dependencyMode determines how imports of packages outside of the prefix
// are resolved.
type dependencyMode int

const (
	// externalMode indicates imports should be resolved to external dependencies
	// (declared in WORKSPACE). Calls out to the network if an import can't be resolved
	// locally.
	externalMode dependencyMode = iota

	// staticMode indicates imports should be resolved only to dependencies known by
	// Gazelle (declared in WORKSPACE). Unknown imports are ignored.
	staticMode

	// vendorMode indicates imports should be resolved to libraries in the
	// vendor directory.
	vendorMode
)

func (m dependencyMode) String() string {
	switch m {
	case externalMode:
		return "external"
	case staticMode:
		return "static"
	case vendorMode:
		return "vendor"
	}
	return ""
}

type externalFlag struct {
	depMode *dependencyMode
}

func (f *externalFlag) Set(value string) error {
	switch value {
	case "external":
		*f.depMode = externalMode
	case "static":
		*f.depMode = staticMode
	case "vendored":
		*f.depMode = vendorMode
	default:
		return fmt.Errorf("unrecognized dependency mode: %q", value)
	}
	return nil
}

func (f *externalFlag) String() string {
	if f == nil || f.depMode == nil {
		return "external"
	}
	return f.depMode.String()
}

type tagsFlag func(string) error

func (f tagsFlag) Set(value string) error {
	return f(value)
}

func (f tagsFlag) String() string {
	return ""
}

type namingConventionFlag struct {
	nc *namingConvention
}

func (f namingConventionFlag) Set(value string) error {
	if nc, err := namingConventionFromString(value); err != nil {
		return err
	} else {
		*f.nc = nc
		return nil
	}
}

func (f *namingConventionFlag) String() string {
	if f == nil || f.nc == nil {
		return "naming_convention"
	}
	return f.nc.String()
}

// namingConvention determines how go targets are named.
type namingConvention int

const (
	// Try to detect the naming convention in use.
	unknownNamingConvention namingConvention = iota

	// 'go_default_library' and 'go_default_test'
	goDefaultLibraryNamingConvention

	// For an import path that ends with foo, the go_library rules target is
	// named 'foo', the go_test is named 'foo_test'.
	// For a main package, the go_binary takes the 'foo' name, the library
	// is named 'foo_lib', and the go_test is named 'foo_test'.
	importNamingConvention

	// Same as importNamingConvention, but generate alias rules for libraries that have
	// the legacy 'go_default_library' name.
	importAliasNamingConvention
)

func (nc namingConvention) String() string {
	switch nc {
	case goDefaultLibraryNamingConvention:
		return "go_default_library"
	case importNamingConvention:
		return "import"
	case importAliasNamingConvention:
		return "import_alias"
	}
	return ""
}

func namingConventionFromString(s string) (namingConvention, error) {
	switch s {
	case "":
		return unknownNamingConvention, nil
	case "go_default_library":
		return goDefaultLibraryNamingConvention, nil
	case "import":
		return importNamingConvention, nil
	case "import_alias":
		return importAliasNamingConvention, nil
	default:
		return unknownNamingConvention, fmt.Errorf("unknown naming convention %q", s)
	}
}

type moduleRepo struct {
	repoName, modulePath string
}

var (
	validBuildExternalAttr       = []string{"external", "vendored"}
	validBuildFileGenerationAttr = []string{"auto", "on", "off"}
	validBuildFileProtoModeAttr  = []string{"default", "legacy", "disable", "disable_global", "package"}
)

func (*goLang) KnownDirectives() []string {
	return []string{
		"build_tags",
		"go_generate_proto",
		"go_grpc_compilers",
		"go_naming_convention",
		"go_naming_convention_external",
		"go_proto_compilers",
		"go_test",
		"go_visibility",
		"importmap_prefix",
		"prefix",
	}
}

func (*goLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
	gc := newGoConfig()
	switch cmd {
	case "fix", "update":
		fs.Var(
			tagsFlag(gc.setBuildTags),
			"build_tags",
			"comma-separated list of build tags. If not specified, Gazelle will not\n\tfilter sources with build constraints.")
		fs.Var(
			&gzflag.ExplicitFlag{Value: &gc.prefix, IsSet: &gc.prefixSet},
			"go_prefix",
			"prefix of import paths in the current workspace")
		fs.Var(
			&externalFlag{&gc.depMode},
			"external",
			"external: resolve external packages with go_repository\n\tvendored: resolve external packages as packages in vendor/")
		fs.Var(
			&gzflag.MultiFlag{Values: &gc.goProtoCompilers, IsSet: &gc.goProtoCompilersSet},
			"go_proto_compiler",
			"go_proto_library compiler to use (may be repeated)")
		fs.Var(
			&gzflag.MultiFlag{Values: &gc.goGrpcCompilers, IsSet: &gc.goGrpcCompilersSet},
			"go_grpc_compiler",
			"go_proto_library compiler to use for gRPC (may be repeated)")
		fs.BoolVar(
			&gc.goRepositoryMode,
			"go_repository_mode",
			false,
			"set when gazelle is invoked by go_repository")
		fs.BoolVar(
			&gc.moduleMode,
			"go_repository_module_mode",
			false,
			"set when gazelle is invoked by go_repository in module mode")
		fs.Var(
			&namingConventionFlag{&gc.goNamingConvention},
			"go_naming_convention",
			"controls generated library names. One of (go_default_library, import, import_alias)")
		fs.Var(
			&namingConventionFlag{&gc.goNamingConventionExternal},
			"go_naming_convention_external",
			"controls naming convention used when resolving libraries in external repositories with unknown conventions")

	case "update-repos":
		fs.StringVar(&gc.buildDirectivesAttr,
			"build_directives",
			"",
			"Sets the build_directives attribute for the generated go_repository rule(s).")
		fs.Var(&gzflag.AllowedStringFlag{Value: &gc.buildExternalAttr, Allowed: validBuildExternalAttr},
			"build_external",
			"Sets the build_external attribute for the generated go_repository rule(s).")
		fs.StringVar(&gc.buildExtraArgsAttr,
			"build_extra_args",
			"",
			"Sets the build_extra_args attribute for the generated go_repository rule(s).")
		fs.Var(&gzflag.AllowedStringFlag{Value: &gc.buildFileGenerationAttr, Allowed: validBuildFileGenerationAttr},
			"build_file_generation",
			"Sets the build_file_generation attribute for the generated go_repository rule(s).")
		fs.StringVar(&gc.buildFileNamesAttr,
			"build_file_names",
			"",
			"Sets the build_file_name attribute for the generated go_repository rule(s).")
		fs.Var(&gzflag.AllowedStringFlag{Value: &gc.buildFileProtoModeAttr, Allowed: validBuildFileProtoModeAttr},
			"build_file_proto_mode",
			"Sets the build_file_proto_mode attribute for the generated go_repository rule(s).")
		fs.StringVar(&gc.buildTagsAttr,
			"build_tags",
			"",
			"Sets the build_tags attribute for the generated go_repository rule(s).")
	}
	c.Exts[goName] = gc
}

func (*goLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	// The base of the -go_prefix flag may be used to generate proto_library
	// rule names when there are no .proto sources (empty rules to be deleted)
	// or when the package name can't be determined.
	// TODO(jayconrod): deprecate and remove this behavior.
	gc := getGoConfig(c)
	if pc := proto.GetProtoConfig(c); pc != nil {
		pc.GoPrefix = gc.prefix
	}

	// List modules that may refer to internal packages in this module.
	for _, r := range c.Repos {
		if r.Kind() != "go_repository" {
			continue
		}
		modulePath := r.AttrString("importpath")
		if !strings.HasPrefix(modulePath, gc.prefix+"/") {
			continue
		}
		m := moduleRepo{
			repoName:   r.Name(),
			modulePath: modulePath,
		}
		gc.submodules = append(gc.submodules, m)
	}

	return nil
}

func (*goLang) Configure(c *config.Config, rel string, f *rule.File) {
	var gc *goConfig
	if raw, ok := c.Exts[goName]; !ok {
		gc = newGoConfig()
	} else {
		gc = raw.(*goConfig).clone()
	}
	c.Exts[goName] = gc

	if rel == "" {
		moduleToApparentName, err := module.ExtractModuleToApparentNameMapping(c.RepoRoot)
		if err != nil {
			log.Print(err)
		} else {
			gc.rulesGoRepoName = moduleToApparentName("rules_go")
		}
		if gc.rulesGoRepoName == "" {
			// The legacy name used in WORKSPACE.
			gc.rulesGoRepoName = "io_bazel_rules_go"
		}

		const message = `Gazelle may not be compatible with this version of rules_go.
Update io_bazel_rules_go to a newer version in your WORKSPACE file.`
		gc.rulesGoVersion, err = findRulesGoVersion(c)
		if c.ShouldFix {
			// Only check the version when "fix" is run. Generated build files
			// frequently work with older version of rules_go, and we don't want to
			// nag too much since there's no way to disable this warning.
			// Also, don't print a warning if the rules_go repo hasn't been fetched,
			// since that's a common issue when Gazelle is run as a separate binary.
			if err != nil && err != errRulesGoRepoNotFound && c.ShouldFix {
				log.Printf("%v\n%s", err, message)
			} else if err == nil && gc.rulesGoVersion.Compare(minimumRulesGoVersion) < 0 {
				log.Printf("Found RULES_GO_VERSION %s. Minimum compatible version is %s.\n%s", gc.rulesGoVersion, minimumRulesGoVersion, message)
			}
		}
		repoNamingConvention := map[string]namingConvention{}
		for _, repo := range c.Repos {
			if repo.Kind() == "go_repository" {
				if attr := repo.AttrString("build_naming_convention"); attr == "" {
					// No naming convention specified.
					// go_repsitory uses importAliasNamingConvention by default, so we
					// could use whichever name.
					// resolveExternal should take that as a signal to follow the current
					// naming convention to avoid churn.
					repoNamingConvention[repo.Name()] = importAliasNamingConvention
				} else if nc, err := namingConventionFromString(attr); err != nil {
					log.Printf("in go_repository named %q: %v", repo.Name(), err)
				} else {
					repoNamingConvention[repo.Name()] = nc
				}
			}
		}
		gc.repoNamingConvention = repoNamingConvention
	}

	if !gc.moduleMode {
		st, err := os.Stat(filepath.Join(c.RepoRoot, filepath.FromSlash(rel), "go.mod"))
		if err == nil && !st.IsDir() {
			gc.moduleMode = true
		}
	}

	if path.Base(rel) == "vendor" {
		gc.importMapPrefix = InferImportPath(c, rel)
		gc.importMapPrefixRel = rel
		gc.prefix = ""
		gc.prefixRel = rel
	}

	if f != nil {
		setPrefix := func(prefix string) {
			if err := checkPrefix(prefix); err != nil {
				log.Print(err)
				return
			}
			gc.prefix = prefix
			gc.prefixSet = true
			gc.prefixRel = rel
		}
		for _, d := range f.Directives {
			switch d.Key {
			case "build_tags":
				if err := gc.setBuildTags(d.Value); err != nil {
					log.Print(err)
					continue
				}
				gc.preprocessTags()
				if err := gc.setBuildTags(d.Value); err != nil {
					log.Print(err)
				}

			case "go_generate_proto":
				if goGenerateProto, err := strconv.ParseBool(d.Value); err == nil {
					gc.goGenerateProto = goGenerateProto
				} else {
					log.Printf("parsing go_generate_proto: %v", err)
				}

			case "go_naming_convention":
				if nc, err := namingConventionFromString(d.Value); err == nil {
					gc.goNamingConvention = nc
				} else {
					log.Print(err)
				}

			case "go_naming_convention_external":
				if nc, err := namingConventionFromString(d.Value); err == nil {
					gc.goNamingConventionExternal = nc
				} else {
					log.Print(err)
				}

			case "go_grpc_compilers":
				// Special syntax (empty value) to reset directive.
				if d.Value == "" {
					gc.goGrpcCompilersSet = false
					gc.goGrpcCompilers = defaultGoGrpcCompilers
				} else {
					gc.goGrpcCompilersSet = true
					gc.goGrpcCompilers = splitValue(d.Value)
				}

			case "go_proto_compilers":
				// Special syntax (empty value) to reset directive.
				if d.Value == "" {
					gc.goProtoCompilersSet = false
					gc.goProtoCompilers = defaultGoProtoCompilers
				} else {
					gc.goProtoCompilersSet = true
					gc.goProtoCompilers = splitValue(d.Value)
				}

			case "go_test":
				mode, err := testModeFromString(d.Value)
				if err != nil {
					log.Print(err)
					continue
				}
				gc.testMode = mode

			case "go_visibility":
				gc.goVisibility = append(gc.goVisibility, strings.TrimSpace(d.Value))

			case "importmap_prefix":
				gc.importMapPrefix = d.Value
				gc.importMapPrefixRel = rel

			case "prefix":
				setPrefix(d.Value)
			}
		}

		if !gc.prefixSet {
			for _, r := range f.Rules {
				switch r.Kind() {
				case "go_prefix":
					args := r.Args()
					if len(args) != 1 {
						continue
					}
					s, ok := args[0].(*bzl.StringExpr)
					if !ok {
						continue
					}
					setPrefix(s.Value)

				case "gazelle":
					if prefix := r.AttrString("prefix"); prefix != "" {
						setPrefix(prefix)
					}
				}
			}
		}
		if !gc.prefixSet {
			// Parse the module directive out of the go.mod file, if present.
			goModPath := filepath.Join(c.RepoRoot, filepath.FromSlash(rel), "go.mod")
			goMod, err := os.ReadFile(goModPath)
			// Reading the go.mod file is best-effort and may fail for various reasons, such as
			// the file not existing or being a directory. Do not report errors.
			if err == nil {
				goModFile, err := modfile.ParseLax(goModPath, goMod, nil)
				// If the go.mod file exists but is malformed, report the error.
				if err != nil {
					log.Printf("parsing %s: %s", goModPath, err)
				} else {
					setPrefix(goModFile.Module.Mod.Path)
				}
			}
		}
	}

	if gc.goNamingConvention == unknownNamingConvention {
		gc.goNamingConvention = detectNamingConvention(c, f)
	}
}

// checkPrefix checks that a string may be used as a prefix. We forbid local
// (relative) imports and those beginning with "/". We allow the empty string,
// but generated rules must not have an empty importpath.
func checkPrefix(prefix string) error {
	if strings.HasPrefix(prefix, "/") || build.IsLocalImport(prefix) {
		return fmt.Errorf("invalid prefix: %q", prefix)
	}
	return nil
}

// splitDirective splits a comma-separated directive value into its component
// parts, trimming each of any whitespace characters.
func splitValue(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		values = append(values, strings.TrimSpace(part))
	}
	return values
}

// findRulesGoVersion attempts to infer the version of io_bazel_rules_go.
// It can read the external directory (if bazel has fetched it), or it can
// read WORKSPACE. Neither method is completely reliable.
func findRulesGoVersion(c *config.Config) (version.Version, error) {
	const message = `Gazelle may not be compatible with this version of rules_go.
Update io_bazel_rules_go to a newer version in your WORKSPACE file.`

	var vstr string
	if rulesGoPath, err := repo.FindExternalRepo(c.RepoRoot, config.RulesGoRepoName); err == nil {
		// Bazel has already fetched io_bazel_rules_go. We can read its version
		// from //go:def.bzl.
		defBzlPath := filepath.Join(rulesGoPath, "go", "def.bzl")
		defBzlContent, err := ioutil.ReadFile(defBzlPath)
		if err != nil {
			return nil, err
		}
		versionRe := regexp.MustCompile(`(?m)^RULES_GO_VERSION = ['"]([0-9.]*)['"]`)
		match := versionRe.FindSubmatch(defBzlContent)
		if match == nil {
			return nil, fmt.Errorf("RULES_GO_VERSION not found in @%s//go:def.bzl.\n%s", config.RulesGoRepoName, message)
		}
		vstr = string(match[1])
	} else {
		// Bazel has not fetched io_bazel_rules_go. Maybe we can find it in the
		// WORKSPACE file.
		re := regexp.MustCompile(`github\.com/bazelbuild/rules_go/releases/download/v([0-9.]+)/`)
	RepoLoop:
		for _, r := range c.Repos {
			if r.Kind() == "http_archive" && r.Name() == "io_bazel_rules_go" {
				for _, u := range r.AttrStrings("urls") {
					if m := re.FindStringSubmatch(u); m != nil {
						vstr = m[1]
						break RepoLoop
					}
				}
			}
		}
	}

	if vstr == "" {
		// Couldn't find a version. We return a specific value since this is not
		// usually a useful error to report.
		return nil, errRulesGoRepoNotFound
	}

	return version.ParseVersion(vstr)
}

var errRulesGoRepoNotFound = errors.New(config.RulesGoRepoName + " external repository not found")

// detectNamingConvention attempts to detect the naming convention in use by
// reading build files in subdirectories of the repository root directory.
//
// If detectNamingConvention can't detect the naming convention (for example,
// because no build files are found or multiple naming conventions are found),
// importNamingConvention is returned.
func detectNamingConvention(c *config.Config, rootFile *rule.File) namingConvention {
	if !c.IndexLibraries {
		// Indexing is disabled, which usually means speed is important and I/O
		// should be minimized. Let's not open extra files or directories.
		return importNamingConvention
	}

	detectInFile := func(f *rule.File) namingConvention {
		for _, r := range f.Rules {
			// NOTE: map_kind is not supported. c.KindMap will not be accurate in
			// subdirectories.
			kind := r.Kind()
			name := r.Name()
			if kind != "alias" && name == defaultLibName {
				// Assume any kind of rule with the name "go_default_library" is some
				// kind of go library. The old version of go_proto_library used this
				// name, and it's possible with map_kind as well.
				return goDefaultLibraryNamingConvention
			} else if isGoLibrary(kind) && name == path.Base(r.AttrString("importpath")) {
				return importNamingConvention
			}
		}
		return unknownNamingConvention
	}

	detectInDir := func(dir, rel string) namingConvention {
		var f *rule.File
		for _, name := range c.ValidBuildFileNames {
			fpath := filepath.Join(dir, name)
			data, err := ioutil.ReadFile(fpath)
			if err != nil {
				continue
			}
			f, err = rule.LoadData(fpath, rel, data)
			if err != nil {
				continue
			}
		}
		if f == nil {
			return unknownNamingConvention
		}
		return detectInFile(f)
	}

	nc := unknownNamingConvention
	if rootFile != nil {
		if rootNC := detectInFile(rootFile); rootNC != unknownNamingConvention {
			return rootNC
		}
	}

	infos, err := ioutil.ReadDir(c.RepoRoot)
	if err != nil {
		return importNamingConvention
	}
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		dirName := info.Name()
		dirNC := detectInDir(filepath.Join(c.RepoRoot, dirName), dirName)
		if dirNC == unknownNamingConvention {
			continue
		}
		if nc != unknownNamingConvention && dirNC != nc {
			// Subdirectories use different conventions. Return the default.
			return importNamingConvention
		}
		nc = dirNC
	}
	if nc == unknownNamingConvention {
		return importNamingConvention
	}
	return nc
}
