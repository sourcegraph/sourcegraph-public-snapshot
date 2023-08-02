package confvalidators

import (
	"fmt"
	"regexp"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func validateGitCloneURLMappings(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
	for _, c := range c.SiteConfig().GitCloneURLToRepositoryName {
		if _, err := regexp.Compile(c.From); err != nil {
			problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid regexp: %s. See the valid syntax: https://golang.org/pkg/regexp/", c.From)))
		}
	}
	return
}
