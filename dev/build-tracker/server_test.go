package main

import (
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegraph/log/logtest"
)

func TestOldBuildsGetDeleted(t *testing.T) {
	logger := logtest.Scoped(t)

	finishedBuild := func(num int, state string, finishedAt time.Time) *Build {
		b := buildkite.Build{}
		b.State = &state
		b.Number = &num
		b.FinishedAt = &buildkite.Timestamp{Time: finishedAt}

		return &Build{Build: b}
	}

	t.Run("All old builds get removed", func(t *testing.T) {
		server := NewServer(logger, config{})
		b := finishedBuild(1, "passed", time.Now().AddDate(-1, 0, 0))
		server.store.builds[*b.Number] = b

		b = finishedBuild(2, "canceled", time.Now().AddDate(0, -1, 0))
		server.store.builds[*b.Number] = b

		b = finishedBuild(3, "failed", time.Now().AddDate(0, 0, -1))
		server.store.builds[*b.Number] = b
		builds := server.store.FinishedBuilds()

		stopFunc := server.startOldBuildCleaner(10*time.Millisecond, 24*time.Hour)
		time.Sleep(20 * time.Millisecond)
		stopFunc()

		builds = server.store.FinishedBuilds()

		if len(builds) != 0 {
			t.Errorf("Not all old builds removed. Got %d, wanted %d", len(builds), 0)
		}
	})
	t.Run("1 build left after old builds are removed", func(t *testing.T) {
		server := NewServer(logger, config{})
		b := finishedBuild(1, "canceled", time.Now().AddDate(-1, 0, 0))
		server.store.builds[*b.Number] = b

		b = finishedBuild(2, "passed", time.Now().AddDate(0, -1, 0))
		server.store.builds[*b.Number] = b

		b = finishedBuild(3, "failed", time.Now())
		server.store.builds[*b.Number] = b

		stopFunc := server.startOldBuildCleaner(10*time.Millisecond, 24*time.Hour)
		time.Sleep(20 * time.Millisecond)
		stopFunc()

		builds := server.store.FinishedBuilds()

		if len(builds) != 1 {
			t.Errorf("Expected one build to be left over. Got %d, wanted %d", len(builds), 1)
		}
	})

}
