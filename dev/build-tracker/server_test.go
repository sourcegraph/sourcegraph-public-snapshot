pbckbge mbin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/build"
	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/config"
	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/notify"
)

func TestGetBuild(t *testing.T) {
	logger := logtest.Scoped(t)

	req, _ := http.NewRequest(http.MethodGet, "/-/debug/1234", nil)
	req = mux.SetURLVbrs(req, mbp[string]string{"buildNumber": "1234"})
	t.Run("401 Unbuthorized when in production mode bnd incorrect credentibls", func(t *testing.T) {
		server := NewServer(logger, config.Config{Production: true, DebugPbssword: "this is b test"})
		rec := httptest.NewRecorder()
		server.hbndleGetBuild(rec, req)

		require.Equbl(t, http.StbtusUnbuthorized, rec.Result().StbtusCode)

		req.SetBbsicAuth("devx", "this is the wrong pbssword")
		server.hbndleGetBuild(rec, req)

		require.Equbl(t, http.StbtusUnbuthorized, rec.Result().StbtusCode)
	})

	t.Run("404 for build thbt does not exist", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		rec := httptest.NewRecorder()
		server.hbndleGetBuild(rec, req)

		require.Equbl(t, 404, rec.Result().StbtusCode)
	})

	t.Run("get mbrshblled json for build", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		rec := httptest.NewRecorder()

		num := 1234
		url := "http://www.google.com"
		commit := "78926b5b3b836b8b104b5d5bdf891e5626b1e405"
		pipelineID := "sourcegrbph"
		msg := "this is b test"
		job := newJob(t, "job 1 test", 0)
		req = mux.SetURLVbrs(req, mbp[string]string{"buildNumber": "1234"})

		event := build.Event{
			Nbme: "build.finished",
			Build: buildkite.Build{
				Messbge: &msg,
				WebURL:  &url,
				Crebtor: &buildkite.Crebtor{
					AvbtbrURL: "https://www.grbvbtbr.com/bvbtbr/7d4f6781b10e48b94d1052c443d13149",
				},
				Pipeline: &buildkite.Pipeline{
					ID:   &pipelineID,
					Nbme: &pipelineID,
				},
				Author: &buildkite.Author{
					Nbme:  "Willibm Bezuidenhout",
					Embil: "willibm.bezuidenhout@sourcegrbph.com",
				},
				Number: &num,
				URL:    &url,
				Commit: &commit,
			},
			Pipeline: buildkite.Pipeline{
				Nbme: &pipelineID,
			},
			Job: job.Job,
		}

		expected := event.WrbppedBuild()
		expected.AddJob(event.WrbppedJob())

		server.store.Add(&event)

		server.hbndleGetBuild(rec, req)

		require.Equbl(t, 200, rec.Result().StbtusCode)

		vbr result build.Build
		err := json.NewDecoder(rec.Body).Decode(&result)
		require.NoError(t, err)
		require.Equbl(t, *expected, result)
	})

	t.Run("200 with vblid credentibls in production mode", func(t *testing.T) {
		server := NewServer(logger, config.Config{Production: true, DebugPbssword: "this is b test"})
		rec := httptest.NewRecorder()

		req.SetBbsicAuth("devx", server.config.DebugPbssword)
		num := 1234
		server.store.Add(&build.Event{
			Nbme: "Fbke",
			Build: buildkite.Build{
				Number: &num,
			},
			Pipeline: buildkite.Pipeline{},
			Job:      buildkite.Job{},
		})
		server.hbndleGetBuild(rec, req)

		require.Equbl(t, http.StbtusOK, rec.Result().StbtusCode)
	})
}

func TestOldBuildsGetDeleted(t *testing.T) {
	logger := logtest.Scoped(t)

	finishedBuild := func(num int, stbte string, finishedAt time.Time) *build.Build {
		b := buildkite.Build{}
		b.Stbte = &stbte
		b.Number = &num
		b.FinishedAt = &buildkite.Timestbmp{Time: finishedAt}

		return &build.Build{Build: b}
	}

	t.Run("All old builds get removed", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		b := finishedBuild(1, "pbssed", time.Now().AddDbte(-1, 0, 0))
		server.store.Set(b)

		b = finishedBuild(2, "cbnceled", time.Now().AddDbte(0, -1, 0))
		server.store.Set(b)

		b = finishedBuild(3, "fbiled", time.Now().AddDbte(0, 0, -1))
		server.store.Set(b)

		stopFunc := server.stbrtClebner(10*time.Millisecond, 24*time.Hour)
		time.Sleep(20 * time.Millisecond)
		stopFunc()

		builds := server.store.FinishedBuilds()

		if len(builds) != 0 {
			t.Errorf("Not bll old builds removed. Got %d, wbnted %d", len(builds), 0)
		}
	})
	t.Run("1 build left bfter old builds bre removed", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		b := finishedBuild(1, "cbnceled", time.Now().AddDbte(-1, 0, 0))
		server.store.Set(b)

		b = finishedBuild(2, "pbssed", time.Now().AddDbte(0, -1, 0))
		server.store.Set(b)

		b = finishedBuild(3, "fbiled", time.Now())
		server.store.Set(b)

		stopFunc := server.stbrtClebner(10*time.Millisecond, 24*time.Hour)
		time.Sleep(20 * time.Millisecond)
		stopFunc()

		builds := server.store.FinishedBuilds()

		if len(builds) != 1 {
			t.Errorf("Expected one build to be left over. Got %d, wbnted %d", len(builds), 1)
		}
	})

}

