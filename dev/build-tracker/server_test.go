package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
)

func TestGetBuild(t *testing.T) {
	logger := logtest.Scoped(t)

	req, _ := http.NewRequest(http.MethodGet, "/-/debug/1234", nil)
	req = mux.SetURLVars(req, map[string]string{"buildNumber": "1234"})

	t.Run("401 Unauthorized when in production mode and incorrect credentials", func(t *testing.T) {
		server := NewServer(logger, config{Production: true, DebugPassword: "this is a test"})
		rec := httptest.NewRecorder()
		server.handleGetBuild(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)

		req.SetBasicAuth("devx", "this is the wrong password")
		server.handleGetBuild(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)
	})

	t.Run("404 for build that does not exist", func(t *testing.T) {
		server := NewServer(logger, config{})
		rec := httptest.NewRecorder()
		server.handleGetBuild(rec, req)

		require.Equal(t, 404, rec.Result().StatusCode)
	})

	t.Run("get marshalled json for build", func(t *testing.T) {
		server := NewServer(logger, config{})
		rec := httptest.NewRecorder()

		num := 1234
		url := "http://www.google.com"
		commit := "78926a5b3b836a8a104a5d5adf891e5626b1e405"
		pipelineID := "sourcegraph"
		msg := "this is a test"
		job := newJob("job 1", 0)

		req = mux.SetURLVars(req, map[string]string{"buildNumber": "1234"})

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

		expected := event.build()
		server.store.builds[event.buildNumber()] = expected

		server.handleGetBuild(rec, req)

		require.Equal(t, 200, rec.Result().StatusCode)

		var result Build
		err := json.NewDecoder(rec.Body).Decode(&result)
		require.NoError(t, err)
		require.Equal(t, *expected, result)
	})

	t.Run("200 with valid credentials in production mode", func(t *testing.T) {
		server := NewServer(logger, config{Production: true, DebugPassword: "this is a test"})
		rec := httptest.NewRecorder()

		req.SetBasicAuth("devx", server.config.DebugPassword)
		server.store.builds[1234] = (&Event{}).build()
		server.handleGetBuild(rec, req)

		require.Equal(t, http.StatusOK, rec.Result().StatusCode)
	})
}

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

func TestShouldNotify(t *testing.T) {
	logger := logtest.Scoped(t)

	num := 160000
	build := Build{
		Build: buildkite.Build{
			Number: &num,
		},
		Jobs: map[string]Job{},
	}
	newEvent := func(name, job, state string) *Event {
		return &Event{
			Name:     name,
			Build:    build.Build,
			Pipeline: buildkite.Pipeline{},
			Job: buildkite.Job{
				Name:  &job,
				State: &state,
			},
		}
	}
	server := NewServer(logger, config{})
	server.store.builds[*build.Number] = &build
	server.store.Add(newEvent("job.finished", "Job 1", "failed"))
	server.store.Add(newEvent("job.finished", "Job 2", "failed"))

	// build is still busy, so we shouldn't notify
	if server.shouldNotify(&build, newEvent("job.finished", "", "failed")) {
		t.Errorf("expected shouldNotify to be false, for 2 failed jobs and non-failed build")
	}

	// now fail the build
	state := "failed"
	build.State = &state
	if !server.shouldNotify(&build, newEvent("build.finished", "", "failed")) {
		t.Errorf("expected shouldNotify to be true, for 2 failed jobs and a failed build")
	}

	// Should still notify if we receive late jobs
	if !server.shouldNotify(&build, newEvent("job.finished", ":jest:", "failed")) {
		t.Errorf("expected shouldNotify to be true, for failed job after build is also finished")
	}

	// Huzzah! the build passed, but now we shouldn't send notification
	state = "passed"
	build.State = &state
	if server.shouldNotify(&build, newEvent("job.finished", ":jest:", "success")) {
		t.Errorf("expected shouldNotify to be false, for success job after build is also finished")
	}
}
