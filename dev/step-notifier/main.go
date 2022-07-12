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

type BuildEvent struct {
	Event    string          `json:"event"`
	Build    buildkite.Build `json:"build,omitempty"`
	Job      buildkite.Job   `json:"job,omitempty"`
	Finished bool
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

func (b *BuildEvent) GetName() string {
	if b.Job.Name != nil {
		return *b.Job.Name
	}

	return ""
}

type BuildStore struct {
	builds map[int][]BuildEvent
	m      sync.RWMutex
}

func NewBuildStore() *BuildStore {
	return &BuildStore{
		builds: make(map[int][]BuildEvent),
		m:      sync.RWMutex{},
	}
}

func (s *BuildStore) addIfFailed(event *BuildEvent) {
	if !event.HasFailed() {
		log.Printf("skipping step %+v - not failed", event.GetName())
		return
	}

	s.m.Lock()
	defer s.m.Unlock()
	v, ok := s.builds[*event.Build.Number]
	if !ok {
		v = make([]BuildEvent, 0)
	}
	s.builds[*event.Build.Number] = append(v, *event)
	log.Printf("step %s added", event.GetName())
}

func (s *BuildStore) DelByBuildNumber(num int) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.builds, num)
	log.Printf("build %d deleted", num)
}

func (s *BuildStore) GetByBuildNumber(num int) []BuildEvent {
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

func processBuildkiteRequest(req *http.Request, token string) (map[string]json.RawMessage, string, error) {
	h, ok := req.Header["X-Buildkite-Token"]
	if !ok || len(h) == 0 {
		return nil, "", ErrInvalidToken
	} else if h[0] != token {
		return nil, "", ErrInvalidToken
	}

	h, ok = req.Header["X-Buildkite-Event"]
	if !ok || len(h) == 0 {
		return nil, "", ErrInvalidHeader
	}

	event := h[0]
	log.Printf("received event: %s", event)

	var payload map[string]json.RawMessage
	err := readBody(req, &payload)
	if errors.Is(err, ErrRequestBody) {
		log.Printf("faield to read body of request")
		return nil, "", ErrRequestBody
	} else if errors.Is(err, ErrJSONUnmarshall) {
		log.Printf("faield to unmarshall body of request")
		return nil, "", ErrJSONUnmarshall
	}

	return payload, event, nil
}

func (s *StepServer) handleEvent(w http.ResponseWriter, req *http.Request) {
	payload, event, err := processBuildkiteRequest(req, s.bkToken)

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

	var step BuildEvent
	// read the build number from the Build payload
	err = json.Unmarshal(payload["build"], &step)
	if err != nil {
		log.Printf("failed to unmarshall build from payload: %v", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	switch event {
	case "build.finished":
		step.Job.Name = &event
		step.Finished = true
	case "job.finished":
		err = json.Unmarshal(payload["job"], &step.Job)
		if err != nil {
			log.Printf("failed to unmarshall job from payload: %v", err)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
	default:
		w.WriteHeader(http.StatusOK)
		return

	}

	go s.processStep(&step)
	w.WriteHeader(http.StatusOK)
}

func readBody[T any](req *http.Request, target T) error {
	log.Println("reading step detail from request")
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

func (s *StepServer) shouldNotify(step *BuildEvent) bool {
	if !step.Finished {
		log.Printf("build %d isn't finished - not notifying", *step.Build.Number)
		return false
	}
	return true
}

func (s *StepServer) notify(event *BuildEvent) error {
	steps := s.store.GetByBuildNumber(*event.Build.Number)
	if len(steps) == 0 {
		log.Printf("build %d has no failed steps - not notifying\n", *event.Build.Number)
		return nil
	}

	failed := make([]BuildEvent, 0)
	for _, step := range steps {
		if *step.Job.ExitStatus != 0 {
			failed = append(failed, step)
		}
	}

	out := ""
	for _, f := range failed {
		out = fmt.Sprintf("%s\n%s", out, f.GetName())
	}
	log.Printf("\nBuild %d failed%s", *event.Build.Number, out)
	return nil
}

func (s *StepServer) processStep(step *BuildEvent) {
	s.store.addIfFailed(step)
	if s.shouldNotify(step) {
		s.notify(step)
		// since we've sent a notification of the job we can remove it
		s.store.DelByBuildNumber(*step.Build.Number)
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
