package builder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"

	droneexec "github.com/drone/drone-exec/exec"
	droneparser "github.com/drone/drone-exec/parser"
	dronerunner "github.com/drone/drone-exec/runner"
	"github.com/drone/drone/yaml/matrix"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/services/worker/plan"
)

// A Builder executes a single build (with its Exec func). It manages
// state by calling CreateTasks as well as by using the returned
// TaskState objects for each task.
//
// It is not inherently aware of a Sourcegraph server; the parent
// worker package passes it a CreateTasks func that does perform
// operations on the Sourcegraph server, but it can run just as well
// with an in-memory CreateTasks func (for `src check` local builds).
type Builder struct {
	droneexec.Options
	droneexec.Payload

	// Other options

	// DroneYMLFileExists is true if the repo has a .drone.yml
	// file. This is necessary because an empty string in Payload.Yaml
	// could mean either "empty file" or "no file".
	DroneYMLFileExists bool

	// SrclibImportURL is the URL to POST srclib data during the
	// import. If not set, no srclib import is performed.
	SrclibImportURL *url.URL

	// Inventory performs a repo inventory (or returns an existing
	// cached copy of the inventory), to determine which languages,
	// etc., are in use by the repo.
	Inventory func(ctx context.Context) (*inventory.Inventory, error)

	// CreateTasks is a required func that is called to create tasks
	// of this build (e.g., on the Sourcegraph server, or in-memory if
	// running locally with `src check`).
	CreateTasks func(ctx context.Context, labels []string) ([]TaskState, error)

	// FinalBuildConfig, if non-nil, is called with the final build
	// configuration (.drone.yml) after inferred and srclib steps have
	// been added.
	FinalBuildConfig func(ctx context.Context, configYAML string) error
}

// TaskState manages a task's state. An implementation could, for
// example, perform operations against a Sourcegraph server to track
// state (this is what happens in the worker). Or it could just
// perform them in memory and update the terminal display (this is
// what `src check` does).
type TaskState interface {
	// Start marks the task as having started.
	Start(ctx context.Context) error

	// Skip marks the task as having been skipped.
	Skip(ctx context.Context) error

	// Warnings marks the task as having warnings.
	Warnings(ctx context.Context) error

	// End updates the task's final state.
	End(ctx context.Context, execErr error) error

	CreateSubtask(ctx context.Context, label string) (TaskState, error)

	Log() io.Writer

	String() string
}

