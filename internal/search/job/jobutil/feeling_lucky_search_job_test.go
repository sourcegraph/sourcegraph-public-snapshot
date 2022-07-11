package jobutil

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/search"
	alertobserver "github.com/sourcegraph/sourcegraph/internal/search/alert"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/mockjob"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/require"
)

func TestNewFeelingLuckySearchJob(t *testing.T) {
	test := func(q string) string {
		inputs := &run.SearchInputs{
			UserSettings: &schema.Settings{},
			Protocol:     search.Streaming,
			PatternType:  query.SearchTypeLucky,
		}
		plan, _ := query.Pipeline(query.InitLiteral(q))
		fj := NewFeelingLuckySearchJob(nil, inputs, plan)
		var autoQ *autoQuery
		type want struct {
			Description string
			Query       string
		}
		generated := []want{}

		for _, next := range fj.generators {
			for {
				autoQ, next = next()
				if autoQ == nil {
					if next == nil {
						// No job and generator is exhausted.
						break
					}
					continue
				}
				generated = append(generated, want{Description: autoQ.description, Query: query.StringHuman(autoQ.query.ToParseTree())})
				if next == nil {
					break
				}
			}
		}
		result, _ := json.MarshalIndent(generated, "", "  ")
		return string(result)
	}

	t.Run("trigger unquoted rule", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`repo:^github\.com/sourcegraph/sourcegraph$ "monitor" "*Monitor"`)))
	})

	t.Run("trigger unordered patterns", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`context:global parse func`)))
	})

	t.Run("two basic jobs", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`context:global ((type:file parse func) or (type:commit parse func))`)))
	})

	t.Run("single pattern as lang", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`context:global python`)))
	})

	t.Run("one of many patterns as lang", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`context:global parse python`)))
	})

	t.Run("pattern as type", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`context:global fix commit`)))
	})

	t.Run("pattern as type multi patterns", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`context:global code monitor commit`)))
	})

	t.Run("pattern as type with expression", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`context:global code or monitor commit`)))
	})

	t.Run("type and lang multi rule", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`context:global go commit monitor code`)))
	})
}

func TestNewFeelingLuckySearchJob_Run(t *testing.T) {
	// Setup: A child job that sends the same result
	mockJob := mockjob.NewMockJob()
	mockJob.RunFunc.SetDefaultHook(func(ctx context.Context, _ job.RuntimeClients, s streaming.Sender) (*search.Alert, error) {
		s.Send(streaming.SearchEvent{
			Results: []result.Match{&result.FileMatch{
				File: result.File{Path: "haut-medoc"},
			}},
		})
		return nil, nil
	})

	mockAutoQuery := &autoQuery{description: "mock", query: query.Basic{}}

	j := FeelingLuckySearchJob{
		initialJob: mockJob,
		generators: []next{func() (*autoQuery, next) { return mockAutoQuery, nil }},
		newGeneratedJob: func(*autoQuery) job.Job {
			return mockJob
		},
	}

	var sent []result.Match
	stream := streaming.StreamFunc(func(e streaming.SearchEvent) {
		sent = append(sent, e.Results...)
	})

	t.Run("deduplicate results returned by generated jobs", func(t *testing.T) {
		j.Run(context.Background(), job.RuntimeClients{}, stream)
		require.Equal(t, 1, len(sent))
	})
}

func TestGeneratedSearchJob(t *testing.T) {
	mockJob := mockjob.NewMockJob()
	setMockJobResultSize := func(n int) {
		mockJob.RunFunc.SetDefaultHook(func(ctx context.Context, _ job.RuntimeClients, s streaming.Sender) (*search.Alert, error) {
			for i := 0; i < n; i++ {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
				}
				s.Send(streaming.SearchEvent{
					Results: []result.Match{&result.FileMatch{
						File: result.File{Path: strconv.Itoa(i)},
					}},
				})
			}
			return nil, nil
		})
	}

	test := func(resultSize int) string {
		setMockJobResultSize(resultSize)
		inputs := &run.SearchInputs{
			UserSettings: &schema.Settings{},
			Protocol:     search.Streaming,
			PatternType:  query.SearchTypeLucky,
		}

		q, _ := query.ParseStandard("test")
		mockQuery, _ := query.ToBasicQuery(q)
		j, _ := NewGeneratedSearchJob(inputs, &autoQuery{description: "test", query: mockQuery})
		j.(*generatedSearchJob).Child = mockJob
		_, err := j.Run(context.Background(), job.RuntimeClients{}, streaming.NewAggregatingStream())
		if err == nil {
			return ""
		}
		return err.(*alertobserver.ErrLuckyQueries).ProposedQueries[0].Description
	}

	autogold.Want("0 results", autogold.Raw("")).Equal(t, autogold.Raw(test(0)))
	autogold.Want("1 result", autogold.Raw("test (1 result)")).Equal(t, autogold.Raw(test(1)))
	autogold.Want("limit results", autogold.Raw("test (500+ results)")).Equal(t, autogold.Raw(test(limits.DefaultMaxSearchResultsStreaming)))
}

