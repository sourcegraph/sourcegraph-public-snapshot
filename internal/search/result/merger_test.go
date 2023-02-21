package result

import (
	"math/rand"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func mkFileMatch(repo types.MinimalRepo, path string, lineNumbers ...int) Match {
	var hms ChunkMatches
	for _, n := range lineNumbers {
		hms = append(hms, ChunkMatch{
			Ranges: []Range{{
				Start: Location{Line: n},
				End:   Location{Line: n},
			}},
		})
	}

	return &FileMatch{
		File: File{
			Path: path,
			Repo: repo,
		},
		ChunkMatches: hms,
	}
}

func TestMerger(t *testing.T) {
	sources := 3
	m := NewMerger(sources)
	repo := types.MinimalRepo{Name: "r"}

	sourcedMatch := []struct {
		match  Match
		source int
	}{
		// all sources
		{mkFileMatch(repo, "all_sources", 1), 0},
		{mkFileMatch(repo, "all_sources", 1), 1},
		{mkFileMatch(repo, "all_sources", 1), 2},
		// 2 sources
		{mkFileMatch(repo, "2_of_3", 1), 0},
		{mkFileMatch(repo, "2_of_3", 1), 1}, // should be deduped by merger
		// 1 source
		{mkFileMatch(repo, "1_of_3", 1), 0},
		{mkFileMatch(repo, "1_of_3_other", 1), 1},
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(sourcedMatch), func(i, j int) {
		sourcedMatch[i], sourcedMatch[j] = sourcedMatch[j], sourcedMatch[i]
	})

	for _, sm := range sourcedMatch {
		m.addMatch(sm.match, sm.source)
	}

	unsent := m.UnsentTracked()

	// all matches seen by a subset of sources minus deduped results.
	wantUnsent := 3
	if gotUnsent := len(unsent); gotUnsent != wantUnsent {
		t.Fatalf("len(unsent): wanted %d, got %d", wantUnsent, gotUnsent)
	}

	wantPath := "2_of_3"
	if gotPath := unsent[0].(*FileMatch).Path; gotPath != wantPath {
		t.Fatalf("best unsent match: want %s, got %s", wantPath, gotPath)
	}
}
