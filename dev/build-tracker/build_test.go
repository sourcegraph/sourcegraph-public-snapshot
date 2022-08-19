package main

import (
	"testing"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
)

func TestBuildStoreAdd(t *testing.T) {
	failed := "failed"
	pipeline := "bobheadxi"
	eventFailed := func(n int) *Event {
		return &Event{Name: "build.finished", Build: buildkite.Build{State: &failed, Number: &n, Pipeline: &buildkite.Pipeline{Name: &pipeline}}}
	}
	eventSucceeded := func(n int) *Event {
		// no state === not failed
		return &Event{Name: "build.finished", Build: buildkite.Build{State: nil, Number: &n, Pipeline: &buildkite.Pipeline{Name: &pipeline}}}
	}

	store := NewBuildStore(logtest.Scoped(t))

	t.Run("subsequent failures should increment ConsecutiveFailure", func(t *testing.T) {
		store.Add(eventFailed(1))
		build := store.GetByBuildNumber(1)
		assert.Equal(t, build.ConsecutiveFailure, 1)

		store.Add(eventFailed(2))
		build = store.GetByBuildNumber(2)
		assert.Equal(t, build.ConsecutiveFailure, 2)

		store.Add(eventFailed(3))
		build = store.GetByBuildNumber(3)
		assert.Equal(t, build.ConsecutiveFailure, 3)
	})

	t.Run("a pass should reset ConsecutiveFailure", func(t *testing.T) {
		store.Add(eventSucceeded(4))
		build := store.GetByBuildNumber(4)
		assert.Equal(t, build.ConsecutiveFailure, 0)

		store.Add(eventSucceeded(5))
		build = store.GetByBuildNumber(5)
		assert.Equal(t, build.ConsecutiveFailure, 0)
	})
}
