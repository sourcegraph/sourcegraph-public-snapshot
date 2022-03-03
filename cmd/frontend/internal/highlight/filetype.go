package highlight

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/grafana/regexp"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

type EngineType int64

const (
	EngineInvalid    EngineType = 0
	EngineTreeSitter            = 1
	EngineSyntect               = 2
)

type ftPattern struct {
	pattern  *regexp.Regexp
	filetype string
}

// TODO: Decide on capitalization, cause it's a nightmare otherwise.
// TODO: Validate that those are available filetypes, otherwise it's kind of
//       also pointless (for example, how to tell them that C# is not OK, but c_sharp is?)
type syntaxHighlightConfig struct {
	// Order does not matter. Evaluated before Patterns
	Extensions map[string]string

	// Order matters for this. First matching pattern matches.
	// Matches against the entire string.
	Patterns []ftPattern

	// TODO:
	// Shebang
}

type syntaxEngineConfig struct {
	Default   EngineType
	Overrides map[string]EngineType
}

type SyntaxEngineQuery struct {
	Engine           EngineType
	Filetype         string
	FiletypeOverride bool
}

var highlightConfig = syntaxHighlightConfig{}

var engineConfig = syntaxEngineConfig{
	// This sets the default syntax engine for the sourcegraph server.
	Default:   EngineSyntect,
	Overrides: map[string]EngineType{},
}

func init() {
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		highlights := c.SiteConfig().Highlights
		if highlights == nil {
			return
		}

		if _, ok := engineNameToEngineType(highlights.Engine.Default); !ok {
			problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid highlights.Engine.Default: `%s`.", highlights.Engine.Default)))
		}
		// TODO: Probably should validate the other ones?... but they are validated in schema

		for _, pattern := range highlights.Filetypes.Patterns {
			if _, err := regexp.Compile(pattern.Pattern); err != nil {
				problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid regexp: `%s`. See the valid syntax: https://golang.org/pkg/regexp/", pattern.Pattern)))
			}
		}

		return
	})

	go func() {
		conf.Watch(func() {
			config := conf.Get()
			if config == nil {
				return
			}

			if config.Highlights == nil {
				return
			}

			if defaultEngine, ok := engineNameToEngineType(config.Highlights.Engine.Default); ok {
				engineConfig.Default = defaultEngine
			}

			engineConfig.Overrides = map[string]EngineType{}
			for name, engine := range config.Highlights.Engine.Overrides {
				if overrideEngine, ok := engineNameToEngineType(engine); ok {
					engineConfig.Overrides[name] = overrideEngine
				}
			}

			highlightConfig.Extensions = config.Highlights.Filetypes.Extensions
			highlightConfig.Patterns = []ftPattern{}
			for _, pattern := range config.Highlights.Filetypes.Patterns {
				if re, err := regexp.Compile(pattern.Pattern); err == nil {
					highlightConfig.Patterns = append(highlightConfig.Patterns, ftPattern{pattern: re, filetype: pattern.Filetype})
				}
			}
		})
	}()
}

func engineNameToEngineType(engineName string) (engine EngineType, ok bool) {
	switch engineName {
	case "tree-sitter":
		return EngineTreeSitter, true
	case "syntect":
		return EngineSyntect, true
	default:
		return EngineInvalid, false
	}
}

// Matches against config, otherwise uses enry to get default
func matchConfig(path string) (string, bool) {
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if ft, ok := highlightConfig.Extensions[extension]; ok {
		return ft, true
	}

	for _, pattern := range highlightConfig.Patterns {
		if pattern.pattern != nil && pattern.pattern.MatchString(path) {
			return pattern.filetype, true
		}
	}

	return "", false
}

func getFiletype(path string, contents string) (string, bool) {
	ft, found := matchConfig(path)
	if found {
		return ft, true
	}

	return enry.GetLanguage(path, []byte(contents)), false
}

// TODO: Expose as an endpoint so you can type in a path and get the result in the front end?
func DetectSyntaxHighlightingFiletype(path string, contents string) SyntaxEngineQuery {
	ft, override := getFiletype(path, contents)

	engine := engineConfig.Default
	if overrideEngine, ok := engineConfig.Overrides[ft]; ok {
		engine = overrideEngine
	}

	return SyntaxEngineQuery{
		Filetype:         ft,
		FiletypeOverride: override,
		Engine:           engine,
	}
}
