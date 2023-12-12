package jobutil

import (
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp/syntax" //nolint:depguard // using the grafana fork of regexp clashes with zoekt, which uses the std regexp/syntax.
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"
	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"

	zoektquery "github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/api"
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
		want: autogold.Expect(`
(LOG
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
                  (indexed . false))))
            (REPOSEARCH
              (repoOpts.repoFilters . [foo])
              (repoOpts.searchContextSpec . @userA)
              (repoNamePatterns . [(?i)foo])))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.searchContextSpec . @userA))
          (PARALLEL
            NOOP
            NOOP))))))`),
	}, {
		query:      `foo context:global`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Expect(`
(LOG
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
            (ZOEKTGLOBALTEXTSEARCH
              (query . substr:"foo")
              (type . text)
              (repoOpts.searchContextSpec . global))
            (REPOSEARCH
              (repoOpts.repoFilters . [foo])
              (repoOpts.searchContextSpec . global)
              (repoNamePatterns . [(?i)foo])))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.searchContextSpec . global))
          NOOP)))))`),
	}, {
		query:      `foo`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Expect(`
(LOG
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
            (ZOEKTGLOBALTEXTSEARCH
              (query . substr:"foo")
              (type . text))
            (REPOSEARCH
              (repoOpts.repoFilters . [foo])
              (repoNamePatterns . [(?i)foo])))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `foo repo:sourcegraph/sourcegraph`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Expect(`
(LOG
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
              (repoOpts.repoFilters . [sourcegraph/sourcegraph])
              (PARTIALREPOS
                (ZOEKTREPOSUBSETTEXTSEARCH
                  (query . substr:"foo")
                  (type . text))))
            (REPOPAGER
              (repoOpts.repoFilters . [sourcegraph/sourcegraph])
              (PARTIALREPOS
                (SEARCHERTEXTSEARCH
                  (indexed . false))))
            (REPOSEARCH
              (repoOpts.repoFilters . [sourcegraph/sourcegraph foo])
              (repoNamePatterns . [(?i)sourcegraph/sourcegraph (?i)foo])))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.repoFilters . [sourcegraph/sourcegraph]))
          (PARALLEL
            NOOP
            NOOP))))))`),
	}, {
		query:      `ok ok`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            (ZOEKTGLOBALTEXTSEARCH
              (query . regex:"ok(?-s:.)*?ok")
              (type . text))
            (REPOSEARCH
              (repoOpts.repoFilters . [(?:ok).*?(?:ok)])
              (repoNamePatterns . [(?i)(?:ok).*?(?:ok)])))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `ok @thing`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLiteral,
		want: autogold.Expect(`