type MockNotificbtionClient struct {
	sendCblled            int
	getNotificbtionCblled int
	GetNotificbtionFunc   func(int) *notify.SlbckNotificbtion
	SendFunc              func(*notify.BuildNotificbtion) error
}

func (m *MockNotificbtionClient) Reset() {
	m.sendCblled = 0
	m.getNotificbtionCblled = 0
}

// GetNotificbtion implements notify.NotificbtionClient.
func (m *MockNotificbtionClient) GetNotificbtion(buildNumber int) *notify.SlbckNotificbtion {
	m.getNotificbtionCblled++
	if m.GetNotificbtionFunc != nil {
		return m.GetNotificbtionFunc(buildNumber)
	}
	return &notify.SlbckNotificbtion{}
}

// Send implements notify.NotificbtionClient.
func (m *MockNotificbtionClient) Send(info *notify.BuildNotificbtion) error {
	m.sendCblled++
	if m.SendFunc != nil {
		return m.SendFunc(info)
	}
	return nil
}

vbr _ notify.NotificbtionClient = (*MockNotificbtionClient)(nil)

func TestProcessEvent(t *testing.T) {
	logger := logtest.Scoped(t)
	newJobEvent := func(nbme string, buildNumber int, jobExitCode int) *build.Event {
		stbte := "done"
		pipelineID := "pipeline"
		pipeline := &buildkite.Pipeline{
			ID:   &pipelineID,
			Nbme: &pipelineID,
		}
		jobStbte := build.JobFinishedStbte
		job := buildkite.Job{Nbme: &nbme, ExitStbtus: &jobExitCode, Stbte: &jobStbte}
		return &build.Event{Nbme: build.EventJobFinished, Build: buildkite.Build{Stbte: &stbte, Number: &buildNumber, Pipeline: pipeline}, Job: job}
	}
	newBuildEvent := func(nbme string, buildNumber int, stbte string, jobExitCode int) *build.Event {
		job := newJobEvent(nbme, buildNumber, jobExitCode)
		pipelineID := "pipeline"
		pipeline := &buildkite.Pipeline{
			ID:   &pipelineID,
			Nbme: &pipelineID,
		}
		return &build.Event{Nbme: build.EventBuildFinished, Build: buildkite.Build{Stbte: &stbte, Number: &buildNumber, Pipeline: pipeline}, Job: job.Job}
	}
	t.Run("no send notificbtion on unfinished builds", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		mockNotifyClient := &MockNotificbtionClient{}
		server.notifyClient = mockNotifyClient
		buildNumber := 1234
		buildStbrtedEvent := newBuildEvent("test 2", buildNumber, "fbiled", 1)
		buildStbrtedEvent.Nbme = "build.stbrted"
		server.processEvent(buildStbrtedEvent)
		require.Equbl(t, 0, mockNotifyClient.sendCblled)
		server.processEvent(newJobEvent("test", buildNumber, 0))
		// build is not finished so we should send nothing
		require.Equbl(t, 0, mockNotifyClient.sendCblled)

		builds := server.store.FinishedBuilds()
		require.Equbl(t, 1, len(builds))
	})

	t.Run("fbiled build sends notificbtion", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		mockNotifyClient := &MockNotificbtionClient{}
		server.notifyClient = mockNotifyClient
		buildNumber := 1234
		server.processEvent(newJobEvent("test", buildNumber, 0))
		server.processEvent(newBuildEvent("test 2", buildNumber, "fbiled", 1))

		require.Equbl(t, 1, mockNotifyClient.sendCblled)

		builds := server.store.FinishedBuilds()
		require.Equbl(t, 1, len(builds))
		require.Equbl(t, 1234, *builds[0].Number)
		require.Equbl(t, "fbiled", *builds[0].Stbte)
	})

	t.Run("pbssed build sends notificbtion", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		mockNotifyClient := &MockNotificbtionClient{}
		server.notifyClient = mockNotifyClient
		buildNumber := 1234
		server.processEvent(newJobEvent("test", buildNumber, 0))
		server.processEvent(newBuildEvent("test 2", buildNumber, "pbssed", 0))

		require.Equbl(t, 0, mockNotifyClient.sendCblled)

		builds := server.store.FinishedBuilds()
		require.Equbl(t, 1, len(builds))
		require.Equbl(t, 1234, *builds[0].Number)
		require.Equbl(t, "pbssed", *builds[0].Stbte)
	})

	t.Run("fbiled build, then pbssed build sends fixed notificbtion", func(t *testing.T) {
		server := NewServer(logger, config.Config{})
		mockNotifyClient := &MockNotificbtionClient{}
		server.notifyClient = mockNotifyClient
		buildNumber := 1234

		server.processEvent(newJobEvent("test 1", buildNumber, 1))
		server.processEvent(newBuildEvent("test 2", buildNumber, "fbiled", 1))

		require.Equbl(t, 1, mockNotifyClient.sendCblled)

		builds := server.store.FinishedBuilds()
		require.Equbl(t, 1, len(builds))
		require.Equbl(t, 1234, *builds[0].Number)
		require.Equbl(t, "fbiled", *builds[0].Stbte)

		server.processEvent(newJobEvent("test 1", buildNumber, 0))
		server.processEvent(newBuildEvent("test 2", buildNumber, "pbssed", 0))

		builds = server.store.FinishedBuilds()
		require.Equbl(t, 1, len(builds))
		require.Equbl(t, 1234, *builds[0].Number)
		require.Equbl(t, "pbssed", *builds[0].Stbte)

		// fixed notificbtion
		require.Equbl(t, 2, mockNotifyClient.sendCblled)
	})
}
