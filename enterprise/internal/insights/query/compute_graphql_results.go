package query

import (
	"time"
)

type ComputeResult interface {
	RepoName() string
	RepoID() string
	Revhash() string
	FilePath() string
	MatchValues() []string
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
	for _, value := range c.MatchValues() {
		distinct[value] = distinct[value] + 1
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

func (c ComputeMatchContext) MatchValues() []string {
	var results []string
	for _, match := range c.Matches {
		for _, entry := range match.Environment {
			results = append(results, entry.Value)
		}
	}
	return results
}

type ComputeMatch struct {
	Value       string
	Environment []ComputeEnvironmentEntry
}

type ComputeEnvironmentEntry struct {
	Variable string
	Value    string
}
