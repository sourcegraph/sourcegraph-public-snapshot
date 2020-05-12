package db

import (
	"github.com/keegancsmith/sqlf"
)

// intsToQueries converts a slice of ints into a slice of queries.
func intsToQueries(values []int) []*sqlf.Query {
	var queries []*sqlf.Query
	for _, value := range values {
		queries = append(queries, sqlf.Sprintf("%d", value))
	}

	return queries
}

// diff returns a slice containing the elements of left not present in right.
func diff(left, right []string) []string {
	rightSet := map[string]struct{}{}
	for _, v := range right {
		rightSet[v] = struct{}{}
	}

	var diff []string
	for _, v := range left {
		if _, ok := rightSet[v]; !ok {
			diff = append(diff, v)
		}
	}

	return diff
}
