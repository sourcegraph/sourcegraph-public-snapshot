package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/gorilla/mux"
	"github.com/sourcegraph/log/logtest"
)

func TestGetBuild(t *testing.T) {
	logger := logtest.Scoped(t)
	server := NewServer(logger, config{})

	req, _ := http.NewRequest(http.MethodGet, "/-/debug/1234", nil)
	req = mux.SetURLVars(req, map[string]string{"buildNumber": "1234"})

	t.Run("404 for build that does not exist", func(t *testing.T) {
		rec := httptest.NewRecorder()
		server.handleGetBuild(rec, req)

		if rec.Result().StatusCode != 404 {
			t.Errorf("expected 404 status code for build that does not exist")
		}
	})

	t.Run("200 for build that does exist", func(t *testing.T) {
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

		server.store.builds[event.buildNumber()] = event.build()

		t.Logf("+%v\n", server.store.builds)
		t.Logf("+%v\n", server.store.builds[1234])

		server.handleGetBuild(rec, req)

		if rec.Result().StatusCode != 200 {
			t.Errorf("expected status 200 but got %d", rec.Result().StatusCode)
		}
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
