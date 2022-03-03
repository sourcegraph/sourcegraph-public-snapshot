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

type ftPattern struct {
	pattern  *regexp.Regexp
	filetype string
}

// TODO: Decide on capitalization, cause it's a nightmare otherwise.
// TODO: Validate that those are available filetypes, otherwise it's kind of
//       also pointless (for example, how to tell them that C# is not OK, but c_sharp is?)
type ftConfig struct {
	// Order does not matter. Evaluated before Patterns
	Extensions map[string]string

	// Order matters for this. First matching pattern matches.
	// Matches against the entire string.
	Patterns []ftPattern

	// TODO:
	// Shebang
}

var highlightConfig = ftConfig{}

func init() {
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		highlights := c.SiteConfig().Highlights
		if highlights == nil {
			return
		}

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

func getFiletype(path string, contents string) string {
	ft, found := matchConfig(path)
	if found {
		return ft
	}

	return enry.GetLanguage(query.path, []byte(query.contents))
}

// TODO: Expose as an endpoint so you can type in a path and get the result in the front end?
func DetectSyntaxHighlightingFiletype(query ftQuery) string {
	return normalizeFilepath(getFiletype(query))
}
