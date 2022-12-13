package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrInvalidToken = errors.New("buildkite token is invalid")
var ErrInvalidHeader = errors.New("Header of request is invalid")
var ErrUnwantedEvent = errors.New("Unwanted event received")

var nowFunc func() time.Time = time.Now

const DefaultChannel = "#william-buildchecker-webhook-test"

// Server is the http server that listens for events from Buildkite. The server tracks builds and their associated jobs
// with the use of a BuildStore. Once a build is finished and has failed, the server sends a notification.
type Server struct {
	logger       log.Logger
	store        *BuildStore
	config       *config
	notifyClient *NotificationClient
	github       *github.Client
	http         *http.Server
}

type config struct {
	BuildkiteToken string
	SlackToken     string
	GithubToken    string
	SlackChannel   string
	Production     bool
	DebugPassword  string
}

func configFromEnv() (*config, error) {
	var c config

	err := envVar("BUILDKITE_WEBHOOK_TOKEN", &c.BuildkiteToken)
	if err != nil {
		return nil, err
	}
	err = envVar("SLACK_TOKEN", &c.SlackToken)
	if err != nil {
		return nil, err
	}
	err = envVar("GITHUB_TOKEN", &c.GithubToken)
	if err != nil {
		return nil, err
	}

	err = envVar("SLACK_CHANNEL", &c.SlackChannel)
	if err != nil {
		c.SlackChannel = DefaultChannel
	}

	err = envVar("BUILDTRACKER_PRODUCTION", &c.Production)
	if err != nil {
		c.Production = false
	}

	if c.Production {
		_ = envVar("BUILDTRACKER_DEBUG_PASSWORD", &c.DebugPassword)
		if c.DebugPassword == "" {
			return nil, errors.New("BUILDTRACKER_DEBUG_PASSWORD is required when BUILDTRACKER_PRODUCTION is true")
		}
	}

	return &c, nil
}

