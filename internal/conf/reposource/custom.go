package reposource

import (
	"fmt"
	"strings"

	"github.com/grafana/regexp"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func init() {
	conf.ContributeValidator(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		for _, c := range c.SiteConfig().GitCloneURLToRepositoryName {
			if _, err := regexp.Compile(c.From); err != nil {
				problems = append(problems, conf.NewSiteProblem(fmt.Sprintf("Not a valid regexp: %s. See the valid syntax: https://golang.org/pkg/regexp/", c.From)))
			}
		}
		return
	})
}

type cloneURLResolver struct {
	from *regexp.Regexp
	to   string
}

// cloneURLResolvers is the list of clone-URL-to-repo-URI mappings, derived
// from the site config
var cloneURLResolvers = conf.Cached[[]*cloneURLResolver](func() []*cloneURLResolver {
	cloneURLConfig := conf.Get().GitCloneURLToRepositoryName
	var resolvers []*cloneURLResolver
	for _, c := range cloneURLConfig {
		from, err := regexp.Compile(c.From)
		if err != nil {
			// Skip if there's an error. A user-visible validation error will appear due to the ContributeValidator call above.
			log15.Error("Site config: unable to compile Git clone URL mapping regexp", "regexp", c.From)
			continue
		}
		resolvers = append(resolvers, &cloneURLResolver{
			from: from,
			to:   c.To,
		})
	}
	return resolvers
})

// CustomCloneURLToRepoName maps from clone URL to repo name using custom mappings specified by the
// user in site config. An empty string return value indicates no match.
func CustomCloneURLToRepoName(cloneURL string) (repoName api.RepoName) {
	for _, r := range cloneURLResolvers() {
		if name := mapString(r.from, cloneURL, r.to); name != "" {
			return api.RepoName(name)
		}
	}
	return ""
}

func mapString(r *regexp.Regexp, in string, outTmpl string) string {
	namedMatches := make(map[string]string)
	matches := r.FindStringSubmatch(in)
	if matches == nil {
		return ""
	}
	for i, name := range r.SubexpNames() {
		if i == 0 {
			continue
		}
		namedMatches[name] = matches[i]
	}

	replacePairs := make([]string, 0, len(namedMatches)*2)
	for k, v := range namedMatches {
		replacePairs = append(replacePairs, fmt.Sprintf("{%s}", k), v)
	}
	return strings.NewReplacer(replacePairs...).Replace(outTmpl)
}