func TestCombinations(t *testing.T) {
	test := func(input string, rulesNarrow, rulesWiden []rule) string {
		q, _ := query.ParseStandard(input)
		b, _ := query.ToBasicQuery(q)
		g := NewGenerator(b, rulesNarrow, rulesWiden)

		var autoQ *autoQuery
		type want struct {
			Description string
			Query       string
		}
		generated := []want{}

		for {
			autoQ, g = g()
			if autoQ != nil {
				generated = append(
					generated,
					want{
						Description: autoQ.description,
						Query:       query.StringHuman(autoQ.query.ToParseTree()),
					})
			}

			if g == nil {
				break
			}
		}

		result, _ := json.MarshalIndent(generated, "", "  ")
		return string(result)
	}

	t.Run("narrow and widen rules", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`go commit yikes derp`, rulesNarrow, rulesWiden)))
	})

	t.Run("only narrow rules", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`go commit yikes derp`, rulesNarrow, nil)))
	})

	t.Run("only widen rules", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`go commit yikes derp`, nil, rulesWiden)))
	})
}

func Test_patternsAsFilters(t *testing.T) {
	test := func(input string, rules []rule) string {
		q, _ := query.ParseStandard(input)
		b, _ := query.ToBasicQuery(q)
		g := NewGenerator(b, nil, rules)

		var autoQ *autoQuery
		type want struct {
			Description string
			Input       string
			Query       string
		}
		generated := []want{}

		for {
			autoQ, g = g()
			if autoQ != nil {
				generated = append(
					generated,
					want{
						Description: autoQ.description,
						Input:       input,
						Query:       query.StringHuman(autoQ.query.ToParseTree()),
					})
			}

			if g == nil {
				break
			}
		}

		result, _ := json.MarshalIndent(generated, "", "  ")
		return string(result)
	}

	rules := []rule{
		{
			description: "patterns to code host filters",
			transform:   transform{patternsToCodeHostFilters},
		},
	}

	t.Run("URL pattern as fully qualified repo filter", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`https://github.com/sourcegraph/sourcegraph`, rules)))
	})

	t.Run("URL pattern as partially qualified repo filter", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`https://github.com/sourcegraph`, rules)))
	})

	t.Run("schemaless URL pattern as repo filter", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`github.com/sourcegraph`, rules)))
	})

	t.Run("URL blob", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`https://github.com/sourcegraph/sourcegraph/blob/main/lib/README.md#L50`, rules)))
	})

	t.Run("URL tree path", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`https://github.com/sourcegraph/sourcegraph/tree/main/lib`, rules)))
	})

	t.Run("URL tree branch revision", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`https://github.com/sourcegraph/sourcegraph/tree/2.12`, rules)))
	})

	t.Run("URL tree commit revision", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`https://github.com/sourcegraph/sourcegraph/commit/abc`, rules)))
	})
}

func Test_regexpPatterns(t *testing.T) {
	test := func(input string, rules []rule) string {
		q, _ := query.ParseStandard(input)
		b, _ := query.ToBasicQuery(q)
		g := NewGenerator(b, nil, rules)

		var autoQ *autoQuery
		type want struct {
			Description string
			Input       string
			Query       string
		}
		generated := []want{}

		for {
			autoQ, g = g()
			if autoQ != nil {
				generated = append(
					generated,
					want{
						Description: autoQ.description,
						Input:       input,
						Query:       query.StringHuman(autoQ.query.ToParseTree()),
					})
			}

			if g == nil {
				break
			}
		}

		result, _ := json.MarshalIndent(generated, "", "  ")
		return string(result)
	}

	rules := []rule{
		{
			description: "patterns as regular expressions",
			transform:   transform{regexpPatterns},
		},
	}

	t.Run("valid regular expression", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`[a-z]+`, rules)))
	})

	t.Run("valid regular expression", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`a.*b`, rules)))
	})

	t.Run("valid regular expression with capture group", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`(ab)*`, rules)))
	})

	t.Run("invalid regular expression", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`c++`, rules)))
	})

	t.Run("pattern without enough regexp syntax", func(t *testing.T) {
		autogold.Equal(t, autogold.Raw(test(`my.yaml.conf`, rules)))
	})
}
