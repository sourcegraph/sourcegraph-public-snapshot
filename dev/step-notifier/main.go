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

type BuildStep struct {
	Build int `json:"number"`
	buildkite.Job
	Finished bool
}

func (b *BuildStep) HasFailed() bool {
	if b.ExitStatus == nil {
		return false
	}
	if b.SoftFailed || *b.ExitStatus == 0 {
		return false
	}

	return true
}

func (b *BuildStep) GetName() string {
	if b.Job.Name != nil {
		return *b.Job.Name
	}

	return ""
}

type BuildStore struct {
	builds map[int][]BuildStep
	m      sync.RWMutex
}

func NewBuildStore() *BuildStore {
	return &BuildStore{
		builds: make(map[int][]BuildStep),
		m:      sync.RWMutex{},
	}
}

func (s *BuildStore) addIfFailed(step *BuildStep) {
	if !step.HasFailed() {
		log.Printf("skipping step %+v - not failed", step.GetName())
		return
	}

	s.m.Lock()
	defer s.m.Unlock()
	v, ok := s.builds[step.Build]
	if !ok {
		v = make([]BuildStep, 0)
	}
	s.builds[step.Build] = append(v, *step)
	log.Printf("step %s added", step.GetName())
}

func (s *BuildStore) DelByBuildNumber(num int) {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.builds, num)
	log.Printf("build %d deleted", num)
}

func (s *BuildStore) GetByBuildNumber(num int) []BuildStep {
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

	var step BuildStep
	// read the build number from the Build payload
	err = json.Unmarshal(payload["build"], &step)
	if err != nil {
		log.Printf("failed to unmarshall build from payload: %v", err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	switch event {
	case "build.finished":
		step.Name = &event
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

func (s *StepServer) shouldNotify(step *BuildStep) bool {
	if !step.Finished {
		log.Printf("build %d isn't finished - not notifying", step.Build)
		return false
	}
	return true
}

func (s *StepServer) notify(step *BuildStep) error {
	steps := s.store.GetByBuildNumber(step.Build)
	if len(steps) == 0 {
		log.Printf("build %d has no failed steps - not notifying\n", step.Build)
		return nil
	}

	failed := make([]BuildStep, 0)
	for _, step := range steps {
		if *step.Job.ExitStatus != 0 {
			failed = append(failed, step)
		}
	}

	out := ""
	for _, f := range failed {
		out = fmt.Sprintf("%s\n%s", out, f.GetName())
	}
	log.Printf("\nBuild %d failed%s", step.Build, out)
	return nil
}

func (s *StepServer) processStep(step *BuildStep) {
	s.store.addIfFailed(step)
	if s.shouldNotify(step) {
		s.notify(step)
		// since we've sent a notification of the job we can remove it
		s.store.DelByBuildNumber(step.Build)
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