// Exec starts executing a build.
func (b *Builder) Exec(ctx context.Context) error {
	finalConfig, axes, err := b.plan(ctx)
	if err != nil {
		return err
	}
	b.Payload.Yaml = finalConfig

	// Create tasks.
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
	taskStates, err := b.CreateTasks(ctx, taskLabels)
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
				if err := taskState.Start(ctx); err != nil {
					if ctx.Err() == nil {
						log15.Error("Updating task starting state failed", "task", taskState.String(), "err", err)
					}
					return err
				}

				var execErr error
				defer func() {
					if err := taskState.End(ctx, execErr); err != nil && ctx.Err() == nil {
						log15.Error("Updating task final state failed", "task", taskState.String(), "err", err)
					}
				}()

				return b.execAxis(ctx, axis, taskState)
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

func (b *Builder) plan(ctx context.Context) (finalConfig string, axes []matrix.Axis, err error) {
	tasks, err := b.CreateTasks(ctx, []string{"Configure build"})
	if err != nil {
		return "", nil, err
	}
	task := tasks[0]

	defer func() {
		if err2 := task.End(ctx, err); err2 != nil {
			if err == nil {
				err = err2
				return
			}
			log15.Error("Error marking plan task as ended", "task", task, "err", err2)
		}
	}()
	if err := task.Start(ctx); err != nil {
		return "", nil, err
	}

	var planLabel string
	if b.DroneYMLFileExists {
		planLabel = "Add srclib indexing steps to existing .drone.yml"
	} else {
		planLabel = "Infer build & test configuration"
	}

	// Assigned by the subtaskFuncs fn funcs below.
	var inv *inventory.Inventory

	subtaskFuncs := []struct {
		label string
		fn    func(ctx context.Context, state TaskState) error
	}{
		{
			label: "Detect projects",
			fn: func(ctx context.Context, state TaskState) error {
				var err error
				inv, err = b.Inventory(ctx)

				if err == nil {
					w := state.Log()
					langs := make([]*inventory.Lang, 0, len(inv.Languages))
					skipped := make([]*inventory.Lang, 0, len(inv.Languages))
					for _, l := range inv.Languages {
						if _, ok := skipLangs[l.Name]; ok {
							skipped = append(skipped, l)
						} else {
							langs = append(langs, l)
						}
					}
					inv.Languages = langs
					if len(inv.Languages) == 0 {
						if !b.DroneYMLFileExists {
							fmt.Fprintln(w, "No recognized programming languages were detected in this repository.")
						}
					} else {
						fmt.Fprintf(w, "Detected %d programming languages in use by this repository:\n", len(inv.Languages))
						for _, lang := range inv.Languages {
							fmt.Fprintf(w, " - %s\n", lang.Name)
						}
					}
					if len(skipped) > 0 {
						fmt.Fprintf(w, "\nFound %d languages in use by this repository that will not be built or analysed:\n", len(skipped))
						for _, lang := range skipped {
							fmt.Fprintf(w, " - %s\n", lang.Name)
						}
					}
				}
				return err
			},
		},
		{
			label: planLabel,
			fn: func(ctx context.Context, state TaskState) error {
				finalConfig, axes, err = plan.Create(b.Payload.Yaml, b.DroneYMLFileExists, inv, b.SrclibImportURL)
				if err != nil {
					return err
				}

				w := state.Log()
				if b.DroneYMLFileExists {
					fmt.Fprintln(w, "# Using .drone.yml file with srclib indexing steps added.")
				} else {
					fmt.Fprintln(w, "# Because this repository has no .drone.yml file, Sourcegraph attempted to infer this repository's build and test configuration.")
					fmt.Fprintln(w, "#")
					fmt.Fprintln(w, "# If this configuration is incorrect or incomplete, add a .drone.yml file to your repository (using this as a starter) with the correct configuration. See http://readme.drone.io/usage/overview/ for instructions.")
					fmt.Fprintln(w, "#")
					fmt.Fprintln(w, "# Tip: You can test your .drone.yml locally by running `src check` in your repository, after downloading the `src` CLI.")
				}

				fmt.Fprintln(state.Log())
				fmt.Fprintf(state.Log(), finalConfig)

				if b.FinalBuildConfig != nil {
					if err := b.FinalBuildConfig(ctx, finalConfig); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}

	for _, stf := range subtaskFuncs {
		err2 := func() (err error) {
			// Run in closure so we can use defer and so we can treat
			// this body's returned error different from the outer
			// func's.

			subtask, err := task.CreateSubtask(ctx, stf.label)
			if err != nil {
				return err
			}

			defer func() {
				if err != nil {
					fmt.Fprintln(subtask.Log(), "FAIL:", err)
				}
				if err2 := subtask.End(ctx, err); err2 != nil {
					if err == nil {
						err = err2
						return
					}
					log15.Error("Error marking plan subtask as ended", "task", task, "err", err2)
				}
			}()
			if err := subtask.Start(ctx); err != nil {
				return err
			}

			return stf.fn(ctx, subtask)
		}()
		if err2 != nil {
			// Use a different error to avoid confusion that would
			// arise if the subtask and the parent task ended with the
			// same error.
			return "", nil, fmt.Errorf("plan subtask failed: %q", stf.label)
		}
	}

	return
}

// execAxis executes one axis of a build. (An axis is one combination
// of variables in a matrix build. A build that doesn't use matrix has
// a single axis with no variables.)
func (b *Builder) execAxis(ctx context.Context, axis matrix.Axis, taskState TaskState) error {
	// Deep-copy the payload so we can inject this axis's variables
	// and pass to droneexec.Exec without worrying about concurrent
	// modifications by other execAxis invocations. (We use JSON just
	// to deep-copy, not because this has anything to do with JSON.)
	data, err := json.Marshal(b.Payload)
	if err != nil {
		return err
	}
	var axisPayload droneexec.Payload
	if err := json.Unmarshal(data, &axisPayload); err != nil {
		return err
	}

	axisPayload.Job.Environment = axis

	opt := b.Options // copy
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

		subtaskState, err := taskState.CreateSubtask(ctx, label)
		if err != nil {
			log15.Error("Creating subtask failed", "task", taskState.String(), "err", err)
			return noopMonitor{}
		}

		// Currently the only way to mark a task as having warnings is
		// if the label contains the string "warning".
		if strings.Contains(strings.ToLower(label), "warning") {
			if err := subtaskState.Warnings(ctx); err != nil {
				log15.Error("Marking subtask as having warnings failed", "subtask", subtaskState.String(), "err", err)
			}
		}

		return taskMonitor{ctx: ctx, s: subtaskState}
	}

	return ignoreBtrfsCircleCIError(droneexec.Exec(ctx, axisPayload, opt))
}

// Ignores errors due to CircleCI forbidding certain operations inside
// their unprivileged lxc containers. See
// https://discuss.circleci.com/t/docker-error-removing-intermediate-container/70.
func ignoreBtrfsCircleCIError(err error) error {
	const msg = "Failed to destroy btrfs snapshot: operation not permitted"
	if os.Getenv("CIRCLECI") != "" && strings.Contains(err.Error(), msg) {
		return nil
	}
	return err
}

// taskMonitor is a Drone task monitor.
type taskMonitor struct {
	ctx context.Context
	s   TaskState
}

func (m taskMonitor) Start() {
	if err := m.s.Start(m.ctx); err != nil {
		log15.Error("Error marking monitored task as started", "task", m.s.String(), "err", err)
	}
}

func (m taskMonitor) Skip() {
	if err := m.s.Skip(m.ctx); err != nil {
		log15.Error("Error marking monitored task as skipped", "task", m.s.String(), "err", err)
	}
}

func (m taskMonitor) End(ok, allowFailure bool) {
	if !ok && allowFailure {
		if err := m.s.Warnings(m.ctx); err != nil && m.ctx.Err() == nil {
			log15.Error("Error marking monitored task as ended (with allowable failure)", "task", m.s.String(), "err", err)
		}
		if err := m.s.End(m.ctx, nil); err != nil && m.ctx.Err() == nil {
			log15.Error("Error marking monitored task as ended (with warnings)", "task", m.s.String(), "err", err)
		}
	} else {
		var execErr error
		if !ok {
			execErr = errors.New("failed")
		}
		if err := m.s.End(m.ctx, execErr); err != nil && m.ctx.Err() == nil {
			log15.Error("Error marking monitored task as ended", "task", m.s.String(), "err", err)
		}
	}
}

func (m taskMonitor) Logger() (stdout, stderr io.Writer) {
	return m.s.Log(), m.s.Log()
}

// noopMonitor is a no-op Drone task monitor. It is necessary because
// MonitorFuncs may never return nil (even if an error occurs while
// generating the monitor).
type noopMonitor struct{}

func (noopMonitor) Start() {}

func (noopMonitor) Skip() {}

func (noopMonitor) End(ok, allowFailure bool) {}

func (noopMonitor) Logger() (stdout, stderr io.Writer) { return ioutil.Discard, ioutil.Discard }

// skipLangs are languages we shouldn't build or analyse. They are usually
// languages that are common in repos, but including them in warnings would be
// noisy.
var skipLangs = map[string]struct{}{
	"Ant Build System": struct{}{},
	"Batchfile":        struct{}{},
	"Dockerfile":       struct{}{},
	"Graphviz (DOT)":   struct{}{},
	"Makefile":         struct{}{},
	"Markdown":         struct{}{},
	"PLpgSQL":          struct{}{},
	"SaltStack":        struct{}{},
	"Tcsh":             struct{}{},
	"YAML":             struct{}{},
	"fish":             struct{}{},
}