(LOG
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
            (ZOEKTGLOBALTEXTSEARCH
              (query . substr:"ok @thing")
              (type . text))
            (REPOSEARCH
              (repoOpts.repoFilters . [ok ])
              (repoNamePatterns . [(?i)ok ])))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `@nope`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            (type . text))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `repo:sourcegraph/sourcegraph rev:*refs/heads/*`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLucky,
		want: autogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originalQuery . )
    (patternType . lucky)
    (FEELINGLUCKYSEARCH
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            (REPOSCOMPUTEEXCLUDED
              (repoOpts.repoFilters . [sourcegraph/sourcegraph@*refs/heads/*]))
            (REPOSEARCH
              (repoOpts.repoFilters . [sourcegraph/sourcegraph@*refs/heads/*])
              (repoNamePatterns . [(?i)sourcegraph/sourcegraph]))))))))`),
	}, {
		query:      `repo:sourcegraph/sourcegraph@*refs/heads/*`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeLucky,
		want: autogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originalQuery . )
    (patternType . lucky)
    (FEELINGLUCKYSEARCH
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            (REPOSCOMPUTEEXCLUDED
              (repoOpts.repoFilters . [sourcegraph/sourcegraph@*refs/heads/*]))
            (REPOSEARCH
              (repoOpts.repoFilters . [sourcegraph/sourcegraph@*refs/heads/*])
              (repoNamePatterns . [(?i)sourcegraph/sourcegraph]))))))))`),
	}, {
		query:      `foo @bar`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            (type . text))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:symbol test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            (type . symbol))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:commit test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            (diff . false)
            (limit . 500)
            (repoOpts.onlyCloned . true))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:diff test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            (diff . true)
            (limit . 500)
            (repoOpts.onlyCloned . true))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:file type:commit test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            (type . text))
          (COMMITSEARCH
            (query . *protocol.MessageMatches(test))
            (diff . false)
            (limit . 500)
            (repoOpts.onlyCloned . true))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:file type:path type:repo type:commit type:symbol repo:test test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (ZOEKTREPOSUBSETTEXTSEARCH
                  (query . substr:"test")
                  (type . text))))
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (SEARCHERTEXTSEARCH
                  (indexed . false))))
            (REPOSEARCH
              (repoOpts.repoFilters . [test test])
              (repoNamePatterns . [(?i)test (?i)test])))
          (REPOPAGER
            (repoOpts.repoFilters . [test])
            (PARTIALREPOS
              (ZOEKTSYMBOLSEARCH
                (query . sym:substr:"test"))))
          (COMMITSEARCH
            (query . *protocol.MessageMatches(test))
            (diff . false)
            (limit . 500)
            (repoOpts.repoFilters . [test])
            (repoOpts.onlyCloned . true))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.repoFilters . [test]))
          (PARALLEL
            NOOP
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (SEARCHERSYMBOLSEARCH
                  (patternInfo.pattern . test)
                  (patternInfo.isRegexp . true)
                  (patternInfo.fileMatchLimit . 500)
                  (patternInfo.patternMatchesPath . true)
                  (numRepos . 0)
                  (limit . 500))))
            NOOP))))))`),
	}, {
		query:      `type:file type:commit test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            (type . text))
          (COMMITSEARCH
            (query . *protocol.MessageMatches(test))
            (diff . false)
            (limit . 500)
            (repoOpts.onlyCloned . true))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:file type:path type:repo type:commit type:symbol repo:test test`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (ZOEKTREPOSUBSETTEXTSEARCH
                  (query . substr:"test")
                  (type . text))))
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (SEARCHERTEXTSEARCH
                  (indexed . false))))
            (REPOSEARCH
              (repoOpts.repoFilters . [test test])
              (repoNamePatterns . [(?i)test (?i)test])))
          (REPOPAGER
            (repoOpts.repoFilters . [test])
            (PARTIALREPOS
              (ZOEKTSYMBOLSEARCH
                (query . sym:substr:"test"))))
          (COMMITSEARCH
            (query . *protocol.MessageMatches(test))
            (diff . false)
            (limit . 500)
            (repoOpts.repoFilters . [test])
            (repoOpts.onlyCloned . true))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.repoFilters . [test]))
          (PARALLEL
            NOOP
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (SEARCHERSYMBOLSEARCH
                  (patternInfo.pattern . test)
                  (patternInfo.isRegexp . true)
                  (patternInfo.fileMatchLimit . 500)
                  (patternInfo.patternMatchesPath . true)
                  (numRepos . 0)
                  (limit . 500))))
            NOOP))))))`),
	}, {
		query:      `(type:commit or type:diff) (a or b)`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		// TODO this output doesn't look right. There shouldn't be any zoekt or repo jobs
		want: autogold.Expect(`
(LOG
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
              (diff . false)
              (limit . 500)
              (repoOpts.onlyCloned . true))
            REPOSCOMPUTEEXCLUDED
            (OR
              NOOP
              NOOP))))
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            (DIFFSEARCH
              (query . (*protocol.DiffMatches((?:a)|(?:b))))
              (diff . true)
              (limit . 500)
              (repoOpts.onlyCloned . true))
            REPOSCOMPUTEEXCLUDED
            (OR
              NOOP
              NOOP)))))))`),
	}, {
		query:      `(type:repo a) or (type:file b)`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            REPOSCOMPUTEEXCLUDED
            (REPOSEARCH
              (repoOpts.repoFilters . [a])
              (repoNamePatterns . [(?i)a])))))
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            (ZOEKTGLOBALTEXTSEARCH
              (query . content_substr:"b")
              (type . text))
            REPOSCOMPUTEEXCLUDED
            NOOP))))))`),
	}, {
		query:      `type:symbol a or b`,
		protocol:   search.Streaming,
		searchType: query.SearchTypeRegex,
		want: autogold.Expect(`
(LOG
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
            (type . symbol))
          REPOSCOMPUTEEXCLUDED
          (OR
            NOOP
            NOOP))))))`),
	},
		{
			query:      `repo:contains.path(a) repo:contains.content(b)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Expect(`
(LOG
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
            (repoOpts.hasFileContent[0].path . a)
            (repoOpts.hasFileContent[1].content . b))
          (REPOSEARCH
            (repoOpts.hasFileContent[0].path . a)
            (repoOpts.hasFileContent[1].content . b)
            (repoNamePatterns . [])))))))`),
		}, {
			query:      `repo:contains.file(path:a content:b)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Expect(`
(LOG
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
            (repoOpts.hasFileContent[0].path . a)
            (repoOpts.hasFileContent[0].content . b))
          (REPOSEARCH
            (repoOpts.hasFileContent[0].path . a)
            (repoOpts.hasFileContent[0].content . b)
            (repoNamePatterns . [])))))))`),
		}, {
			query:      `repo:has(key:value)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Expect(`
(LOG
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
            (repoOpts.hasKVPs[0].key . key)
            (repoOpts.hasKVPs[0].value . value))
          (REPOSEARCH
            (repoOpts.hasKVPs[0].key . key)
            (repoOpts.hasKVPs[0].value . value)
            (repoNamePatterns . [])))))))`),
		}, {
			query:      `repo:has.tag(tag)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Expect(`
(LOG
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
            (repoOpts.hasKVPs[0].key . tag)
            (repoNamePatterns . [])))))))`),
		}, {
			query:      `repo:has.topic(mytopic)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Expect(`
(LOG
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
            (repoOpts.hasTopics[0].topic . mytopic))
          (REPOSEARCH
            (repoOpts.hasTopics[0].topic . mytopic)
            (repoNamePatterns . [])))))))`),
		}, {
			query:      `repo:has.tag(tag) foo`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeRegex,
			want: autogold.Expect(`
(LOG
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
              (repoOpts.hasKVPs[0].key . tag)
              (PARTIALREPOS
                (ZOEKTREPOSUBSETTEXTSEARCH
                  (query . substr:"foo")
                  (type . text))))
            (REPOPAGER
              (repoOpts.hasKVPs[0].key . tag)
              (PARTIALREPOS
                (SEARCHERTEXTSEARCH
                  (indexed . false))))
            (REPOSEARCH
              (repoOpts.repoFilters . [foo])
              (repoOpts.hasKVPs[0].key . tag)
              (repoNamePatterns . [(?i)foo])))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.hasKVPs[0].key . tag))
          (PARALLEL
            NOOP
            NOOP))))))`),
		}, {
			query:      `(...)`,
			protocol:   search.Streaming,
			searchType: query.SearchTypeStructural,
			want: autogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originalQuery . )
    (patternType . structural)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          REPOSCOMPUTEEXCLUDED
          (STRUCTURALSEARCH
            (patternInfo.pattern . (:[_]))
            (patternInfo.isStructural . true)
            (patternInfo.fileMatchLimit . 500)))))))`),
		},
	}

	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
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

	autogold.Expect(`
(REPOSEARCH
  (repoOpts.repoFilters . [foo])
  (repoNamePatterns . [(?i)foo]))
`).Equal(t, test("foo", search.Streaming))

	autogold.Expect(`
(REPOSEARCH
  (repoOpts.repoFilters . [foo])
  (repoNamePatterns . [(?i)foo]))
`).Equal(t, test("foo", search.Batch))
}

func TestToTextPatternInfo(t *testing.T) {
	cases := []struct {
		input  string
		output autogold.Value
	}{{
		input:  `type:repo archived`,
		output: autogold.Expect(`{"Pattern":"archived","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `type:repo archived archived:yes`,
		output: autogold.Expect(`{"Pattern":"archived","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `type:repo sgtest/mux`,
		output: autogold.Expect(`{"Pattern":"sgtest/mux","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `type:repo sgtest/mux fork:yes`,
		output: autogold.Expect(`{"Pattern":"sgtest/mux","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `"func main() {\n" patterntype:regexp type:file`,
		output: autogold.Expect(`{"Pattern":"func main\\(\\) \\{\n","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `"func main() {\n" -repo:go-diff patterntype:regexp type:file`,
		output: autogold.Expect(`{"Pattern":"func main\\(\\) \\{\n","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ String case:yes type:file`,
		output: autogold.Expect(`{"Pattern":"String","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":true,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":true,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$@v1 void sendPartialResult(Object requestId, JsonPatch jsonPatch); patterntype:literal type:file`,
		output: autogold.Expect(`{"Pattern":"void sendPartialResult\\(Object requestId, JsonPatch jsonPatch\\);","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$@v1 void sendPartialResult(Object requestId, JsonPatch jsonPatch); patterntype:literal count:1 type:file`,
		output: autogold.Expect(`{"Pattern":"void sendPartialResult\\(Object requestId, JsonPatch jsonPatch\\);","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":1,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$ \nimport index:only patterntype:regexp type:file`,
		output: autogold.Expect(`{"Pattern":"\\nimport","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"only","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$ \nimport index:no patterntype:regexp type:file`,
		output: autogold.Expect(`{"Pattern":"\\nimport","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"no","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/java-langserver$ doesnot734734743734743exist`,
		output: autogold.Expect(`{"Pattern":"doesnot734734743734743exist","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ type:commit test`,
		output: autogold.Expect(`{"Pattern":"test","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ type:diff main`,
		output: autogold.Expect(`{"Pattern":"main","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ repohascommitafter:"2019-01-01" test patterntype:literal`,
		output: autogold.Expect(`{"Pattern":"test","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `^func.*$ patterntype:regexp index:only type:file`,
		output: autogold.Expect(`{"Pattern":"^func.*$","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"only","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `fork:only patterntype:regexp FORK_SENTINEL`,
		output: autogold.Expect(`{"Pattern":"FORK_SENTINEL","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `\bfunc\b lang:go type:file patterntype:regexp`,
		output: autogold.Expect(`{"Pattern":"\\bfunc\\b","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["\\.go$"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":false,"Languages":["go"]}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ make(:[1]) index:only patterntype:structural count:3`,
		output: autogold.Expect(`{"Pattern":"make(:[1])","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":3,"Index":"only","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ make(:[1]) lang:go rule:'where "backcompat" == "backcompat"' patterntype:structural`,
		output: autogold.Expect(`{"Pattern":"make(:[1])","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"where \"backcompat\" == \"backcompat\"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["\\.go$"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":["go"]}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$@adde71 make(:[1]) index:no patterntype:structural count:3`,
		output: autogold.Expect(`{"Pattern":"make(:[1])","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":3,"Index":"no","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ file:^README\.md "basic :[_] access :[_]" patterntype:structural`,
		output: autogold.Expect(`{"Pattern":"\"basic :[_] access :[_]\"","IsNegated":false,"IsRegExp":false,"IsStructuralPat":true,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["^README\\.md"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `no results for { ... } raises alert repo:^github\.com/sgtest/go-diff$`,
		output: autogold.Expect(`{"Pattern":"no results for \\{ \\.\\.\\. \\} raises alert","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ patternType:regexp \ and /`,
		output: autogold.Expect(`{"Pattern":"(?:\\ and).*?(?:/)","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ (not .svg) patterntype:literal`,
		output: autogold.Expect(`{"Pattern":"\\.svg","IsNegated":true,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ (Fetches OR file:language-server.ts)`,
		output: autogold.Expect(`{"Pattern":"Fetches","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ ((file:^renovate\.json extends) or file:progress.ts createProgressProvider)`,
		output: autogold.Expect(`{"Pattern":"extends","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["^renovate\\.json"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ (type:diff or type:commit) author:felix yarn`,
		output: autogold.Expect(`{"Pattern":"yarn","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegraph-typescript$ (type:diff or type:commit) subscription after:"june 11 2019" before:"june 13 2019"`,
		output: autogold.Expect(`{"Pattern":"subscription","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `(repo:^github\.com/sgtest/go-diff$@garo/lsif-indexing-campaign:test-already-exist-pr or repo:^github\.com/sgtest/sourcegraph-typescript$) file:README.md #`,
		output: autogold.Expect(`{"Pattern":"#","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":["README.md"],"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `(repo:^github\.com/sgtest/sourcegraph-typescript$ or repo:^github\.com/sgtest/go-diff$) package diff provides`,
		output: autogold.Expect(`{"Pattern":"package diff provides","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:contains.file(path:noexist.go) test`,
		output: autogold.Expect(`{"Pattern":"test","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:contains.file(path:go.mod) count:100 fmt`,
		output: autogold.Expect(`{"Pattern":"fmt","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":100,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `type:commit LSIF`,
		output: autogold.Expect(`{"Pattern":"LSIF","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:contains.file(path:diff.pb.go) type:commit LSIF`,
		output: autogold.Expect(`{"Pattern":"LSIF","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:repo`,
		output: autogold.Expect(`{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["repo"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:file`,
		output: autogold.Expect(`{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["file"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:content`,
		output: autogold.Expect(`{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["content"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize`,
		output: autogold.Expect(`{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:commit`,
		output: autogold.Expect(`{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["commit"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal HunkNoChunksize select:symbol`,
		output: autogold.Expect(`{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["symbol"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:go-diff patterntype:literal type:symbol HunkNoChunksize select:symbol`,
		output: autogold.Expect(`{"Pattern":"HunkNoChunksize","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":["symbol"],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":false,"PatternMatchesPath":false,"Languages":null}`),
	}, {
		input:  `foo\d "bar*" patterntype:regexp`,
		output: autogold.Expect(`{"Pattern":"(?:foo\\d).*?(?:bar\\*)","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `patterntype:regexp // literal slash`,
		output: autogold.Expect(`{"Pattern":"(?://).*?(?:literal).*?(?:slash)","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repo:contains.path(Dockerfile)`,
		output: autogold.Expect(`{"Pattern":"","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
	}, {
		input:  `repohasfile:Dockerfile`,
		output: autogold.Expect(`{"Pattern":"","IsNegated":false,"IsRegExp":true,"IsStructuralPat":false,"CombyRule":"","IsWordMatch":false,"IsCaseSensitive":false,"FileMatchLimit":30,"Index":"yes","Select":[],"IncludePatterns":null,"ExcludePattern":"","PathPatternsAreCaseSensitive":false,"PatternMatchesContent":true,"PatternMatchesPath":true,"Languages":null}`),
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
		resultTypes := computeResultTypes(b, query.SearchTypeLiteral)
		p := toTextPatternInfo(b, resultTypes, limits.DefaultMaxSearchResults)
		v, _ := json.Marshal(p)
		return string(v)
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
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

func Test_computeResultTypes(t *testing.T) {
	test := func(input string, searchType query.SearchType) string {
		plan, _ := query.Pipeline(query.Init(input, searchType))
		b := plan[0]
		resultTypes := computeResultTypes(b, searchType)
		return resultTypes.String()
	}

	t.Run("standard, only search file content when type not set", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test("path:foo content:bar", query.SearchTypeStandard)))
	})

	t.Run("standard, plain pattern searches repo path file content", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test("path:foo bar", query.SearchTypeStandard)))
	})

	t.Run("newStandardRC1, only search file content when type not set", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test("path:foo content:bar", query.SearchTypeNewStandardRC1)))
	})

	t.Run("newStandardRC1, plain pattern searches repo path file content", func(t *testing.T) {
		autogold.ExpectFile(t, autogold.Raw(test("path:foo bar", query.SearchTypeNewStandardRC1)))
	})
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

	zoekt := &searchbackend.FakeStreamer{}

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

	zoekt := &searchbackend.FakeStreamer{}

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

	zoekt := &searchbackend.FakeStreamer{}

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
	slices.SortFunc(matchKeys, result.Key.Less)

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
		repoRevs, err := query.ParseRepositoryRevisions(repospec)
		if err != nil {
			panic(errors.Errorf("unexpected error parsing repo spec %s", repospec))
		}

		revs := make([]string, 0, len(repoRevs.Revs))
		for _, revSpec := range repoRevs.Revs {
			revs = append(revs, revSpec.RevSpec)
		}
		if len(revs) == 0 {
			// treat empty list as HEAD
			revs = []string{""}
		}
		r[i] = &search.RepositoryRevisions{Repo: mkRepos(repoRevs.Repo)[0], Revs: revs}
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
	zoekt *searchbackend.FakeStreamer,
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

		fieldTypes, _ := q.StringValues(query.FieldType)
		var resultTypes result.Types
		if len(fieldTypes) == 0 {
			resultTypes = result.TypeFile | result.TypePath | result.TypeRepo
		} else {
			for _, t := range fieldTypes {
				resultTypes = resultTypes.With(result.TypeFromString[t])
			}
		}

		typ := search.TextRequest
		zoektQuery, err := zoektutil.QueryToZoektQuery(b, resultTypes, &search.Features{}, typ)
		if err != nil {
			return nil, streaming.Stats{}, err
		}

		zoektParams := &search.ZoektParameters{
			FileMatchLimit: patternInfo.FileMatchLimit,
			Select:         patternInfo.Select,
			// TODO: numContextLines
		}

		zoektJob := &zoektutil.RepoSubsetTextSearchJob{
			Repos:       indexed,
			Query:       zoektQuery,
			Typ:         search.TextRequest,
			ZoektParams: zoektParams,
			Since:       nil,
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
			NumContextLines: searcherArgs.NumContextLines,
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
