package query

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
)

type ComputeResult interface {
	RepoName() string
	Revhash() string
	FilePath() string
	MatchValues() []string
	Counts() map[string]int
}

type GroupedResults struct {
	Value string
	Count int
}

type TimeDataPoint struct {
	Time  time.Time
	Count int
}

func GroupByCaptureMatch(results []ComputeResult) []GroupedResults {
	if len(results) < 1 {
		return nil
	}

	mapped := results[0].Counts()
	for i := 1; i < len(results); i++ {
		currentCounts := results[i].Counts()
		for value, count := range currentCounts {
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

func decodeComputeResult(result json.RawMessage) (ComputeResult, error) {
	typeName := struct {
		TypeName string `json:"__typeName"`
	}{}
	if err := json.Unmarshal(result, &typeName); err != nil {
		return nil, err
	}
	switch typeName.TypeName {
	case "ComputeMatchContext":
		var v computeMatchContext
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

type computeMatchContext struct {
	Commit     string
	Repository struct {
		Name string
	}
	Path    string
	Matches []computeMatch
}

func (c computeMatchContext) Counts() map[string]int {
	distinct := make(map[string]int)
	for _, value := range c.MatchValues() {
		distinct[value] = distinct[value] + 1
	}
	return distinct
}

func (c computeMatchContext) RepoName() string {
	return c.Repository.Name
}

func (c computeMatchContext) Revhash() string {
	return c.Commit
}

func (c computeMatchContext) FilePath() string {
	return c.Path
}

func (c computeMatchContext) MatchValues() []string {
	var results []string
	for _, match := range c.Matches {
		for _, entry := range match.Environment {
			results = append(results, entry.Value)
		}
	}
	return results
}

type computeMatch struct {
	Value       string
	Environment []computeEnvironmentEntry
}

type computeEnvironmentEntry struct {
	Variable string
	Value    string
}
