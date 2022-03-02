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

type ftQuery struct {
	path     string
	contents string
}

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
}

// Matches against config, otherwise uses enry to get default
func matchConfig(config ftConfig, query ftQuery) (string, bool) {
	extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(query.path), "."))
	if ft, ok := config.Extensions[extension]; ok {
		return ft, true
	}

	for _, pattern := range config.Patterns {
		if pattern.pattern != nil && pattern.pattern.MatchString(query.path) {
			return pattern.filetype, true
		}
	}

	return "", false
}

func getFiletype(config ftConfig, query ftQuery) string {
	ft, found := matchConfig(config, query)
	if found {
		return ft
	}

	return enry.GetLanguage(query.path, []byte(query.contents))
}

// TODO: Expose as an endpoint so you can type in a path and get the result in the front end?
func DetectSyntaxHighlightingFiletype(config ftConfig, query ftQuery) string {
	return normalizeFilepath(getFiletype(config, query))
}
