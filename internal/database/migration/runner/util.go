package runner

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

func extractIDs(definitions []definition.Definition) []int {
	ids := make([]int, 0, len(definitions))
	for _, def := range definitions {
		ids = append(ids, def.ID)
	}

	return ids
}

func intSet(vs []int) map[int]struct{} {
	m := make(map[int]struct{}, len(vs))
	for _, v := range vs {
		m[v] = struct{}{}
	}

	return m
}

func intsToStrings(ints []int) []string {
	strs := make([]string, 0, len(ints))
	for _, value := range ints {
		strs = append(strs, strconv.Itoa(value))
	}

	return strs
}

func wait(ctx context.Context, duration time.Duration) error {
	select {
	case <-time.After(duration):
		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}
