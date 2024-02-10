package gomockgengazelle

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	golang "github.com/bazelbuild/bazel-gazelle/language/go"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"golang.org/x/exp/maps"

	mockgenc "github.com/sourcegraph/sourcegraph/dev/go-mockgen-transformer/config"
)

type gomockgen struct {
	language.BaseLang
	language.BaseLifecycleManager
}

var (
	_ (language.Language)            = (*gomockgen)(nil)
	_ (language.LifecycleManager)    = (*gomockgen)(nil)
	_ (language.ModuleAwareLanguage) = (*gomockgen)(nil)
)

var (
	yamlPayload mockgenc.YamlPayload

	unvisitedDirs = make(map[string]bool)
	allOutputDirs = make(map[string]mockgenc.YamlMock)
	rootDir       string
	loadConfig    = sync.OnceValue[error](func() (err error) {
		yamlPayload, err = mockgenc.ReadManifest(filepath.Join(rootDir, "mockgen.yaml"))
		if err != nil {
			return err
		}

		for _, mock := range yamlPayload.Mocks {
			allOutputDirs[filepath.Dir(mock.Filename)] = mock
			unvisitedDirs[filepath.Dir(mock.Filename)] = true
		}
		return nil
	})
)

func NewLanguage() language.Language {
	return &gomockgen{}
}

func (*gomockgen) Name() string { return "gomockgen" }

func (*gomockgen) Configure(c *config.Config, rel string, f *rule.File) {
	// need to tell gazelle to to run generators with the following name (aka ours)
	c.Langs = append(c.Langs, "gomockgen")
}

func (*gomockgen) KnownDirectives() []string {
	return []string{
		"go_library",
		"go_mockgen",
	}
}

func (*gomockgen) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		"go_mockgen": {
			MatchAny: true,
			// I cant tell if these work or not...
			MergeableAttrs: map[string]bool{
				"deps":      true,
				"out":       true,
				"manifests": true,
			},
		},
		"multirun": {
			MatchAttrs: []string{"go_mockgen"},
			MergeableAttrs: map[string]bool{
				"commands": true,
			},
		},
	}
}

func (*gomockgen) ApparentLoads(moduleToApparentName func(string) string) []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name:    "//dev:go_mockgen.bzl",
			Symbols: []string{"go_mockgen"},
		},
		{
			Name:    "@rules_multirun//:defs.bzl",
			Symbols: []string{"multirun"},
		},
	}
}

// Deprecated: use ApparentLoads instead
func (*gomockgen) Loads() []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name:    "//dev:go_mockgen.bzl",
			Symbols: []string{"go_mockgen"},
		},
		{
			Name:    "@rules_multirun//:defs.bzl",
			Symbols: []string{"multirun"},
		},
	}
}

func (g *gomockgen) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	if args.Rel == "dev" {
		var targets []string
		for _, mock := range yamlPayload.Mocks {
			targets = append(targets, "//"+filepath.Dir(mock.Filename)+":generate_mocks")
		}

		slices.Sort(targets)
		targets = slices.Compact(targets)

		multirunRule := rule.NewRule("multirun", "go_mockgen")
		multirunRule.SetAttr("commands", targets)
		multirunRule.SetAttr("jobs", 1)

		return language.GenerateResult{
			Gen:     []*rule.Rule{multirunRule},
			Imports: []interface{}{nil},
		}
	}

	rootDir = args.Config.RepoRoot
	if err := loadConfig(); err != nil {
		panic(err)
	}

	mock, ok := allOutputDirs[args.Rel]
	if !ok {
		return language.GenerateResult{}
	}

	delete(unvisitedDirs, args.Rel)

	outputFilename := filepath.Base(mock.Filename)

	r := rule.NewRule("go_mockgen", "generate_mocks")
	r.SetAttr("out", outputFilename)

	manifests := []string{"//:mockgen.yaml"}
	for _, manifest := range yamlPayload.IncludeConfigPaths {
		manifests = append(manifests, "//:"+manifest)
	}
	r.SetAttr("manifests", manifests)

	goRuleIndex := slices.IndexFunc(args.OtherGen, func(r *rule.Rule) bool {
		if strings.HasSuffix(outputFilename, "_test.go") {
			return r.Kind() == "go_test"
		} else {
			return r.Kind() == "go_library"
		}
	})
	if goRuleIndex == -1 {
		panic(fmt.Sprintf("couldn't find a go_library rule in dir %q", args.Rel))
	}

	goRule := args.OtherGen[goRuleIndex]
	goRule.SetAttr("srcs", append(goRule.AttrStrings("srcs"), filepath.Base(mock.Filename)))

	imports := gatherDependencies(mock)

	return language.GenerateResult{
		Gen:     []*rule.Rule{r, goRule},
		Imports: []interface{}{imports, nil},
	}
}

func (g *gomockgen) DoneGeneratingRules() {
	if len(unvisitedDirs) > 0 {
		panic(fmt.Sprintf("Some declared mock output files were not created due to their output directory missing. Please create the following directories so : %v", maps.Keys(unvisitedDirs)))
	}
}

func (g *gomockgen) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, rawImports interface{}, from label.Label) {
	if r.Kind() != "go_mockgen" {
		return
	}
	imports := rawImports.([]string)

	r.DelAttr("deps")

	var labels []string
	for _, i := range imports {
		result, err := golang.ResolveGo(c, ix, rc, i, from)
		if err != nil {
			panic(err)
		}
		labels = append(labels, result.Rel(from.Repo, from.Pkg).String())
	}
	r.SetAttr("deps", labels)
}

func gatherDependencies(mock mockgenc.YamlMock) (deps []string) {
	if mock.Path != "" {
		deps = append(deps, mock.Path)
	}
	for _, source := range mock.Sources {
		deps = append(deps, source.Path)
	}
	return
}
