package resolvers

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func TestValidateLineChartSearchInsightInput(t *testing.T) {
	btrue := true
	bfalse := false
	t.Run("should pass validation with capture groups", func(t *testing.T) {
		input := graphqlbackend.LineChartSearchInsightDataSeriesInput{
			Query: "",
			TimeScope: graphqlbackend.TimeScopeInput{StepInterval: &graphqlbackend.TimeIntervalStepInput{
				Unit:  "MONTH",
				Value: 1,
			}},
			RepositoryScope: graphqlbackend.RepositoryScopeInput{Repositories: []string{"github.com/mycompany/myrepo"}},
			Options: graphqlbackend.LineChartDataSeriesOptionsInput{
				Label:     addrStr("mylabel"),
				LineColor: addrStr("red"),
			},
			GeneratedFromCaptureGroups: &btrue,
		}

		err := validateLineChartSearchInsightInput(input)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("should pass validation without capture groups", func(t *testing.T) {
		input := graphqlbackend.LineChartSearchInsightDataSeriesInput{
			Query: "",
			TimeScope: graphqlbackend.TimeScopeInput{StepInterval: &graphqlbackend.TimeIntervalStepInput{
				Unit:  "MONTH",
				Value: 1,
			}},
			RepositoryScope: graphqlbackend.RepositoryScopeInput{Repositories: []string{}},
			Options: graphqlbackend.LineChartDataSeriesOptionsInput{
				Label:     addrStr("mylabel"),
				LineColor: addrStr("red"),
			},
			GeneratedFromCaptureGroups: &bfalse,
		}

		err := validateLineChartSearchInsightInput(input)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("fails because global capture groups", func(t *testing.T) {
		input := graphqlbackend.LineChartSearchInsightDataSeriesInput{
			Query: "",
			TimeScope: graphqlbackend.TimeScopeInput{StepInterval: &graphqlbackend.TimeIntervalStepInput{
				Unit:  "MONTH",
				Value: 1,
			}},
			RepositoryScope: graphqlbackend.RepositoryScopeInput{Repositories: []string{}},
			Options: graphqlbackend.LineChartDataSeriesOptionsInput{
				Label:     addrStr("mylabel"),
				LineColor: addrStr("red"),
			},
			GeneratedFromCaptureGroups: &btrue,
		}

		err := validateLineChartSearchInsightInput(input)
		if err == nil {
			t.Error(err)
		}
	})
}

func addrStr(input string) *string {
	return &input
}
