package jobutil

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp/syntax"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"
	"github.com/hexops/autogold"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/printer"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
	zoektquery "github.com/sourcegraph/zoekt/query"
)

func TestNewPlanJob(t *testing.T) {
	cases := []struct {
		query      string
		protocol   search.Protocol
		searchType query.SearchType
		want       autogold.Value
	}{{
		query:      `foo context:@userA`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("user search context", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (SEQUENTIAL
          (ensureUnique . false)
          (REPOPAGER
            (repoOpts.searchContextSpec . @userA)
            (PARTIALREPOS
              (ZOEKTREPOSUBSETTEXTSEARCH
                (query . substr:"foo")
                (type . text))))
          (REPOPAGER
            (repoOpts.searchContextSpec . @userA)
            (PARTIALREPOS
              (SEARCHERTEXTSEARCH
                (indexed . false)))))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.searchContextSpec . @userA))
        (PARALLEL
          NoopJob
          (REPOSEARCH
            (repoOpts.repoFilters.0 . foo)(repoOpts.searchContextSpec . @userA)))))))`),
	}, {
		query:      `foo context:global`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("global search explicit context", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . substr:"foo")
          (type . text)
          (repoOpts.searchContextSpec . global))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.searchContextSpec . global))
        (REPOSEARCH
          (repoOpts.repoFilters.0 . foo)(repoOpts.searchContextSpec . global))))))`),
	}, {
		query:      `foo`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("global search implicit context", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . substr:"foo")
          (type . text)
          )
        (REPOSCOMPUTEEXCLUDED
          )
        (REPOSEARCH
          (repoOpts.repoFilters.0 . foo))))))`),
	}, {
		query:      `foo repo:sourcegraph/sourcegraph`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("nonglobal repo", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (SEQUENTIAL
          (ensureUnique . false)
          (REPOPAGER
            (repoOpts.repoFilters.0 . sourcegraph/sourcegraph)
            (PARTIALREPOS
              (ZOEKTREPOSUBSETTEXTSEARCH
                (query . substr:"foo")
                (type . text))))
          (REPOPAGER
            (repoOpts.repoFilters.0 . sourcegraph/sourcegraph)
            (PARTIALREPOS
              (SEARCHERTEXTSEARCH
                (indexed . false)))))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.repoFilters.0 . sourcegraph/sourcegraph))
        (PARALLEL
          NoopJob
          (REPOSEARCH
            (repoOpts.repoFilters.0 . sourcegraph/sourcegraph)(repoOpts.repoFilters.1 . foo)))))))`),
	}, {
		query:      `ok ok`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("supported repo job", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . regex:"ok(?-s:.)*?ok")
          (type . text)
          )
        (REPOSCOMPUTEEXCLUDED
          )
        (REPOSEARCH
          (repoOpts.repoFilters.0 . (?:ok).*?(?:ok)))))))`),
	}, {
		query:      `ok @thing`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Want("supported repo job literal", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . literal)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . substr:"ok @thing")
          (type . text)
          )
        (REPOSCOMPUTEEXCLUDED
          )
        (REPOSEARCH
          (repoOpts.repoFilters.0 . ok ))))))`),
	}, {
		query:      `@nope`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("unsupported repo job literal", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . substr:"@nope")
          (type . text)
          )
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `foo @bar`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("unsupported repo job regexp", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . regex:"foo(?-s:.)*?@bar")
          (type . text)
          )
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:symbol test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("symbol", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALSYMBOLSEARCH
          (query . sym:substr:"test")
          (type . symbol)
          )
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:commit test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("commit", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts.onlyCloned . true)
          (diff . false)
          (limit . 500))
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:diff test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("diff", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (DIFFSEARCH
          (query . *protocol.DiffMatches(test))
          (repoOpts.onlyCloned . true)
          (diff . true)
          (limit . 500))
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:file type:commit test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("streaming file or commit", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . content_substr:"test")
          (type . text)
          )
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts.onlyCloned . true)
          (diff . false)
          (limit . 500))
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:file type:path type:repo type:commit type:symbol repo:test test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("streaming many types", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (SEQUENTIAL
          (ensureUnique . false)
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (PARTIALREPOS
              (ZOEKTREPOSUBSETTEXTSEARCH
                (query . substr:"test")
                (type . text))))
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (PARTIALREPOS
              (SEARCHERTEXTSEARCH
                (indexed . false)))))
        (REPOPAGER
          (repoOpts.repoFilters.0 . test)
          (PARTIALREPOS
            (ZOEKTSYMBOLSEARCH
              (query . sym:substr:"test"))))
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts.repoFilters.0 . test)(repoOpts.onlyCloned . true)
          (diff . false)
          (limit . 500))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.repoFilters.0 . test))
        (PARALLEL
          NoopJob
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (PARTIALREPOS
              (SEARCHERSYMBOLSEARCH
                (patternInfo.pattern . test)(patternInfo.isRegexp . true)(patternInfo.fileMatchLimit . 500)(patternInfo.patternMatchesPath . true)
                (numRepos . 0)
                (limit . 500))))
          (REPOSEARCH
            (repoOpts.repoFilters.0 . test)(repoOpts.repoFilters.1 . test)))))))`),
	}, {
		query:      `type:file type:commit test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("batched file or commit", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALTEXTSEARCH
          (query . content_substr:"test")
          (type . text)
          )
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts.onlyCloned . true)
          (diff . false)
          (limit . 500))
        (REPOSCOMPUTEEXCLUDED
          )
        NoopJob))))`),
	}, {
		query:      `type:file type:path type:repo type:commit type:symbol repo:test test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("batched many types", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (SEQUENTIAL
          (ensureUnique . false)
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (PARTIALREPOS
              (ZOEKTREPOSUBSETTEXTSEARCH
                (query . substr:"test")
                (type . text))))
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (PARTIALREPOS
              (SEARCHERTEXTSEARCH
                (indexed . false)))))
        (REPOPAGER
          (repoOpts.repoFilters.0 . test)
          (PARTIALREPOS
            (ZOEKTSYMBOLSEARCH
              (query . sym:substr:"test"))))
        (COMMITSEARCH
          (query . *protocol.MessageMatches(test))
          (repoOpts.repoFilters.0 . test)(repoOpts.onlyCloned . true)
          (diff . false)
          (limit . 500))
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.repoFilters.0 . test))
        (PARALLEL
          NoopJob
          (REPOPAGER
            (repoOpts.repoFilters.0 . test)
            (PARTIALREPOS
              (SEARCHERSYMBOLSEARCH
                (patternInfo.pattern . test)(patternInfo.isRegexp . true)(patternInfo.fileMatchLimit . 500)(patternInfo.patternMatchesPath . true)
                (numRepos . 0)
                (limit . 500))))
          (REPOSEARCH
            (repoOpts.repoFilters.0 . test)(repoOpts.repoFilters.1 . test)))))))`),
	}, {
		query:      `(type:commit or type:diff) (a or b)`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		// TODO this output doesn't look right. There shouldn't be any zoekt or repo jobs
		want: autogold.Want("complex commit diff", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (OR
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (COMMITSEARCH
            (query . (*protocol.MessageMatches((?:a)|(?:b))))
            (repoOpts.onlyCloned . true)
            (diff . false)
            (limit . 500))
          (REPOSCOMPUTEEXCLUDED
            )
          (OR
            NoopJob
            NoopJob))))
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (DIFFSEARCH
            (query . (*protocol.DiffMatches((?:a)|(?:b))))
            (repoOpts.onlyCloned . true)
            (diff . true)
            (limit . 500))
          (REPOSCOMPUTEEXCLUDED
            )
          (OR
            NoopJob
            NoopJob))))))`),
	}, {
		query:      `(type:repo a) or (type:file b)`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("disjunct types", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (OR
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (REPOSCOMPUTEEXCLUDED
            )
          (REPOSEARCH
            (repoOpts.repoFilters.0 . a)))))
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (ZOEKTGLOBALTEXTSEARCH
            (query . content_substr:"b")
            (type . text)
            )
          (REPOSCOMPUTEEXCLUDED
            )
          NoopJob)))))`),
	}, {
		query:      `type:symbol a or b`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Want("symbol with or", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (ZOEKTGLOBALSYMBOLSEARCH
          (query . (or sym:substr:"a" sym:substr:"b"))
          (type . symbol)
          )
        (REPOSCOMPUTEEXCLUDED
          )
        (OR
          NoopJob
          NoopJob)))))`),
	},
		{
			query:      `repo:contains.path(a) repo:contains.content(b)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Want("repo contains path and repo contains content", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.hasFileContent[0].path . a)(repoOpts.hasFileContent[1].content . b))
        (REPOSEARCH
          (repoOpts.hasFileContent[0].path . a)(repoOpts.hasFileContent[1].content . b))))))`),
		}, {
			query:      `repo:contains.file(path:a content:b)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Want("repo contains path and content", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.hasFileContent[0].path . a)(repoOpts.hasFileContent[0].content . b))
        (REPOSEARCH
          (repoOpts.hasFileContent[0].path . a)(repoOpts.hasFileContent[0].content . b))))))`),
		}, {
			query:      `repo:has(key:value)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Want("repo has kvp", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.hasKVPs[0].key . key)(repoOpts.hasKVPs[0].value . value))
        (REPOSEARCH
          (repoOpts.hasKVPs[0].key . key)(repoOpts.hasKVPs[0].value . value))))))`),
		}, {
			query:      `repo:has.tag(tag)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Want("repo has tag", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . regex)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (REPOSCOMPUTEEXCLUDED
          (repoOpts.hasKVPs[0].key . tag))
        (REPOSEARCH
          (repoOpts.hasKVPs[0].key . tag))))))`),
		}, {
			query:      `(...)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeStructural,
			want: autogold.Want("stream structural search", `
(ALERT
  (query . )
  (originalQuery . )
  (patternType . structural)
  (TIMEOUT
    (timeout . 20s)
    (LIMIT
      (limit . 500)
      (PARALLEL
        (REPOSCOMPUTEEXCLUDED
          )
        (STRUCTURALSEARCH
          (patternInfo.pattern . (:[_]))(patternInfo.isStructural . true)(patternInfo.fileMatchLimit . 500)
          )))))`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			plan, err := query.Pipeline(query.Init(tc.query, tc.searchType))
			require.NoError(t, err)

			inputs := &search.Inputs{
				UserSettings:        &schema.Settings{},
				PatternType:         tc.searchType,
				Protocol:            tc.protocol,
				Features:            &search.Features{},
				OnSourcegraphDotCom: true,
			}

			j, err := NewPlanJob(inputs, plan)
			require.NoError(t, err)

			tc.want.Equal(t, "\n"+printer.SexpPretty(j))
		})
	}
}

func TestToEvaluateJob(t *testing.T) {
	test := func(input string, protocol search.Protocol) string {
		q, _ := query.ParseLiteral(input)
		inputs := &search.Inputs{
			UserSettings:        &schema.Settings{},
			PatternType:         query.SearchTypeLiteral,
			Protocol:            protocol,
			OnSourcegraphDotCom: true,
		}

		b, _ := query.ToBasicQuery(q)
		j, _ := toFlatJobs(inputs, b)
		return "\n" + printer.SexpPretty(j) + "\n"
	}

	autogold.Want("root limit for streaming search", `
(REPOSEARCH
  (repoOpts.repoFilters.0 . foo))
`).Equal(t, test("foo", search.Streaming))

	autogold.Want("root limit for batch search", `
(REPOSEARCH
  (repoOpts.repoFilters.0 . foo))
`).Equal(t, test("foo", search.Batch))
}

func TestToTextPatternInfo(t *testing.T) {
	cases := []struct {
		input  string
		output autogold.Value
	}{{
		input:  `type:repo archived`,
		output: autogold.Want("01", `{"Pattern":"archived","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `type:repo archived archived:yes`,
		output: autogold.Want("02", `{"Pattern":"archived","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `type:repo sgtest/mux`,
		output: autogold.Want("04", `{"Pattern":"sgtest/mux","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `type:repo sgtest/mux fork:yes`,
		output: autogold.Want("05", `{"Pattern":"sgtest/mux","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `"func main() {\n" patterntype:regexp type:file`,
		output: autogold.Want("10", `{"Pattern":"func main\\(\\) \\{\n","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `"func main() {\n" -repo:go-diff patterntype:regexp type:file`,
		output: autogold.Want("11", `{"Pattern":"func main\\(\\) \\{\n","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ String case:yes type:file`,
		output: autogold.Want("12", `{"Pattern":"String","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":true,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":true,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$@v1 void sendPartialResult(Object requestId, JsonPatch jsonPatch); patterntype:literal type:file`,
		output: autogold.Want("13", `{"Pattern":"void sendPartialResult\\(Object requestId, JsonPatch jsonPatch\\);","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$@v1 void sendPartialResult(Object requestId, JsonPatch jsonPatch); patterntype:literal count:1 type:file`,
		output: autogold.Want("14", `{"Pattern":"void sendPartialResult\\(Object requestId, JsonPatch jsonPatch\\);","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":1,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$ \nimport index:only patterntype:regexp type:file`,
		output: autogold.Want("15", `{"Pattern":"\\nimport","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"only","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$ \nimport index:no patterntype:regexp type:file`,
		output: autogold.Want("16", `{"Pattern":"\\nimport","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"no","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$ doesnot734734743734743exist`,
		output: autogold.Want("17", `{"Pattern":"doesnot734734743734743exist","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ type:commit test`,
		output: autogold.Want("21", `{"Pattern":"test","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ type:diff main`,
		output: autogold.Want("22", `{"Pattern":"main","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ repohascommitafter:"2019-01-01" test patterntype:literal`,
		output: autogold.Want("23", `{"Pattern":"test","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `^func.*$ patterntype:regexp index:only type:file`,
		output: autogold.Want("24", `{"Pattern":"^func.*$","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"only","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `fork:only patterntype:regexp FORK_SENTINEL`,
		output: autogold.Want("25", `{"Pattern":"FORK_SENTINEL","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `\bfunc\b lang:go type:file patterntype:regexp`,
		output: autogold.Want("26", `{"Pattern":"\\bfunc\\b","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["\\.go$"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":["go"]}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ make(:[1]) index:only patterntype:structural count:3`,
		output: autogold.Want("29", `{"Pattern":"make(:[1])","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":3,"Index":"only","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ make(:[1]) lang:go rule:'where "backcompat" == "backcompat"' patterntype:structural`,
		output: autogold.Want("30", `{"Pattern":"make(:[1])","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"where \"backcompat\" == \"backcompat\"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["\\.go$"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":["go"]}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$@adde71 make(:[1]) index:no patterntype:structural count:3`,
		output: autogold.Want("31", `{"Pattern":"make(:[1])","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":3,"Index":"no","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ file:^README\.md "basic :[_] access :[_]" patterntype:structural`,
		output: autogold.Want("32", `{"Pattern":"\"basic :[_] access :[_]\"","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["^README\\.md"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `no results for { ... } raises alert repo:^github\.com/sgtest/go-diff$`,
		output: autogold.Want("34", `{"Pattern":"no results for \\{ \\.\\.\\. \\} raises alert","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ patternType:regexp \ and /`,
		output: autogold.Want("49", `{"Pattern":"(?:\\ and).*?(?:/)","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ (not .svg) patterntype:literal`,
		output: autogold.Want("52", `{"Pattern":"\\.svg","IsNegated":true,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ (Fetches OR file:language-server.ts)`,
		output: autogold.Want("72", `{"Pattern":"Fetches","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ ((file:^renovate\.json extends) or file:progress.ts createProgressProvider)`,
		output: autogold.Want("73", `{"Pattern":"extends","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["^renovate\\.json"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ (type:diff or type:commit) author:felix yarn`,
		output: autogold.Want("74", `{"Pattern":"yarn","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ (type:diff or type:commit) subscription after:"june 11 2019" before:"june 13 2019"`,
		output: autogold.Want("75", `{"Pattern":"subscription","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `(repo:^github\.com/sgtest/go-diff$@garo/lsif-indexing-campaign:test-already-exist-pr or repo:^github\.com/sgtest/sourcegraph-typescript$) file:README.md #`,
		output: autogold.Want("78", `{"Pattern":"#","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["README.md"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `(repo:^github\.com/sgtest/sourcegraph-typescript$ or repo:^github\.com/sgtest/go-diff$) package diff provides`,
		output: autogold.Want("79", `{"Pattern":"package diff provides","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:contains.file(path:noexist.go) test`,
		output: autogold.Want("83", `{"Pattern":"test","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:contains.file(path:go.mod) count:100 fmt`,
		output: autogold.Want("87", `{"Pattern":"fmt","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":100,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `type:commit LSIF`,
		output: autogold.Want("90", `{"Pattern":"LSIF","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:contains.file(path:diff.pb.go) type:commit LSIF`,
		output: autogold.Want("91", `{"Pattern":"LSIF","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:repo`,
		output: autogold.Want("93", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["repo"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:file`,
		output: autogold.Want("96", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["file"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:content`,
		output: autogold.Want("98", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["content"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize`,
		output: autogold.Want("99", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:commit`,
		output: autogold.Want("100", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["commit"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:symbol`,
		output: autogold.Want("101", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["symbol"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal type:symbol HunkNoChunksize select:symbol`,
		output: autogold.Want("102", `{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["symbol"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `foo\d "bar*" patterntype:regexp`,
		output: autogold.Want("105", `{"Pattern":"(?:foo\\d).*?(?:bar\\*)","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `patterntype:regexp // literal slash`,
		output: autogold.Want("107", `{"Pattern":"(?://).*?(?:literal).*?(?:slash)","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:contains.path(Dockerfile)`,
		output: autogold.Want("108", `{"Pattern":"","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repohasfile:Dockerfile`,
		output: autogold.Want("109", `{"Pattern":"","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}}

	test := func(input string) string {
		searchType := overrideSearchType(input, query.SearchTypeLiteral)
		plan, err := query.Pipeline(query.Init(input, searchType))
		if err != nil {
			return "Error"
		}
		if len(plan) == 0 {
			return "Empty"
		}
		b := plan[0]
		types, _ := b.ToParseTree().StringValues(query.FieldType)
		mode := search.Batch
		resultTypes := computeResultTypes(types, b, query.SearchTypeLiteral)
		p := toTextPatternInfo(b, resultTypes, mode)
		v, _ := json.Marshal(p)
		return string(v)
	}

	for _, tc := range cases {
		t.Run(tc.output.Name(), func(t *testing.T) {
			tc.output.Equal(t, test(tc.input))
		})
	}
}

func overrideSearchType(input string, searchType query.SearchType) query.SearchType {
	q, err := query.Parse(input, query.SearchTypeLiteral)
	q = query.LowercaseFieldNames(q)
	if err != nil {
		// If parsing fails, return the default search type. Any actual
		// parse errors will be raised by subsequent parser calls.
		return searchType
	}
	query.VisitField(q, "patterntype", func(value string, _ bool, _ query.Annotation) {
		switch value {
		case "regex", "regexp":
			searchType = query.SearchTypeRegex
		case "literal":
			searchType = query.SearchTypeLiteral
		case "structural":
			searchType = query.SearchTypeStructural
		}
	})
	return searchType
}

func TestRepoSubsetTextSearch(t *testing.T) {
	searcher.MockSearchFilesInRepo = func(ctx context.Context, repo types.MinimalRepo, gitserverRepo api.RepoName, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration, stream streaming.Sender) (limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo/one":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		case "foo/two":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		case "foo/empty":
			return false, nil
		case "foo/cloning":
			return false, &gitdomain.RepoNotExistError{Repo: repoName, CloneInProgress: true}
		case "foo/missing":
			return false, &gitdomain.RepoNotExistError{Repo: repoName}
		case "foo/missing-database":
			return false, &errcode.Mock{Message: "repo not found: foo/missing-database", IsNotFound: true}
		case "foo/timedout":
			return false, context.DeadlineExceeded
		case "foo/no-rev":
			// TODO we do not specify a rev when searching "foo/no-rev", so it
			// is treated as an empty repository. We need to test the fatal
			// case of trying to search a revision which doesn't exist.
			return false, &gitdomain.RevisionNotFoundError{Repo: repoName, Spec: "missing"}
		default:
			return false, errors.New("Unexpected repo")
		}
	}
	defer func() { searcher.MockSearchFilesInRepo = nil }()

	zoekt := &searchbackend.FakeSearcher{}

	q, err := query.ParseLiteral("foo")
	if err != nil {
		t.Fatal(err)
	}
	repoRevs := makeRepositoryRevisions("foo/one", "foo/two", "foo/empty", "foo/cloning", "foo/missing", "foo/missing-database", "foo/timedout", "foo/no-rev")

	patternInfo := &search.TextPatternInfo{
		FileMatchLimit: limits.DefaultMaxSearchResults,
		Pattern:        "foo",
	}

	matches, common, err := RunRepoSubsetTextSearch(
		context.Background(),
		logtest.Scoped(t),
		patternInfo,
		repoRevs,
		q,
		zoekt,
		endpoint.Static("test"),
		search.DefaultMode,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 {
		t.Errorf("expected two results, got %d", len(matches))
	}
	repoNames := map[api.RepoID]string{}
	for _, rr := range repoRevs {
		repoNames[rr.Repo.ID] = string(rr.Repo.Name)
	}
	assertReposStatus(t, repoNames, common.Status, map[string]search.RepoStatus{
		"foo/cloning":          search.RepoStatusCloning,
		"foo/missing":          search.RepoStatusMissing,
		"foo/missing-database": search.RepoStatusMissing,
		"foo/timedout":         search.RepoStatusTimedout,
	})

	// If we specify a rev and it isn't found, we fail the whole search since
	// that should be checked earlier.
	_, _, err = RunRepoSubsetTextSearch(
		context.Background(),
		logtest.Scoped(t),
		patternInfo,
		makeRepositoryRevisions("foo/no-rev@dev"),
		q,
		zoekt,
		endpoint.Static("test"),
		search.DefaultMode,
		false,
	)
	if !errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
		t.Fatalf("searching non-existent rev expected to fail with RevisionNotFoundError got: %v", err)
	}
}

func TestSearchFilesInReposStream(t *testing.T) {
	searcher.MockSearchFilesInRepo = func(ctx context.Context, repo types.MinimalRepo, gitserverRepo api.RepoName, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration, stream streaming.Sender) (limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo/one":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		case "foo/two":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		case "foo/three":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		default:
			return false, errors.New("Unexpected repo")
		}
	}
	defer func() { searcher.MockSearchFilesInRepo = nil }()

	zoekt := &searchbackend.FakeSearcher{}

	q, err := query.ParseLiteral("foo")
	if err != nil {
		t.Fatal(err)
	}

	patternInfo := &search.TextPatternInfo{
		FileMatchLimit: limits.DefaultMaxSearchResults,
		Pattern:        "foo",
	}

	matches, _, err := RunRepoSubsetTextSearch(
		context.Background(),
		logtest.Scoped(t),
		patternInfo,
		makeRepositoryRevisions("foo/one", "foo/two", "foo/three"),
		q,
		zoekt,
		endpoint.Static("test"),
		search.DefaultMode,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 3 {
		t.Errorf("expected three results, got %d", len(matches))
	}
}

func assertReposStatus(t *testing.T, repoNames map[api.RepoID]string, got search.RepoStatusMap, want map[string]search.RepoStatus) {
	t.Helper()
	gotM := map[string]search.RepoStatus{}
	got.Iterate(func(id api.RepoID, mask search.RepoStatus) {
		name := repoNames[id]
		if name == "" {
			name = fmt.Sprintf("UNKNOWNREPO{ID=%d}", id)
		}
		gotM[name] = mask
	})
	if diff := cmp.Diff(want, gotM); diff != "" {
		t.Errorf("RepoStatusMap mismatch (-want +got):\n%s", diff)
	}
}

func TestSearchFilesInRepos_multipleRevsPerRepo(t *testing.T) {
	searcher.MockSearchFilesInRepo = func(ctx context.Context, repo types.MinimalRepo, gitserverRepo api.RepoName, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration, stream streaming.Sender) (limitHit bool, err error) {
		repoName := repo.Name
		switch repoName {
		case "foo":
			stream.Send(streaming.SearchEvent{
				Results: []result.Match{&result.FileMatch{
					File: result.File{
						Repo:     repo,
						CommitID: api.CommitID(rev),
						Path:     "main.go",
					},
				}},
			})
			return false, nil
		default:
			panic("unexpected repo")
		}
	}
	defer func() { searcher.MockSearchFilesInRepo = nil }()

	trueVal := true
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{SearchMultipleRevisionsPerRepository: &trueVal},
	}})
	defer conf.Mock(nil)

	zoekt := &searchbackend.FakeSearcher{}

	q, err := query.ParseLiteral("foo")
	if err != nil {
		t.Fatal(err)
	}

	patternInfo := &search.TextPatternInfo{
		FileMatchLimit: limits.DefaultMaxSearchResults,
		Pattern:        "foo",
	}

	repos := makeRepositoryRevisions("foo@master:mybranch:branch3:branch4")

	matches, _, err := RunRepoSubsetTextSearch(
		context.Background(),
		logtest.Scoped(t),
		patternInfo,
		repos,
		q,
		zoekt,
		endpoint.Static("test"),
		search.DefaultMode,
		false,
	)
	if err != nil {
		t.Fatal(err)
	}

	matchKeys := make([]result.Key, len(matches))
	for i, match := range matches {
		matchKeys[i] = match.Key()
	}
	sort.Slice(matchKeys, func(i, j int) bool { return matchKeys[i].Less(matchKeys[j]) })

	wantResultKeys := []result.Key{
		{Repo: "foo", Commit: "branch3", Path: "main.go"},
		{Repo: "foo", Commit: "branch4", Path: "main.go"},
		{Repo: "foo", Commit: "master", Path: "main.go"},
		{Repo: "foo", Commit: "mybranch", Path: "main.go"},
	}
	require.Equal(t, wantResultKeys, matchKeys)
}

func TestZoektQueryPatternsAsRegexps(t *testing.T) {
	tests := []struct {
		name  string
		input zoektquery.Q
		want  []*regexp.Regexp
	}{
		{
			name:  "literal substring query",
			input: &zoektquery.Substring{Pattern: "foobar"},
			want:  []*regexp.Regexp{regexp.MustCompile(`(?i)foobar`)},
		},
		{
			name:  "regex query",
			input: &zoektquery.Regexp{Regexp: &syntax.Regexp{Op: syntax.OpLiteral, Name: "foobar"}},
			want:  []*regexp.Regexp{regexp.MustCompile(`(?i)` + zoektquery.Regexp{Regexp: &syntax.Regexp{Op: syntax.OpLiteral, Name: "foobar"}}.Regexp.String())},
		},
		{
			name: "and query",
			input: zoektquery.NewAnd([]zoektquery.Q{
				&zoektquery.Substring{Pattern: "foobar"},
				&zoektquery.Substring{Pattern: "baz"},
			}...),
			want: []*regexp.Regexp{
				regexp.MustCompile(`(?i)foobar`),
				regexp.MustCompile(`(?i)baz`),
			},
		},
		{
			name: "or query",
			input: zoektquery.NewOr([]zoektquery.Q{
				&zoektquery.Substring{Pattern: "foobar"},
				&zoektquery.Substring{Pattern: "baz"},
			}...),
			want: []*regexp.Regexp{
				regexp.MustCompile(`(?i)foobar`),
				regexp.MustCompile(`(?i)baz`),
			},
		},
		{
			name: "literal and regex",
			input: zoektquery.NewAnd([]zoektquery.Q{
				&zoektquery.Substring{Pattern: "foobar"},
				&zoektquery.Regexp{Regexp: &syntax.Regexp{Op: syntax.OpLiteral, Name: "python"}},
			}...),
			want: []*regexp.Regexp{
				regexp.MustCompile(`(?i)foobar`),
				regexp.MustCompile(`(?i)` + zoektquery.Regexp{Regexp: &syntax.Regexp{Op: syntax.OpLiteral, Name: "python"}}.Regexp.String()),
			},
		},
		{
			name: "literal or regex",
			input: zoektquery.NewOr([]zoektquery.Q{
				&zoektquery.Substring{Pattern: "foobar"},
				&zoektquery.Regexp{Regexp: &syntax.Regexp{Op: syntax.OpLiteral, Name: "python"}},
			}...),
			want: []*regexp.Regexp{
				regexp.MustCompile(`(?i)foobar`),
				regexp.MustCompile(`(?i)` + zoektquery.Regexp{Regexp: &syntax.Regexp{Op: syntax.OpLiteral, Name: "python"}}.Regexp.String()),
			},
		},
		{
			name:  "respect case sensitivity setting",
			input: &zoektquery.Substring{Pattern: "foo", CaseSensitive: true},
			want:  []*regexp.Regexp{regexp.MustCompile(regexp.QuoteMeta("foo"))},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := zoektQueryPatternsAsRegexps(tc.input)
			require.Equal(t, tc.want, got)
		})
	}
}

func makeRepositoryRevisions(repos ...string) []*search.RepositoryRevisions {
	r := make([]*search.RepositoryRevisions, len(repos))
	for i, repospec := range repos {
		repoName, revSpecs := search.ParseRepositoryRevisions(repospec)
		revs := make([]string, 0, len(revSpecs))
		for _, revSpec := range revSpecs {
			revs = append(revs, revSpec.RevSpec)
		}
		if len(revs) == 0 {
			// treat empty list as HEAD
			revs = []string{""}
		}
		r[i] = &search.RepositoryRevisions{Repo: mkRepos(repoName)[0], Revs: revs}
	}
	return r
}

func mkRepos(names ...string) []types.MinimalRepo {
	var repos []types.MinimalRepo
	for _, name := range names {
		sum := md5.Sum([]byte(name))
		id := api.RepoID(binary.BigEndian.Uint64(sum[:]))
		if id < 0 {
			id = -(id / 2)
		}
		if id == 0 {
			id++
		}
		repos = append(repos, types.MinimalRepo{ID: id, Name: api.RepoName(name)})
	}
	return repos
}

// RunRepoSubsetTextSearch is a convenience function that simulates the RepoSubsetTextSearch job.
func RunRepoSubsetTextSearch(
	ctx context.Context,
	logger log.Logger,
	patternInfo *search.TextPatternInfo,
	repos []*search.RepositoryRevisions,
	q query.Q,
	zoekt *searchbackend.FakeSearcher,
	searcherURLs *endpoint.Map,
	mode search.GlobalSearchMode,
	useFullDeadline bool,
) ([]*result.FileMatch, streaming.Stats, error) {
	notSearcherOnly := mode != search.SearcherOnly
	searcherArgs := &search.SearcherParameters{
		PatternInfo:     patternInfo,
		UseFullDeadline: useFullDeadline,
	}

	agg := streaming.NewAggregatingStream()

	indexed, unindexed, err := zoektutil.PartitionRepos(
		context.Background(),
		logger,
		repos,
		zoekt,
		search.TextRequest,
		query.Yes,
		query.ContainsRefGlobs(q),
	)
	if err != nil {
		return nil, streaming.Stats{}, err
	}

	g, ctx := errgroup.WithContext(ctx)

	if notSearcherOnly {
		b, err := query.ToBasicQuery(q)
		if err != nil {
			return nil, streaming.Stats{}, err
		}

		types, _ := q.StringValues(query.FieldType)
		var resultTypes result.Types
		if len(types) == 0 {
			resultTypes = result.TypeFile | result.TypePath | result.TypeRepo
		} else {
			for _, t := range types {
				resultTypes = resultTypes.With(result.TypeFromString[t])
			}
		}

		typ := search.TextRequest
		zoektQuery, err := zoektutil.QueryToZoektQuery(b, resultTypes, nil, typ)
		if err != nil {
			return nil, streaming.Stats{}, err
		}

		zoektJob := &zoektutil.RepoSubsetTextSearchJob{
			Repos:          indexed,
			Query:          zoektQuery,
			Typ:            search.TextRequest,
			FileMatchLimit: patternInfo.FileMatchLimit,
			Select:         patternInfo.Select,
			Since:          nil,
		}

		// Run literal and regexp searches on indexed repositories.
		g.Go(func() error {
			_, err := zoektJob.Run(ctx, job.RuntimeClients{
				Logger: logger,
				Zoekt:  zoekt,
			}, agg)
			return err
		})
	}

	// Concurrently run searcher for all unindexed repos regardless whether text or regexp.
	g.Go(func() error {
		searcherJob := &searcher.TextSearchJob{
			PatternInfo:     searcherArgs.PatternInfo,
			Repos:           unindexed,
			Indexed:         false,
			UseFullDeadline: searcherArgs.UseFullDeadline,
		}

		_, err := searcherJob.Run(ctx, job.RuntimeClients{
			Logger:       logger,
			SearcherURLs: searcherURLs,
			Zoekt:        zoekt,
		}, agg)
		return err
	})

	err = g.Wait()

	fms, fmErr := matchesToFileMatches(agg.Results)
	if fmErr != nil && err == nil {
		err = errors.Wrap(fmErr, "searchFilesInReposBatch failed to convert results")
	}
	return fms, agg.Stats, err
}

func matchesToFileMatches(matches []result.Match) ([]*result.FileMatch, error) {
	fms := make([]*result.FileMatch, 0, len(matches))
	for _, match := range matches {
		fm, ok := match.(*result.FileMatch)
		if !ok {
			return nil, errors.Errorf("expected only file match results")
		}
		fms = append(fms, fm)
	}
	return fms, nil
}
