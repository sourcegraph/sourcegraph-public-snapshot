package jobutil

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewFeelingLuckySearchJob(t *testing.T) {
	test := func(q string) string {
		inputs := &run.SearchInputs{
			UserSettings: &schema.Settings{},
			Protocol:     search.Streaming,
			PatternType:  query.SearchTypeLucky,
		}
		var j job.Job
		plan, _ := query.Pipeline(query.InitLiteral(q))
		fj := NewFeelingLuckySearchJob(nil, inputs, plan)
		generated := []job.Job{}

		for _, next := range fj.generators {
			for {
				j, next = next()
				if j == nil {
					if next == nil {
						// No job and generator is exhausted.
						break
					}
					continue
				}
				generated = append(generated, j)
				if next == nil {
					break
				}
			}
		}
		return PrettyJSONVerbose(NewOrJob(generated...))
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
}
