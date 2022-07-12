package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
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

func (b *BuildEvent) HasFailed() bool {
	if b.Job.ExitStatus == nil {
		return false
	}
	if b.Job.SoftFailed || *b.Job.ExitStatus == 0 {
		return false
	}

	return true
}

type BuildStore struct {
	builds map[int]Build
	m      sync.RWMutex
}

func NewBuildStore() *BuildStore {
	return &BuildStore{
		builds: make(map[int]Build),
		m:      sync.RWMutex{},
	}
}

func (s *BuildStore) Add(event *BuildEvent) {
	s.m.Lock()
	defer s.m.Unlock()
	build, ok := s.builds[*event.Build.Number]
	if !ok {
		build = *NewBuildFrom(event)
	}
	build.Jobs = append(build.Jobs, event.Job)
	log.Printf("job %s added for build %d", *event.Job.Name, *build.Number)
}

func (s *BuildStore) DelByBuildNumber(num int) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.builds, num)
	log.Printf("build %d deleted", num)
}

func (s *BuildStore) GetByBuildNumber(num int) Build {
	s.m.RLock()
	defer s.m.RUnlock()

	return s.builds[num]
}

type StepServer struct {
	store   *BuildStore
	bkToken string
}

func NewStepServer() *StepServer {
	token := os.Getenv("BK_WEBHOOK_TOKEN")

	if token == "" {
		panic("Environment variable BK_WEBHOOK_TOKEN cannot be empty")
	}
	return &StepServer{
		store:   NewBuildStore(),
		bkToken: token,
	}
}

func processBuildkiteRequest(req *http.Request, token string) (*BuildEvent, error) {
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
	log.Printf("received event: %s", eventName)

	var event BuildEvent
	err := readBody(req, &event)
	if errors.Is(err, ErrRequestBody) {
		log.Printf("faield to read body of request")
		return nil, ErrRequestBody
	} else if errors.Is(err, ErrJSONUnmarshall) {
		log.Printf("faield to unmarshall body of request")
		return nil, ErrJSONUnmarshall
	}

	return &event, nil
}

func (s *StepServer) handleEvent(w http.ResponseWriter, req *http.Request) {
	event, err := processBuildkiteRequest(req, s.bkToken)

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

	go s.processEvent(event)
	w.WriteHeader(http.StatusOK)
}

func readBody[T any](req *http.Request, target T) error {
	log.Println("reading event detail from request")
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Println("failed to read request body")
		return ErrRequestBody
	}

	err = json.Unmarshal(data, &target)
	if err != nil {
		log.Printf("failed to unmarshall request body: %v", err)
		return ErrJSONUnmarshall
	}

	return nil
}

func (s *StepServer) shouldNotify(event *BuildEvent) bool {
	if !event.IsBuildFinished() {
		log.Printf("build %d isn't finished - not notifying", *event.Build.Number)
		return false
	}
	return true
}

func (s *StepServer) notify(build *Build) error {
	if len(build.Jobs) == 0 {
		log.Printf("build %d has no jobs", *build.Number)
		return nil
	}

	failed := make([]buildkite.Job, 0)
	for _, step := range build.Jobs {
		if *step.ExitStatus != 0 {
			failed = append(failed, step)
		}
	}

	out := ""
	for _, f := range failed {
		out = fmt.Sprintf("%s\n%s", out, *f.Name)
	}
	log.Printf("\nBuild %d failed%s", build.Number, out)
	return nil
}

func (s *StepServer) processEvent(event *BuildEvent) {
	if event.Build.Number == nil {
		//Build number is required!
		return
	}
	s.store.Add(event)
	if s.shouldNotify(event) {
		build := s.store.GetByBuildNumber(*event.Build.Number)
		s.notify(&build)
		// since we've sent a notification of the job we can remove it
		s.store.DelByBuildNumber(*event.Build.Number)
	}
}

func (s *StepServer) Serve() error {
	http.HandleFunc("/buildkite", s.handleEvent)
	log.Print("listening on :8080")
	return http.ListenAndServe(":8080", nil)
}

func main() {
	server := NewStepServer()
	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
