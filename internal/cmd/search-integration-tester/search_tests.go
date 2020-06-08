package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

type test struct {
	Name  string
	Query string
}

const searchTestDataDir = "testdata/search"

var tests = []test{
	// Repo search (part 1).
	{
		Name:  `Global search, repo search by name, nonzero result`,
		Query: `repo:auth0/go-jwt-middleware$`,
	},
	{
		Name:  `Global search, repo search by name, case yes, nonzero result`,
		Query: `String repo:^github.com/adjust/go-wrk$ case:yes count:1 stable:yes`,
	},
	// Text search, focused to repo.
	{
		Name:  `Global search, non-master branch, large repo, nonzero result`,
		Query: `repo:^github.com/facebook/react$@0.3-stable var ExecutionEnvironment = require('ExecutionEnvironment'); patterntype:literal count:1 stable:yes`,
	},
	{
		Name:  `Global search, indexed multiline search, nonzero result`,
		Query: `repo:^github\.com/facebook/react$ componentDidMount\(\) {\n\s*this patterntype:regexp count:1 stable:yes`,
	},
	{
		Name:  `Global search, indexed multiline search, zero results`,
		Query: `repo:^github\.com/facebook/react$ componentDidMount\(\) {\n\s*this\.props\.sourcegraph\(`,
	},
	{
		Name:  `Global search, unindexed multiline search, nonzero result`,
		Query: `repo:^github\.com/facebook/react$ componentDidMount\(\) {\n\s*this index:no count:1 stable:yes`,
	},
	{
		Name:  `Global search, unindexed multiline search, zero results`,
		Query: `repo:^github\.com/facebook/react$ componentDidMount\(\) {\n\s*this\.props\.sourcegraph\( index:no`,
	},
	// Commit search.
	{
		Name:  `Commit search, nonzero result`,
		Query: `repo:^github\.com/facebook/react$ type:commit hello world count:1`,
	},
	// Diff search.
	{
		Name:  `Diff search, nonzero result`,
		Query: `repo:^github\.com/sgtest/mux$ type:diff main count:1`,
	},
	// Timeout search options.
	{
		Name:  `Search timeout option, alert raised`,
		Query: `router index:no timeout:1ns`,
	},
	// Global simple text search.
	{
		Name:  `Global search, zero results`,
		Query: `asdfalksd+jflaksjdfklas patterntype:literal`,
	},
	{
		Name:  `Global search, double-quoted pattern, nonzero result`,
		Query: `"error type:\n" count:1 stable:yes`,
	},
	{
		Name:  `Global search, exclude repo, nonzero result`,
		Query: `"error type:\n" count:1 -repo:DirectXMan12 stable:yes`,
	},
	// Repohascommitafter.
	{
		Name:  `Global search, repohascommitafter, nonzero result`,
		Query: `repohascommitafter:"5 months ago" test patterntype:literal count:1 stable:yes`,
	},
	{
		Name:  `Global search, repohascommitafter, nonzero result`,
		Query: `repohascommitafter:"5 months ago" test patterntype:literal count:1 stable:yes`,
	},
	// Global regex text search.
	{
		Name:  `Global search, regex, unindexed, nonzero result`,
		Query: `^func.*$ index:only count:1 stable:yes`,
	},
	{
		Name:  `Global search, fork only, nonzero result`,
		Query: `fork:only FORK_SENTINEL`,
	},
	{
		Name:  `Global search, filter by language`,
		Query: `\bfunc\b lang:js count:1 stable:yes`,
	},
	{
		Name:  `Global search, filename, zero results`,
		Query: `file:asdfasdf.go`,
	},
	{
		Name:  `Global search, filename, nonzero result`,
		Query: `file:router.go count:1 stable:yes`,
	},
	// Symbol search.
	{
		Name:  `Global search, symbols, nonzero result`,
		Query: `type:symbol test count:1 stable:yes`,
	},
	// Structural search.
	{
		Name:  `Global search, structural, index only, nonzero result`,
		Query: `repo:^github\.com/facebook/react$ index:only patterntype:structural toHaveYielded(:[args]) count:5`,
	},
	// Repo search (part 2).
	{
		Name:  `Global search, archived excluded, zero results`,
		Query: `type:repo facebookarchive`,
	},
	{
		Name:  `Global search, archived included, nonzero result`,
		Query: `type:repo facebookarchive archived:yes`,
	},
	{
		Name:  `Global search, archived included if exact without option, nonzero result`,
		Query: `repo:^github\.com/facebookarchive/httpcontrol$`,
	},
	{
		Name:  `Global search, fork excluded, zero results`,
		Query: `type:repo sgtest/mux`,
	},
	{
		Name:  `Global search, fork included, nonzero result`,
		Query: `type:repo sgtest/mux fork:yes`,
	},
	{
		Name:  `Global search, fork included if exact without option, nonzero result`,
		Query: `repo:^github\.com/sgtest/mux$`,
	},
	{
		Name:  `Global search, exclude counts for fork and archive`,
		Query: `repo:mux|archive|caddy`,
	},
	// And/Or queries.
	{
		Name:  `And operator, basic`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ func and main type:file stable:yes count:1 lang:go file:^cmd/frontend/auth/user_test\.go$ patterntype:regexp`,
	},
	{
		Name:  `Or operator, single and double quoted`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ 'Search()' or "generate" type:file stable:yes lang:go file:^cmd/frontend/graphqlbackend/codeintel_usage_stats\.go$|^cmd/frontend/authz/perms.go$ patterntype:regexp`,
	},
	{
		Name:  `Literals, grouped parens with parens-as-patterns heuristic`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ (() or ()) stable:yes type:file count:1 file:^\.buildkite/hooks/pre-command$ patterntype:regexp`,
	},
	{
		Name:  `Literals, no grouped parens`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ () or () stable:yes type:file count:1 file:^\.buildkite/hooks/pre-command$ patterntype:regexp`,
	},
	{
		Name:  `Literals, escaped parens`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ \(\) or \(\) stable:yes type:file count:1 file:^\.buildkite/hooks/pre-command$ patterntype:regexp`,
	},
	{
		Name:  `Literals, escaped and unescaped parens`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ (() or \(\)) stable:yes type:file count:1 file:^\.buildkite/hooks/pre-command$`,
	},
	{
		Name:  `Literals, escaped and unescaped parens, no group`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ () or \(\)`,
	},
	{
		Name:  `Literals, double paren`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ ()() or ()()`,
	},
	{
		Name:  `Literals, double paren, dangling paren right side`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ ()() or main()(`,
	},
	{
		Name:  `Literals, double paren, dangling paren left side`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ ()( or ()()`,
	},
	{
		Name:  `Mixed regexp and literal`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ func(.*) or does_not_exist_3744`,
	},
	{
		Name:  `Mixed regexp and literal heuristic`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ func(.*) or func(`,
	},
	{
		Name:  `Escape sequences`,
		Query: `repo:^github\.com/sourcegraph/sourcegraph$ \' and \" and \\ and / file:^\.buildkite/hooks/pre-command$`,
	},
}

func sanitizeFilename(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ',', ' ', '.', '-':
			return '_'
		default:
			return r
		}
	}, s)
}

func runSearchTests() error {
	_, err := os.Stat(searchTestDataDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(searchTestDataDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	if err != nil {
		fmt.Println(err)
	}
	for _, test := range tests {
		got, err := search(test.Query)
		if err != nil {
			return err
		}
		filename := strings.ToLower(sanitizeFilename(test.Name))
		goldenPath := path.Join(searchTestDataDir, fmt.Sprintf("%s.golden", filename))
		if err := assertGolden(goldenPath, got); err != nil {
			if update || updateAll {
				err := assertUpdate(goldenPath, got)
				if err != nil {
					return err
				}
				fmt.Printf("Updated Test %s\n", test.Name)
				if update {
					return nil
				}
				continue
			}
			return fmt.Errorf("TEST FAILURE: %s\nQuery: %s\n%s", test.Name, test.Query, err)
		}
	}
	return nil
}
