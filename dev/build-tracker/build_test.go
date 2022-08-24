package main

import (
	"testing"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFromEvent(t *testing.T) {
	num := 1234
	url := "http://www.google.com"
	commit := "78926a5b3b836a8a104a5d5adf891e5626b1e405"
	pipelineID := "sourcegraph"
	msg := "this is a test"
	job := newJob("job 1", 0)

	event := Event{
		Name: "build.finished",
		Build: buildkite.Build{
			Message: &msg,
			WebURL:  &url,
			Creator: &buildkite.Creator{
				AvatarURL: "https://www.gravatar.com/avatar/7d4f6781b10e48a94d1052c443d13149",
			},
			Pipeline: &buildkite.Pipeline{
				ID:   &pipelineID,
				Name: &pipelineID,
			},
			Author: &buildkite.Author{
				Name:  "William Bezuidenhout",
				Email: "william.bezuidenhout@sourcegraph.com",
			},
			Number: &num,
			URL:    &url,
			Commit: &commit,
		},
		Pipeline: buildkite.Pipeline{
			Name: &pipelineID,
		},
		Job: job.Job,
	}

	t.Run("original author get preserved over nil", func(t *testing.T) {
		build := event.build()
		otherEvent := event
		otherEvent.Build.Author = nil

		build.updateFromEvent(&otherEvent)

		require.NotNil(t, build.Author)
	})

	t.Run("original author get preserved if new author is empty", func(t *testing.T) {
		build := event.build()
		otherEvent := event
		otherEvent.Build.Author = &buildkite.Author{}

		build.updateFromEvent(&otherEvent)

		require.NotNil(t, build.Author)
		require.Equal(t, build.Author, event.Build.Author)
	})

	t.Run("build gets updated with new build", func(t *testing.T) {
		build := event.build()
		otherEvent := event
		num := 7777
		otherEvent.Build.Number = &num

		build.updateFromEvent(&otherEvent)

		require.Equal(t, *build.Build.Number, num)
		require.NotEqual(t, *event.Build.Number, *build.Build.Number)
	})

	t.Run("build gets updated with new pipeline", func(t *testing.T) {
		build := event.build()
		otherEvent := event
		name := "the other one"
		otherEvent.Pipeline.Name = &name

		build.updateFromEvent(&otherEvent)

		require.Equal(t, *build.Pipeline.Name, name)
		require.NotEqual(t, *event.Pipeline.Name, *build.Pipeline.Name)
	})
}

func TestBuildStoreAdd(t *testing.T) {
	failed := "failed"
	pipeline := "bobheadxi"
	eventFailed := func(n int) *Event {
		return &Event{Name: "build.finished", Build: buildkite.Build{State: &failed, Number: &n}, Pipeline: buildkite.Pipeline{Name: &pipeline}}
	}
	eventSucceeded := func(n int) *Event {
		// no state === not failed
		return &Event{Name: "build.finished", Build: buildkite.Build{State: nil, Number: &n}, Pipeline: buildkite.Pipeline{Name: &pipeline}}
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
		store.Add(eventFailed(4))
		build := store.GetByBuildNumber(4)
		assert.Equal(t, build.ConsecutiveFailure, 4)

		store.Add(eventSucceeded(5))
		build = store.GetByBuildNumber(5)
		assert.Equal(t, build.ConsecutiveFailure, 0)

		store.Add(eventFailed(6))
		build = store.GetByBuildNumber(6)
		assert.Equal(t, build.ConsecutiveFailure, 1)

		store.Add(eventSucceeded(7))
		build = store.GetByBuildNumber(7)
		assert.Equal(t, build.ConsecutiveFailure, 0)
	})
}