func (s *Server) handleRevert(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	commit, ok := vars["commit"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(commit) < 8 {
		s.logger.Warn("received commith with wrong length", log.Int("length", len(commit)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.logger.Info("received request to revert commit", log.String("commit", commit))

	// We want return a response and continue async
	ctx := context.Background()
	bernash := fmt.Sprintf("build-tracker/reverts/%s", commit[:8])
	bernashPath := strings.ReplaceAll(bernash, "/", "-")

	cmd := run.Bash(ctx, fmt.Sprintf("git worktree add -b %s %s", bernash, bernashPath)).Dir("git-reverts")
	if err := cmd.Run().Wait(); err != nil {
		s.logger.Error("failed to add worktree", log.String("path", bernashPath), log.String("branch", bernash), log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Clean up when shit goes wrong
	defer func() {
		cmd := run.Bash(ctx, fmt.Sprintf("git worktree remove %s --force", bernash)).Dir("git-reverts")
		if err := cmd.Run().Wait(); err != nil {
			s.logger.Error("failed to remove worktree", log.String("path", bernashPath), log.String("branch", bernash), log.Error(err))
		}

		// Ensure the branch is gone!
		cmd = run.Bash(ctx, fmt.Sprintf("git branch -D %s --force", bernash)).Dir("git-reverts")
		if err := cmd.Run().Wait(); err != nil {
			s.logger.Error("failed to remove worktree", log.String("path", bernashPath), log.String("branch", bernash), log.Error(err))
		}
	}()

	cwd := filepath.Join("git-reverts", bernashPath)
	cmd = run.Bash(ctx, fmt.Sprintf("git revert %s --no-edit", commit)).Dir(cwd)
	if out, err := cmd.Run().String(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		fmt.Println(out)
	}

	cmd = run.Bash(ctx, fmt.Sprintf("git push origin %s", bernash)).Dir(cwd)
	if out, err := cmd.Run().String(); err != nil {
		s.logger.Error("encountered an error while trying to push revert branch", log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		fmt.Println(out)
	}

	// Create Github PR
	base := "sourcegraph"
	body := `Test revert`
	title := fmt.Sprintf("Revert of commit %s initiated via Build Tracker", commit)
	pr := github.NewPullRequest{
		Title: &title,
		Head:  &bernash,
		Base:  &base,
		Body:  &body,
	}

	gpr, resp, err := s.github.PullRequests.Create(ctx, "sourcegraph", "sourcegraph", &pr)
	if err != nil {
		s.logger.Error("failed to create pr on github", log.String("branch", bernash))
	}

	if resp.StatusCode > 299 {
		data, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		s.logger.Warn("non 200 response received from github when creating revert pr", log.String("body", string(data)), log.Int("statusCode", resp.StatusCode))
	}

	fmt.Println(gpr.GetURL())
	// send a notifcation of PR

	buildNumber, ok := vars["buildNumber"]
	number, err := strconv.Atoi(buildNumber)
	if err != nil {
		s.logger.Error("failed to parse build number", log.String("buildNumber", buildNumber))
	}

	build := s.store.GetByBuildNumber(number)
	digest := "i+love+a+digestive"
	rawApproveUrl := fmt.Sprintf("/revert/%s/%s/approve", *build.Commit, digest)
	approveUrl, err := url.Parse(rawApproveUrl)
	if err != nil {
		s.logger.Error("failed to parse approve URL", log.String("approveUrl", rawApproveUrl))
	}

	rawRejectUrl := fmt.Sprintf("/revert/%s/%s/reject", *build.Commit, digest)
	rejectUrl, err := url.Parse(rawRejectUrl)
	if err != nil {
		s.logger.Error("failed to parse reject URL", log.String("rejectUrl", rawRejectUrl))
	}

	prUrl, err := url.Parse(gpr.GetURL())
	if err != nil {
		s.logger.Error("failed to parse pr URL", log.String("prtUrl", gpr.GetURL()))
	}

	revert := Revert{
		build:      build,
		approveUrl: approveUrl,
		rejectUrl:  rejectUrl,
		prUrl:      prUrl,
	}

	s.notifyClient.sendRevert(&revert)

	w.WriteHeader(http.StatusOK)
}

func newGitHubClient(logger log.Logger, token string) (*github.Client, error) {
	ctx := context.Background()
	println("🚨" + token)
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	))

	baseURL, err := url.Parse("https://github.com")
	if err != nil {
		return nil, err
	}
	baseURL.Path = "/api/v3"

	return github.NewClient(tc), nil
}

// NewServer creatse a new server to listen for Buildkite webhook events.
func NewServer(logger log.Logger, c config) (*Server, error) {
	logger = logger.Scoped("server", "Server which tracks events received from Buildkite and sends notifications on failures")
	github, err := newGitHubClient(logger, c.GithubToken)
	if err != nil {
		logger.Error("failed to initializing GitHub client", log.Error(err))
		return nil, err
	}
	server := &Server{
		logger:       logger,
		store:        NewBuildStore(logger),
		config:       &c,
		notifyClient: NewNotificationClient(logger, c.SlackToken, c.GithubToken, c.SlackChannel),
		github:       github,
	}

	// Register routes the the server will be responding too
	r := mux.NewRouter()
	r.Path("/buildkite").HandlerFunc(server.handleEvent).Methods(http.MethodPost)
	r.Path("/healthz").HandlerFunc(server.handleHealthz).Methods(http.MethodGet)
	r.Path("/revert/{commit}").HandlerFunc(server.handleRevert).Methods(http.MethodPost)

	debug := r.PathPrefix("/-/debug").Subrouter()
	debug.Path("/{buildNumber}").HandlerFunc(server.handleGetBuild).Methods(http.MethodGet)

	server.http = &http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	return server, nil
}

func (s *Server) handleGetBuild(w http.ResponseWriter, req *http.Request) {
	if s.config.Production {
		user, pass, ok := req.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if user != "devx" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if pass != s.config.DebugPassword {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

	}
	vars := mux.Vars(req)

	buildNumParam, ok := vars["buildNumber"]
	if !ok {
		s.logger.Error("request received with no buildNumber path parameter", log.String("route", req.URL.Path))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	buildNum, err := strconv.Atoi(buildNumParam)
	if err != nil {
		s.logger.Error("invalid build number parameter received", log.String("buildNumParam", buildNumParam))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	s.logger.Info("retrieving build", log.Int("buildNumber", buildNum))
	build := s.store.GetByBuildNumber(buildNum)
	if build == nil {
		s.logger.Debug("no build found", log.Int("buildNumber", buildNum))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	s.logger.Debug("encoding build", log.Int("buildNumber", buildNum))
	json.NewEncoder(w).Encode(build)
	w.WriteHeader(http.StatusOK)
}

// handleEvent handles an event received from the http listener. A event is valid when:
// - Has the correct headers from Buildkite
// - On of the following events
//   - job.finished
//   - build.finished
//
// - Has valid JSON
// Note that if we received an unwanted event ie. the event is not "job.finished" or "build.finished" we respond with a 200 OK regardless.
// Once all the conditions are met, the event is processed in a go routine with `processEvent`
func (s *Server) handleEvent(w http.ResponseWriter, req *http.Request) {
	h, ok := req.Header["X-Buildkite-Token"]
	if !ok || len(h) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if h[0] != s.config.BuildkiteToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h, ok = req.Header["X-Buildkite-Event"]
	if !ok || len(h) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventName := h[0]
	s.logger.Debug("received event", log.String("eventName", eventName))

	data, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		s.logger.Error("failed to read request body", log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var event Event
	err = json.Unmarshal(data, &event)
	if err != nil {
		s.logger.Error("failed to unmarshall request body", log.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go s.processEvent(&event)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHealthz(w http.ResponseWriter, req *http.Request) {
	// do our super exhaustive check
	w.WriteHeader(http.StatusOK)
}

// notifyIfFailed sends a notification over slack if the provided build has failed. If the build is successful not notifcation is sent
func (s *Server) notifyIfFailed(build *Build) error {
	if build.hasFailed() {
		s.logger.Info("detected failed build - sending notification", log.Int("buildNumber", intp(build.Number)))
		return s.notifyClient.sendFailedBuild(build)
	}

	s.logger.Info("build has not failed", log.Int("buildNumber", intp(build.Number)))
	return nil
}

func (s *Server) startOldBuildCleaner(every, window time.Duration) func() {
	ticker := time.NewTicker(every)
	done := make(chan interface{})

	// We could technically remove  the builds immediately after we've sent a notification for or it, or the build has passed.
	// But we keep builds a little longer and prediodically clean them out so that we can in future allow possibly querying
	// of builds and other use cases, like retrying a build etc.
	go func() {
		for {
			select {
			case <-ticker.C:
				oldBuilds := make([]int, 0)
				now := nowFunc()
				for _, b := range s.store.FinishedBuilds() {
					finishedAt := *b.FinishedAt
					delta := now.Sub(finishedAt.Time)
					if delta >= window {
						s.logger.Debug("build past age window", log.Int("buildNumber", *b.Number), log.Time("FinishedAt", finishedAt.Time), log.Duration("window", window))
						oldBuilds = append(oldBuilds, *b.Number)
					}
				}
				s.logger.Info("deleting old builds", log.Int("oldBuildCount", len(oldBuilds)))
				s.store.DelByBuildNumber(oldBuilds...)
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() { done <- nil }
}

// processEvent processes a BuildEvent received from Buildkite. If the event is for a `build.finished` event we get the
// full build which includes all recorded jobs for the build and send a notification.
// processEvent delegates the decision to actually send a notifcation
func (s *Server) processEvent(event *Event) {
	s.logger.Info("processing event", log.String("eventName", event.Name), log.Int("buildNumber", event.buildNumber()), log.String("jobName", event.jobName()))
	s.store.Add(event)
	if event.isBuildFinished() {
		build := s.store.GetByBuildNumber(event.buildNumber())
		if err := s.notifyIfFailed(build); err != nil {
			s.logger.Error("failed to send notification for build", log.Int("buildNumber", event.buildNumber()), log.Error(err))
		}
	}
}

// Serve starts the http server and listens for buildkite build events to be sent on the route "/buildkite"
func main() {
	sync := log.Init(log.Resource{
		Name:      "BuildTracker",
		Namespace: "CI",
	})
	defer sync.Sync()

	logger := log.Scoped("BuildTracker", "main entrypoint for Build Tracking Server")

	serverConf, err := configFromEnv()
	if err != nil {
		logger.Fatal("failed to get config from env", log.Error(err))
	}
	println("🤥 " + serverConf.GithubToken)
	logger.Info("config loaded from environment", log.Object("config", log.String("SlackChannel", serverConf.SlackChannel), log.Bool("Production", serverConf.Production)))
	server, err := NewServer(logger, *serverConf)
	if err != nil {
		logger.Error("failed to initializing server", log.Error(err))
	}

	stopFn := server.startOldBuildCleaner(5*time.Minute, 24*time.Hour)
	defer stopFn()

	if server.config.Production {
		server.logger.Info("server is in production mode!")
	} else {
		server.logger.Info("server is in development mode!")
	}

	go func() {
		logger.Info("initializing sourcegraph repo for reverts")
		ctx := context.Background()
		if stat, err := os.Stat("git-reverts"); err == nil && stat.IsDir() {
			logger.Debug("worktree already exists", log.String("path", "git-reverts"))
		} else {
			logger.Debug("checking out new worktree", log.String("path", "git-reverts"))
			cmd := run.Bash(ctx, "git clone --bare github.com:sourcegraph/sourcegraph git-reverts")
			if err := cmd.Run().Wait(); err != nil {
				logger.Error("problem occured creating worktree", log.Error(err))
			}
		}
	}()

	if err := server.http.ListenAndServe(); err != nil {
		logger.Fatal("server exited with error", log.Error(err))
	}
}
