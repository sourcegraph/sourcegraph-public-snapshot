package validation

import (
	"fmt"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/highlight"
)

func init() {
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		highlights := c.SiteConfig().SyntaxHighlighting
		if highlights == nil {
			return
		}

		if _, ok := highlight.EngineNameToEngineType(highlights.Engine.Default); !ok {
			problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid highlights.Engine.Default: `%s`.", highlights.Engine.Default)))
		}

		for _, engine := range highlights.Engine.Overrides {
			if _, ok := highlight.EngineNameToEngineType(engine); !ok {
				problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid highlights.Engine.Default: `%s`.", engine)))
			}
		}

		for _, pattern := range highlights.Languages.Patterns {
			if _, err := regexp.Compile(pattern.Pattern); err != nil {
				problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid regexp: `%s`. See the valid syntax: https://golang.org/pkg/regexp/", pattern.Pattern)))
			}
		}

		return
	})
}
