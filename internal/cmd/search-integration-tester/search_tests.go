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
		Query: `repo:auth0-go-jwt-middleware$`,
	},
	{
		Name:  `Global search, repo search by name, case yes, nonzero result`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ String case:yes count:1 stable:yes`,
	},
	{
		Name:  `True is an alias for yes when fork is set`,
		Query: `fork:true repo:github\.com/rvantonderp/(beego-mux|sgtest-mux)`,
	},
	// Text search, focused to repo.
	{
		Name:  `Repo search, non-master branch, nonzero result`,
		Query: `repo:^github.com/facebook/react$@0.3-stable var ExecutionEnvironment = require('ExecutionEnvironment'); patterntype:literal count:1 stable:yes`,
	},
	{
		Name:  `Repo search, indexed multiline search, nonzero result`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ \nimport index:only count:1 stable:yes`,
	},
	{
		Name:  `Repo search, unindexed multiline search, nonzero result`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ \nimport index:no count:1 stable:yes`,
	},
	{
		Name:  `Repo search, zero results`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ doesnot734734743734743exist`,
	},
	// Commit search.
	{
		Name:  `Commit search, nonzero result`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ type:commit count:1`,
	},
	// Diff search.
	{
		Name:  `Diff search, nonzero result`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ type:diff os.Args count:1`,
	},
	// Timeout search options.
	{
		Name:  `Search timeout option, alert raised`,
		Query: `router index:no timeout:1ns`,
	},
	// Global simple text search.
	{
		Name:  `Global search, zero results`,
		Query: `asdfalksd+jflaksjdfklas patterntype:literal -repo:sourcegraph`,
	},
	{
		Name:  `Global search, double-quoted pattern, nonzero result`,
		Query: `"error type:\n" patterntype:regexp count:1 stable:yes`,
	},
	{
		Name:  `Global search, exclude repo, nonzero result`,
		Query: `"error type:\n" -repo:DirectXMan12 patterntype:regexp count:1 stable:yes`,
	},
	// Repohascommitafter.
	{
		Name:  `Global search, repohascommitafter, nonzero result`,
		Query: `repohascommitafter:"5 months ago" test patterntype:literal count:1 stable:yes`,
	},
	// Global regex text search.
	{
		Name:  `Global search, regex, unindexed, nonzero result`,
		Query: `^func.*$ patterntype:regexp index:only count:1 stable:yes`,
	},
	{
		Name:  `Global search, fork only, nonzero result`,
		Query: `fork:only FORK_SENTINEL`,
	},
	{
		Name:  `Global search, filter by language`,
		Query: `\bfunc\b lang:go count:1 stable:yes`,
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
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ make(:[1]) index:only patterntype:structural count:3`,
	},
	{
		Name:  `Structural search quotes are interpreted literally`,
		Query: `repo:^github\.com/rvantonderp/auth0-go-jwt-middleware$ file:^README\.md "This :[_] authenticated :[_]" patterntype:structural`,
	},
	{
		Name:  `Alert to activate structural search mode`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ patterntype:literal i can't :[believe] it's not butter`,
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
		Query: `type:repo sgtest-mux`,
	},
	{
		Name:  `Global search, fork included, nonzero result`,
		Query: `type:repo sgtest-mux fork:yes`,
	},
	{
		Name:  `Global search, fork included if exact without option, nonzero result`,
		Query: `repo:^github\.com/rvantonderp/sgtest-mux$`,
	},
	{
		Name:  `Global search, exclude counts for fork and archive`,
		Query: `repo:mux|archive|caddy`,
	},
	{
		Name:  `Repo visibility`,
		Query: `repo:github.com/rvantonderp/adjust-go-wrk visibility:public`,
	},
	// And/Or queries.
	{
		Name:  `And operator, basic`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ func and main count:1 stable:yes type:file`,
	},
	{
		Name:  `Or operator, single and double quoted`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ "readConfig()" or 'buildHeaders' stable:yes type:file count:1 patterntype:regexp`,
	},
	{
		Name:  `Literals, grouped parens with parens-as-patterns heuristic`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ (() or ()) stable:yes type:file count:1 patterntype:regexp`,
	},
	{
		Name:  `Literals, no grouped parens`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ () or () stable:yes type:file count:1 patterntype:regexp`,
	},
	{
		Name:  `Literals, escaped parens`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ \(\) or \(\) stable:yes type:file count:1 patterntype:regexp`,
	},
	{
		Name:  `Literals, escaped and unescaped parens, no group`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ () or \(\) stable:yes type:file count:1 patterntype:regexp`,
	},
	{
		Name:  `Literals, escaped and unescaped parens, grouped`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ (() or \(\)) stable:yes type:file count:1 patterntype:regexp`,
	},
	{
		Name:  `Literals, double paren`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ ()() or ()()`,
	},
	{
		Name:  `Literals, double paren, dangling paren right side`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ ()() or main()(`,
	},
	{
		Name:  `Literals, double paren, dangling paren left side`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ ()( or ()()`,
	},
	{
		Name:  `Mixed regexp and literal`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ func(.*) or does_not_exist_3744 count:1 stable:yes type:file`,
	},
	{
		Name:  `Mixed regexp and literal heuristic`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ func( or func(.*) count:1 stable:yes type:file`,
	},
	{
		Name:  `Mixed regexp and quoted literal`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ "*" and cert.*Load count:1 stable:yes type:file`,
	},
	{
		Name:  `Escape sequences`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ \' and \" and \\ and /`,
	},
	{
		Name:  `Escaped whitespace sequences with 'and'`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ \ and /`,
	},
	{
		Name:  `Concat converted to spaces for literal search`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ file:^client\.go caCertPool := or x509 Pool patterntype:literal`,
	},
	{
		Name:  `Literal parentheses match pattern`,
		Query: `repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$@4b5788e file:^cmd/adapter/adapter\.go main() and InitLogs() patterntype:literal`,
	},
	{
		Name:  `Dangling right parens, heuristic for literal search`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ split) and main patterntype:literal`,
	},
	{
		Name:  `Dangling right parens, heuristic for literal search, double parens`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ respObj.Size and data)) patterntype:literal`,
	},
	{
		Name:  `Dangling right parens, heuristic for literal search, simple group before right paren`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ respObj.Size and (data)) patterntype:literal`,
	},
	{
		Name:  `Dangling right parens, heuristic for literal search, cannot succeed, too confusing`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ (respObj.Size and (data))) patterntype:literal`,
	},
	{
		Name:  `Confusing grouping raises alert`,
		Query: `repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$@4b5788e file:^README\.md (bar and (foo or x\) ()) patterntype:literal`,
	},
	{
		Name:  `Successful grouping removes alert`,
		Query: `repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$@4b5788e file:^README\.md (bar and (foo or (x\) ())) patterntype:literal`,
	},
	{
		Name:  `No dangling right paren with complex group for literal search`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ (respObj.Size and (data)) patterntype:literal`,
	},
	{
		Name:  `Concat converted to .* for regexp search`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ file:^client\.go ca Pool or x509 Pool patterntype:regexp stable:yes type:file`,
	},
	{
		Name:  `Structural search uses literal search parser`,
		Query: `repo:^github\.com/rvantonderp/adjust-go-wrk$ file:^client\.go :[[v]] := x509 and AppendCertsFromPEM(:[_]) patterntype:structural`,
	},
	{
		Name:  `Union file matches per file and accurate counts`,
		Query: `repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$@4b5788e file:^cmd/adapter/adapter\.go func or main`,
	},
	{
		Name:  `Intersect file matches per file and accurate counts`,
		Query: `repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$@4b5788e file:^cmd/adapter/adapter\.go func and main`,
	},
	{
		Name:  `Simple combined union and intersect file matches per file and accurate counts`,
		Query: `repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$@4b5788e file:^cmd/adapter/adapter\.go ((func main and package main) or return prom.NewClient)`,
	},
	{
		Name:  `Complex union of intersect file matches per file and accurate counts`,
		Query: `repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$@4b5788e file:^cmd/adapter/adapter\.go ((main and NamersFromConfig) or (genericPromClient and stopCh <-))`,
	},
	{
		Name:  `Complex intersect of union file matches per file and accurate counts`,
		Query: `repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$@4b5788e file:^cmd/adapter/adapter\.go ((func main or package main) and (baseURL or mprom))`,
	},
	{
		Name:  `Intersect file matches per file against an empty result set`,
		Query: `repo:^github\.com/rvantonderp/DirectXMan12-k8s-prometheus-adapter$@4b5788e file:^cmd/adapter/adapter\.go func and doesnotexist838338`,
	},
	{
		Name:  `Dedupe union operation`,
		Query: `file:cors_filter.go|ginkgo_dsl.go|funcs.go repo:rvantonderp/DirectXMan12-k8s-prometheus-adapter  :[[i]], :[[x]] := range :[src.] { :[[dst]][:[i]] = :[[x]] } or if strings.ToLower(:[s1]) == strings.ToLower(:[s2]) patterntype:structural`,
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
