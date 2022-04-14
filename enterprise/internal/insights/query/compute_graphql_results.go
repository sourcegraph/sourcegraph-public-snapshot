package query

import (
	"encoding/json"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ComputeResult interface {
	RepoName() string
	RepoID() string
	Revhash() string
	FilePath() string
	Counts() map[string]int
}

type GroupedResultsByRepository struct {
	RepoID      string
	RepoName    string
	MatchValues []string
}

type GroupedResults struct {
	Value string
	Count int
}

type TimeDataPoint struct {
	Time  time.Time
	Count int
}

const capturedValueMaxLength = 100

func GroupByCaptureMatch(results []ComputeResult) []GroupedResults {
	if len(results) < 1 {
		return nil
	}

	mapped := map[string]int{}
	for i := 0; i < len(results); i++ {
		currentCounts := results[i].Counts()
		for value, count := range currentCounts {
			if len(value) > capturedValueMaxLength {
				value = value[:capturedValueMaxLength]
			}
			mapped[value] += count
		}
	}

	grouped := make([]GroupedResults, 0, len(mapped))
	for value, count := range mapped {
		grouped = append(grouped, GroupedResults{
			Value: value,
			Count: count,
		})
	}
	return grouped
}

func GroupByRepository(results []ComputeResult) map[string][]ComputeResult {
	if len(results) < 1 {
		return nil
	}
	// map repository ID -> list of matches
	groupedbyRepo := make(map[string][]ComputeResult)
	for _, result := range results {
		groupedbyRepo[result.RepoID()] = append(groupedbyRepo[result.RepoID()], result)
	}
	return groupedbyRepo
}

func decodeComputeResult(result json.RawMessage) (ComputeResult, error) {
	typeName := struct {
		TypeName string `json:"__typeName"`
	}{}
	if err := json.Unmarshal(result, &typeName); err != nil {
		return nil, err
	}
	switch typeName.TypeName {
	case "ComputeMatchContext":
		var v ComputeMatchContext
		if err := json.Unmarshal(result, &v); err != nil {
			return nil, err
		}
		return &v, nil
	case "ComputeText":
		return nil, errors.Errorf("cannot decode search result: unsupported TypeName: %s", string(result))

	default:
		return nil, errors.Errorf("cannot decode search result: unexpected TypeName: %s", string(result))
	}
}

type ComputeMatchContext struct {
	Commit     string
	Repository struct {
		Name string
		Id   string
	}
	Path    string
	Matches []ComputeMatch
}

func (c ComputeMatchContext) RepoID() string {
	return c.Repository.Id
}

func (c ComputeMatchContext) Counts() map[string]int {
	distinct := make(map[string]int)
	for _, match := range c.Matches {
		for _, environment := range match.Environment {
			distinct[environment.Value] = distinct[environment.Value] + 1
		}
	}
	return distinct
}

func (c ComputeMatchContext) RepoName() string {
	return c.Repository.Name
}

func (c ComputeMatchContext) Revhash() string {
	return c.Commit
}

func (c ComputeMatchContext) FilePath() string {
	return c.Path
}

type ComputeMatch struct {
	Value       string
	Environment []ComputeEnvironmentEntry
}

type ComputeEnvironmentEntry struct {
	Variable string
	Value    string
}
