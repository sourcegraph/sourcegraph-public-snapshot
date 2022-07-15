package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sourcegraph/log"
	"io/ioutil"
	_ "log"
	"net/http"
	"os"
	"sync"

	"github.com/buildkite/go-buildkite/v3/buildkite"
)

var ErrRequestBody = fmt.Errorf("failed to read body from request")
var ErrJSONUnmarshall = fmt.Errorf("failed to unmarshal body")
var ErrInvalidToken = fmt.Errorf("buildkite token is invalid")
var ErrInvalidHeader = fmt.Errorf("Header of request is invalid")
var ErrUnwantedEvent = fmt.Errorf("Unwanted event received")

type Build struct {
	buildkite.Build
	Jobs []buildkite.Job
}

func (b *Build) HasFailed() bool {
	for _, j := range b.Jobs {
		if j.ExitStatus != nil && !j.SoftFailed && *j.ExitStatus > 0 {
			return true
		}
	}
	return false
}

func (b *Build) PipelineName() string {
	fmt.Printf("%+v\n\n%+v\n\n", b.Pipeline, b)
	if b.Pipeline == nil {
		return "N/A"
	}
	if b.Pipeline.Name == nil {
		return "N/A"
	}
	return *b.Pipeline.Name

}

func NewBuildFrom(event *BuildEvent) *Build {
	return &Build{
		Build: event.Build,
		Jobs:  make([]buildkite.Job, 0),
	}
}

type BuildEvent struct {
	Event string          `json:"event"`
	Build buildkite.Build `json:"build,omitempty"`
	Job   buildkite.Job   `json:"job,omitempty"`
}

func (b *BuildEvent) IsBuildFinished() bool {
	return b.Event == "build.finished"
}

func (b *BuildEvent) BuildNumber() int {
	if b.Build.Number == nil {
		return -1
	}
	return *b.Build.Number
}

func (b *BuildEvent) JobName() string {
	if b.Job.Name == nil {
		return "N/A"
	}
	return *b.Job.Name
}

type BuildStore struct {
	logger log.Logger
	builds map[int]*Build
	m      sync.RWMutex
}

func NewBuildStore(logger log.Logger) *BuildStore {
	return &BuildStore{
		logger: logger.Scoped("store", "stores all the builds"),
		builds: make(map[int]*Build),
		m:      sync.RWMutex{},
	}
}

func (s *BuildStore) Add(event *BuildEvent) {
	s.m.Lock()
	defer s.m.Unlock()
	build, ok := s.builds[event.BuildNumber()]
	if !ok {
		build = NewBuildFrom(event)
		s.builds[event.BuildNumber()] = build
	}
	// if the build is finished replace the original build with the replaced one since it will be more up to date
	if event.IsBuildFinished() {
		build.Build = event.Build
	}
	build.Jobs = append(build.Jobs, event.Job)

	s.logger.Debug("job added", log.Int("buildNumber", event.BuildNumber()), log.Int("totalJobs", len(build.Jobs)))
}

func (s *BuildStore) DelByBuildNumber(num int) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.builds, num)
	s.logger.Info("build deleted", log.Int("buildNumber", num))
}

func (s *BuildStore) GetByBuildNumber(num int) *Build {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.builds[num]
}

type BuildTrackingServer struct {
	logger  log.Logger
	store   *BuildStore
	bkToken string
	slack   *SlackClient
}

func NewStepServer() (*BuildTrackingServer, error) {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return nil, fmt.Errorf("SLACK_TOKEN cannot be empty")
	}
	token := os.Getenv("BK_WEBHOOK_TOKEN")

	if token == "" {
		return nil, fmt.Errorf("BK_WEBHOOK_TOKEN cannot be empty")
	}
	logger := log.Scoped("server", "Server that tracks completed builds")
	return &BuildTrackingServer{
		logger:  logger,
		store:   NewBuildStore(logger),
		bkToken: token,
		slack:   NewSlackClient(logger, slackToken),
	}, nil
}

