package compression

import (
	"context"
	"testing"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/google/go-cmp/cmp"
)

func TestFilterFrames(t *testing.T) {

	ctx := context.Background()

	maxHistorical := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	commitFilter := CommitFilter{
		maxHistorical: maxHistorical,
	}

	t.Run("test empty frames", func(t *testing.T) {
		want := []Frame{}
		got := commitFilter.FilterFrames(ctx, []Frame{}, 1)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpeted frames filtered from empty input: %v", diff)
		}
	})

	t.Run("test one frame", func(t *testing.T) {
		want := []Frame{{
			maxHistorical, maxHistorical.Add(time.Second * 500), "abcdef",
		}}
		got := commitFilter.FilterFrames(ctx, want, 1)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpeted frames filtered from single input: %v", diff)
		}
	})

	t.Run("test unable to fetch metadata", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		want := []Frame{{
			maxHistorical, maxHistorical.Add(time.Second * 500), "abcdef",
		}, {
			maxHistorical, maxHistorical.Add(time.Second * 500), "fedcba",
		}}

		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{}, errors.New("really bad error"))

		got := commitFilter.FilterFrames(ctx, want, 1)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpeted frames when metadata is unavailable: %v", diff)
		}

	})

	t.Run("test no commits two frames", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []Frame{{
			maxHistorical, maxHistorical.Add(time.Second * 500), "abcdef",
		}, {
			maxHistorical, maxHistorical.Add(time.Second * 500), "fedcba",
		}}

		want := []Frame{{
			maxHistorical, maxHistorical.Add(time.Second * 500), "abcdef",
		}}

		got := commitFilter.FilterFrames(ctx, input, 1)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpeted frames when metadata is unavailable: %v", diff)
		}
	})

	t.Run("test three frames middle has no commits", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []Frame{{
			toTime("2020-05-01"), toTime("2020-06-01"), "abcdef",
		}, {
			toTime("2020-06-01"), toTime("2020-07-01"), "fedcba",
		}, {
			toTime("2020-07-01"), toTime("2020-08-01"), "111222333",
		}}

		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:        1,
			Enabled:       true,
			LastIndexedAt: toTime("2021-01-01"),
		}, nil)

		// The middle commit will actually be the first one to call Get
		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)

		want := []Frame{input[0]}

		got := commitFilter.FilterFrames(ctx, input, 1)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpeted frames: %v", diff)
		}
	})

	t.Run("test three frames middle has no commits but index is behind", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []Frame{{
			toTime("2020-05-01"), toTime("2020-06-01"), "abcdef",
		}, {
			toTime("2020-06-01"), toTime("2020-07-01"), "fedcba",
		}, {
			toTime("2020-07-01"), toTime("2020-08-01"), "111222333",
		}}

		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:        1,
			Enabled:       true,
			LastIndexedAt: toTime("2020-06-02"),
		}, nil)

		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)

		want := []Frame{input[0], input[1]}

		got := commitFilter.FilterFrames(ctx, input, 1)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("unexpeted frames: %v", diff)
		}
	})
}

func toTime(date string) time.Time {
	result, _ := time.Parse("2006-01-02", date)
	return result
}
