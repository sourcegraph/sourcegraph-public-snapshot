package store

import (
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
)

func makePatternCondition(patterns []string, defaultValue bool) *sqlf.Query {
	if len(patterns) == 0 {
		if defaultValue {
			return sqlf.Sprintf("TRUE")
		}

		return sqlf.Sprintf("FALSE")
	}

	conds := make([]*sqlf.Query, 0, len(patterns))
	for _, pattern := range patterns {
		conds = append(conds, sqlf.Sprintf("lower(name) LIKE %s", strings.ToLower(strings.ReplaceAll(pattern, "*", "%"))))
	}

	return sqlf.Join(conds, "OR")
}

func optionalLimit(limit *int) *sqlf.Query {
	if limit != nil {
		return sqlf.Sprintf("LIMIT %d", *limit)
	}

	return sqlf.Sprintf("")
}

func optionalArray[T any](values *[]T) any {
	if values != nil {
		return pq.Array(*values)
	}

	return nil
}

func optionalNumHours(duration *time.Duration) *int {
	if duration != nil {
		v := int(*duration / time.Hour)
		return &v
	}

	return nil
}

func optionalDuration(numHours *int) *time.Duration {
	if numHours != nil {
		v := time.Duration(*numHours) * time.Hour
		return &v
	}

	return nil
}

func optionalSlice[T any](s []T) *[]T {
	if len(s) != 0 {
		return &s
	}

	return nil
}
