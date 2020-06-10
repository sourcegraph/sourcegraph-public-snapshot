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
		Name:  `Literals, escaped`,
		Query: `repo:^github\.com/facebook/react$ (\(\) or \(\)) stable:yes type:file count:1 patterntype:regexp`,
	},
	{
		Name:  `Literals, no parens escaped`,
		Query: `repo:^github\.com/facebook/react$ \(\) or \(\) stable:yes type:file count:1 patterntype:regexp`,
	},
	{
		Name:  `And operator, basic`,
		Query: `repo:^github\.com/facebook/react$ func and main stable:yes type:file count:1 file:^dangerfile.js$`,
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
		if err := assertGolden(test.Name, goldenPath, got, update); err != nil {
			return fmt.Errorf("TEST FAILURE: %s\n%s", test.Name, err)
		}
	}
	return nil
}
