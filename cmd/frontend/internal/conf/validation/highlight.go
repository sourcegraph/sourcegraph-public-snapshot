package validation

import (
	"fmt"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/highlight"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func init() {
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		highlightingConfig := c.SiteConfig().SyntaxHighlighting
		if highlightingConfig == nil {
			return
		}

		if highlightingConfig.Engine != nil {
			if _, ok := highlight.EngineNameToEngineType(highlightingConfig.Engine.Default); !ok {
				problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid highlights.Engine.Default: `%s`.", highlightingConfig.Engine.Default)))
			}
			for _, engine := range highlightingConfig.Engine.Overrides {
				if _, ok := highlight.EngineNameToEngineType(engine); !ok {
					problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid highlights.Engine.Override: `%s`.", engine)))
				}
			}
		}

		if highlightingConfig.Languages != nil {
			for _, pattern := range highlightingConfig.Languages.Patterns {
				if _, err := regexp.Compile(pattern.Pattern); err != nil {
					problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid regexp: `%s`. See the valid syntax: https://golang.org/pkg/regexp/", pattern.Pattern)))
				}
			}
		}

		return
	})
}
