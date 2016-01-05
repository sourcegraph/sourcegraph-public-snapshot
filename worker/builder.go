package worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
	"sync"
	"time"

	droneexec "github.com/drone/drone-exec/exec"
	droneparser "github.com/drone/drone-exec/parser"
	dronerunner "github.com/drone/drone-exec/runner"
	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone/yaml/matrix"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/worker/plan"
)

// startBuild starts and monitors a single build. It manages the
// build's state on the Sourcegraph server.
func startBuild(ctx context.Context, build *sourcegraph.Build) {
	done := startHeartbeat(ctx, build.Spec())
	defer done()

	start := time.Now()
	state := newBuildState(build.Spec())

	log15.Info("Starting build", "build", build.Spec().IDString())
	if err := state.start(ctx); err != nil {
		log15.Error("Updating build starting state failed", "build", build.Spec(), "err", err)
		return
	}

	var execErr error
	defer func() {
		if err := state.end(ctx, execErr); err != nil {
			log15.Error("Updating build final state failed", "build", build.Spec(), "err", err)
		}
	}()

	// Execute build.
	builder := builder{
		build: build,
		state: state,
	}

	execErr = builder.exec(ctx)
	if execErr == nil {
		log15.Info("Build succeeded", "build", build.Spec().IDString(), "time", time.Since(start))
	} else {
		log15.Info("Build failed", "build", build.Spec().IDString(), "time", time.Since(start), "err", execErr)
	}
}

// A builder executes a single build. It is called from startBuild and
// does not manage the build's state on the Sourcegraph server (but it
// does manage the state of the tasks).
type builder struct {
	build *sourcegraph.Build // the build (required prior to calling exec)
	state buildState         // for updating the build/task state on the server

	// All other fields are filled in by the builder's own exec method
	// and should not be supplied by the caller.

	cl *sourcegraph.Client

	repoRev sourcegraph.RepoRevSpec
	repo    *sourcegraph.Repo

	config droneyaml.Config // the .drone.yml config (possibly auto-generated)

	opt     droneexec.Options
	payload droneexec.Payload
}

// exec starts executing a build. It is the only method of builder
// that should be called by external callers.
func (b *builder) exec(ctx context.Context) error {
	// Initialize the builder.
	b.cl = sourcegraph.NewClientFromContext(ctx)
	b.repoRev = sourcegraph.RepoRevSpec{
		RepoSpec: sourcegraph.RepoSpec{URI: b.build.Repo},
		Rev:      b.build.CommitID,
		CommitID: b.build.CommitID,
	}

	repo, err := b.cl.Repos.Get(ctx, &b.repoRev.RepoSpec)
	if err != nil {
		return err
	}
	b.repo = repo

	// Get the app URL from the POV of the Docker containers.
	serverConf, err := b.cl.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	_, containerAppURLStr, err := containerAddrForHost(serverConf.AppURL)
	if err != nil {
		return err
	}
	containerAppURL, err := url.Parse(containerAppURLStr)
	if err != nil {
		return err
	}
	ctx = conf.WithAppURL(ctx, containerAppURL)

	// Generate the test plan.
	config, axes, err := plan.CreateServer(ctx, b.repoRev)
	if err != nil {
		// TODO(native-ci): Support a build status that means "unable
		// to automatically configure a build." This does not fit
		// nicely into any of our existing build statuses.
		return err
	}
	b.config = *config

	// Save config as BuilderConfig on the build.
	configYAML, err := marshalConfigWithMatrix(b.config, axes)
	if err != nil {
		return err
	}
	if _, err := b.cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{Build: b.build.Spec(), Info: sourcegraph.BuildUpdate{BuilderConfig: string(configYAML)}}); err != nil {
		return err
	}

	if err := b.prepare(ctx); err != nil {
		return err
	}

	taskLabels := make([]string, len(axes))
	for i, axis := range axes {
		s := fmt.Sprintf("%v", axis)
		s = strings.TrimPrefix(s, "[")
		s = strings.TrimSuffix(s, "]")
		if s == "" {
			s = "Build"
		}
		taskLabels[i] = s
	}
	taskStates, err := b.state.createTasks(ctx, taskLabels...)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errs := make([]error, len(axes))
	for i, axis := range axes {
		wg.Add(1)
		go func(i int, axis matrix.Axis) {
			defer wg.Done()

			errs[i] = func() error {
				taskState := taskStates[i]
				if err := taskState.start(ctx); err != nil {
					log15.Error("Updating task starting state failed", "task", taskState.task, "err", err)
					return err
				}

				var execErr error
				defer func() {
					if err := taskState.end(ctx, execErr); err != nil {
						log15.Error("Updating task final state failed", "task", taskState.task, "err", err)
					}
				}()

				execErr = b.execAxis(ctx, axis, taskState)
				if execErr != nil && len(axes) > 1 {
					fmt.Fprintf(taskState.log, "FAILED: %v\n", axis)
				}
				return execErr
			}()
		}(i, axis)
	}
	wg.Wait()

	if len(axes) == 1 {
		return errs[0]
	}

	// Return a nicely formatted error message describing the result of each matrix axis.
	var msgs []string
	for i, err := range errs {
		if err != nil {
			msgs = append(msgs, fmt.Sprintf("%v (%s)", axes[i], err))
		}
	}
	if msgs != nil {
		return fmt.Errorf("%d/%d failed: %s", len(msgs), len(axes), strings.Join(msgs, ", "))
	}
	return nil
}

