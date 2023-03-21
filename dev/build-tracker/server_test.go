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

	"github.com/sourcegraph/sourcegraph/dev/build-tracker/build"
	"github.com/sourcegraph/sourcegraph/dev/build-tracker/config"
)

func TestGetBuild(t *testing.T) {
	logger := logtest.Scoped(t)

	req, _ := http.NewRequest(http.MethodGet, "/-/debug/1234", nil)
	req = mux.SetURLVars(req, map[string]string{"buildNumber": "1234"})

	t.Run("401 Unauthorized when in production mode and incorrect credentials", func(t *testing.T) {
		server := NewServer(logger, config.Config{Production: true, DebugPassword: "this is a test"})
		rec := httptest.NewRecorder()
		server.handleGetBuild(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)

		req.SetBasicAuth("devx", "this is the wrong password")
		server.handleGetBuild(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)
	})

	t.Run("404 for build that does not exist", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		rec := httptest.NewRecorder()
		server.handleGetBuild(rec, req)

		require.Equal(t, 404, rec.Result().StatusCode)
	})

	t.Run("get marshalled json for build", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		rec := httptest.NewRecorder()

		num := 1234
		url := "http://www.google.com"
		commit := "78926a5b3b836a8a104a5d5adf891e5626b1e405"
		pipelineID := "sourcegraph"
		msg := "this is a test"
		job := newJob(t, "job 1 test", 0)
		req = mux.SetURLVars(req, map[string]string{"buildNumber": "1234"})

		event := build.Event{
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

		expected := event.WrappedBuild()
		expected.AddJob(event.WrappedJob())

		server.store.Add(&event)

		server.handleGetBuild(rec, req)

		require.Equal(t, 200, rec.Result().StatusCode)

		var result build.Build
		err := json.NewDecoder(rec.Body).Decode(&result)
		require.NoError(t, err)
		require.Equal(t, *expected, result)
	})

	t.Run("200 with valid credentials in production mode", func(t *testing.T) {
		server := NewServer(logger, config.Config{Production: true, DebugPassword: "this is a test"})
		rec := httptest.NewRecorder()

		req.SetBasicAuth("devx", server.config.DebugPassword)
		num := 1234
		server.store.Add(&build.Event{
			Name: "Fake",
			Build: buildkite.Build{
				Number: &num,
			},
			Pipeline: buildkite.Pipeline{},
			Job:      buildkite.Job{},
		})
		server.handleGetBuild(rec, req)

		require.Equal(t, http.StatusOK, rec.Result().StatusCode)
	})
}

func TestOldBuildsGetDeleted(t *testing.T) {
	logger := logtest.Scoped(t)

	finishedBuild := func(num int, state string, finishedAt time.Time) *build.Build {
		b := buildkite.Build{}
		b.State = &state
		b.Number = &num
		b.FinishedAt = &buildkite.Timestamp{Time: finishedAt}

		return &build.Build{Build: b}
	}

	t.Run("All old builds get removed", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		b := finishedBuild(1, "passed", time.Now().AddDate(-1, 0, 0))
		server.store.Set(b)

		b = finishedBuild(2, "canceled", time.Now().AddDate(0, -1, 0))
		server.store.Set(b)

		b = finishedBuild(3, "failed", time.Now().AddDate(0, 0, -1))
		server.store.Set(b)

		stopFunc := server.startCleaner(10*time.Millisecond, 24*time.Hour)
		time.Sleep(20 * time.Millisecond)
		stopFunc()

		builds := server.store.FinishedBuilds()

		if len(builds) != 0 {
			t.Errorf("Not all old builds removed. Got %d, wanted %d", len(builds), 0)
		}
	})
	t.Run("1 build left after old builds are removed", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		b := finishedBuild(1, "canceled", time.Now().AddDate(-1, 0, 0))
		server.store.Set(b)

		b = finishedBuild(2, "passed", time.Now().AddDate(0, -1, 0))
		server.store.Set(b)

		b = finishedBuild(3, "failed", time.Now())
		server.store.Set(b)

		stopFunc := server.startCleaner(10*time.Millisecond, 24*time.Hour)
		time.Sleep(20 * time.Millisecond)
		stopFunc()

		builds := server.store.FinishedBuilds()

		if len(builds) != 1 {
			t.Errorf("Expected one build to be left over. Got %d, wanted %d", len(builds), 1)
		}
	})

}
