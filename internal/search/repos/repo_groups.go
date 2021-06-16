package repos

import (
	"context"
	"fmt"
	"regexp"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A repogroup value is either a exact repo path RepoPath, or a regular
// expression pattern RepoRegexpPattern.
type RepoGroupValue interface {
	value()
	String() string
}

type RepoPath string
type RepoRegexpPattern string

func (RepoPath) value() {}
func (r RepoPath) String() string {
	return string(r)
}

func (RepoRegexpPattern) value() {}
func (r RepoRegexpPattern) String() string {
	return string(r)
}

var MockResolveRepoGroups func() (map[string][]RepoGroupValue, error)

// RepoGroupsToIncludePatterns converts the repo groups from settings to an include pattern
func RepoGroupsToIncludePatterns(groupNames []string, groups map[string][]RepoGroupValue) (string, int) {
	patterns := repoGroupValuesToRegexp(groupNames, groups)
	return UnionRegExps(patterns), len(patterns)
}

// ResolveRepoGroups retrieves the repository group from settings and checks the database for any
// user configured repogroups.
func ResolveRepoGroups(ctx context.Context, settings *schema.Settings) (groups map[string][]RepoGroupValue, err error) {
	if MockResolveRepoGroups != nil {
		return MockResolveRepoGroups()
	}

	groups = ResolveRepoGroupsFromSettings(settings)

	if mode, err := database.GlobalUsers.CurrentUserAllowedExternalServices(ctx); err != nil {
		return groups, err
	} else if mode == conf.ExternalServiceModeDisabled {
		return groups, nil
	}

	a := actor.FromContext(ctx)
	repos, err := database.GlobalRepos.ListRepoNames(ctx, database.ReposListOptions{UserID: a.UID})
	if err != nil {
		log15.Warn("getting user added repos", "err", err)
		return groups, nil
	}

	if len(repos) == 0 {
		return groups, nil
	}

	values := make([]RepoGroupValue, 0, len(repos))
	for _, repo := range repos {
		values = append(values, RepoPath(repo.Name))
	}
	groups["my"] = values

	return groups, nil
}

// ResolveRepoGroupsFromSettings retrieves the repository group configuration from settings. Will not check
// the database for any user configured repogroups.
func ResolveRepoGroupsFromSettings(settings *schema.Settings) map[string][]RepoGroupValue {
	groups := map[string][]RepoGroupValue{}

	for name, values := range settings.SearchRepositoryGroups {
		repos := make([]RepoGroupValue, 0, len(values))

		for _, value := range values {
			switch path := value.(type) {
			case string:
				repos = append(repos, RepoPath(path))
			case map[string]interface{}:
				if stringRegex, ok := path["regex"].(string); ok {
					repos = append(repos, RepoRegexpPattern(stringRegex))
				} else {
					log15.Warn("ignoring repo group value because regex not specified", "regex-string", path["regex"])
				}
			default:
				log15.Warn("ignoring repo group value of unrecognized type", "value", value, "type", fmt.Sprintf("%T", value))
			}
		}
		groups[name] = repos
	}

	return groups
}

// repoGroupValuesToRegexp does a lookup of all repo groups by name and converts
// their values to a list of regular expressions to search.
func repoGroupValuesToRegexp(groupNames []string, groups map[string][]RepoGroupValue) []string {
	var patterns []string
	for _, groupName := range groupNames {
		for _, value := range groups[groupName] {
			switch v := value.(type) {
			case RepoPath:
				patterns = append(patterns, "^"+regexp.QuoteMeta(v.String())+"$")
			case RepoRegexpPattern:
				patterns = append(patterns, v.String())
			default:
				panic("unreachable")
			}
		}
	}
	return patterns
}