// execAxis executes one axis of a build. (An axis is one combination
// of variables in a matrix build. A build that doesn't use matrix has
// a single axis with no variables.)
func (b *builder) execAxis(ctx context.Context, axis matrix.Axis, taskState buildTaskState) error {
	// Deep-copy payload so we can inject this axis's variables and
	// pass to droneexec.Exec without worrying about concurrent
	// modifications by other execAxis invocations.
	data, err := json.Marshal(b.payload)
	if err != nil {
		return err
	}
	var axisPayload droneexec.Payload
	if err := json.Unmarshal(data, &axisPayload); err != nil {
		return err
	}

	axisPayload.Job.Environment = axis

	opt := b.opt // copy
	opt.Monitor = func(section, key string, node droneparser.Node) dronerunner.Monitor {
		capFirst := func(s string) string {
			if s == "" {
				return s
			}
			return strings.ToUpper(s[0:1]) + s[1:]
		}

		var label string
		if section == "build" && (key == "build" || key == "") {
			label = "Build"
		} else if section == "build" {
			label = key
		} else if section != "" && key != "" {
			label = capFirst(section) + ": " + key
		} else {
			label = capFirst(section) + key
		}

		subtaskState, err := taskState.createSubtask(ctx, label)
		if err != nil {
			log15.Error("Creating subtask failed", "task", taskState.task, "err", err)
			return noopMonitor{}
		}

		// Currently the only way to mark a task as having warnings is
		// if the label contains the string "warning".
		if strings.Contains(strings.ToLower(label), "warning") {
			if err := subtaskState.warnings(ctx); err != nil {
				log15.Error("Marking subtask as having warnings failed", "subtask", subtaskState.task, "err", err)
			}
		}

		return taskMonitor{ctx: ctx, s: subtaskState}
	}

	return droneexec.Exec(axisPayload, opt)
}

// taskMonitor is a Drone task monitor.
type taskMonitor struct {
	ctx context.Context
	s   *buildTaskState
}

func (m taskMonitor) Start() {
	if err := m.s.start(m.ctx); err != nil {
		log15.Error("Error marking monitored task as started", "task", m.s.task, "err", err)
	}
}

func (m taskMonitor) Skip() {
	if err := m.s.skip(m.ctx); err != nil {
		log15.Error("Error marking monitored task as skipped", "task", m.s.task, "err", err)
	}
}

func (m taskMonitor) End(ok, allowFailure bool) {
	if !ok && allowFailure {
		if err := m.s.warnings(m.ctx); err != nil {
			log15.Error("Error marking monitored task as ended (with allowable failure)", "task", m.s.task, "err", err)
		}
	} else {
		var execErr error
		if !ok {
			execErr = errors.New("failed")
		}
		if err := m.s.end(m.ctx, execErr); err != nil {
			log15.Error("Error marking monitored task as ended", "task", m.s.task, "err", err)
		}
	}
}

func (m taskMonitor) Logger() (stdout, stderr io.Writer) {
	return m.s.log, m.s.log
}

// noopMonitor is a no-op Drone task monitor. It is necessary because
// MonitorFuncs may never return nil (even if an error occurs while
// generating the monitor).
type noopMonitor struct{}

func (noopMonitor) Start() {}

func (noopMonitor) Skip() {}

func (noopMonitor) End(ok, allowFailure bool) {}

func (noopMonitor) Logger() (stdout, stderr io.Writer) { return ioutil.Discard, ioutil.Discard }

func now() *pbtypes.Timestamp {
	now := pbtypes.NewTimestamp(time.Now())
	return &now
}