func (s *BuildTrackingServer) processBuildkiteRequest(req *http.Request, token string) (*BuildEvent, error) {
	h, ok := req.Header["X-Buildkite-Token"]
	if !ok || len(h) == 0 {
		return nil, ErrInvalidToken
	} else if h[0] != token {
		return nil, ErrInvalidToken
	}

	h, ok = req.Header["X-Buildkite-Event"]
	if !ok || len(h) == 0 {
		return nil, ErrInvalidHeader
	}

	eventName := h[0]
	s.logger.Debug("receied event", log.String("eventName", eventName))

	var event BuildEvent
	err := readBody(s.logger, req, &event)
	if errors.Is(err, ErrRequestBody) {
		s.logger.Error("failed to read body of request", log.Error(err))
		return nil, ErrRequestBody
	} else if errors.Is(err, ErrJSONUnmarshall) {
		s.logger.Error("failed to unmarshall body of request", log.Error(err))
		return nil, ErrJSONUnmarshall
	}

	return &event, nil
}

func (s *BuildTrackingServer) handleEvent(w http.ResponseWriter, req *http.Request) {
	event, err := s.processBuildkiteRequest(req, s.bkToken)

	switch err {
	case ErrInvalidToken:
	case ErrInvalidHeader:
		w.WriteHeader(http.StatusBadRequest)
		return
	case ErrUnwantedEvent:
		w.WriteHeader(http.StatusOK)
		return
	case ErrRequestBody:
		w.WriteHeader(http.StatusBadRequest)
		return
	case ErrJSONUnmarshall:
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	s.logger.Info("processing event", log.String("eventName", event.Event), log.Int("buildNumber", event.BuildNumber()), log.String("JobName", event.JobName()))
	go s.processEvent(event)
	w.WriteHeader(http.StatusOK)
}

func readBody[T any](logger log.Logger, req *http.Request, target T) error {
	data, err := ioutil.ReadAll(req.Body)
	logger.Debug("read body of request", log.String("data", string(data)))
	if err != nil {
		logger.Error("failed to read request body", log.Error(err))
		return ErrRequestBody
	}

	err = json.Unmarshal(data, &target)
	if err != nil {
		logger.Error("failed to unmarshall request body", log.Error(err))
		return ErrJSONUnmarshall
	}

	return nil
}

func (s *BuildTrackingServer) notify(build *Build) error {
	if len(build.Jobs) == 0 {
		s.logger.Info("build has no jobs", log.Int("buildNumber", *build.Number))
		return nil
	}

	if build.HasFailed() {
		s.logger.Info("detected failed build - sending notification", log.Int("buildNumber", *build.Number))
		return s.slack.sendNotification(build)
	}

	s.logger.Info("build successful", log.Int("buildNumber", *build.Number))
	return nil
}

func (s *BuildTrackingServer) processEvent(event *BuildEvent) {
	if event.Build.Number == nil {
		//Build number is required!
		return
	}

	s.store.Add(event)
	if event.IsBuildFinished() {
		build := s.store.GetByBuildNumber(event.BuildNumber())
		if err := s.notify(build); err != nil {
			s.logger.Error("failed to send notification for build", log.Int("buildNumber", event.BuildNumber()), log.Error(err))
		}
		// since the build is done we don't need it anymore
		s.store.DelByBuildNumber(*event.Build.Number)
	}
}

func (s *BuildTrackingServer) Serve() error {
	http.HandleFunc("/buildkite", s.handleEvent)
	s.logger.Info("listening on :8080")
	return http.ListenAndServe(":8080", nil)
}

func main() {
	sync := log.Init(log.Resource{
		Name:      "BuildTracker",
		Namespace: "CI",
	})
	defer sync.Sync()
	server, err := NewStepServer()
	if err != nil {
		panic(err)
	}
	if err := server.Serve(); err != nil {
		server.logger.Fatal("server exited with error", log.Error(err))
	}
}
