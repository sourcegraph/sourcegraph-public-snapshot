package run

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"

	"github.com/nxadm/tail"

	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type IBazel struct {
	targets   []string
	events    chan iBazelEvent
	eventsDir string
	dir       string
	proc      *startedCmd
}

func (ibazel *IBazel) GetName() string {
	return fmt.Sprintf("bazel targets (%s)", strings.Join(ibazel.targets, ", "))
}

func (ibazel *IBazel) RunInstall(ctx context.Context, env map[string]string) error {
	if len(ibazel.targets) == 0 {
		// no Bazel commands so we return
		return nil
	}

	err := ibazel.Build(ctx)
	if err != nil {
		return err
	}

	p := pool.New().WithContext(ctx).WithCancelOnError()

	p.Go(func(ctx context.Context) error {
		return ibazel.Watch(ctx)
	})

	// block until initial ibazel build is completed
	return ibazel.WaitForInitialBuild(ctx)
}

func (ib *IBazel) GetExec(ctx context.Context) *exec.Cmd {
	// Writes iBazel events out to a log file. These are much easier to parse
	// than trying to understand the output directly
	profilePath := "--profile_dev=" + ib.profileEventsFilePath()
	// This enables iBazel to try to apply the fixes from .bazel_fix_commands.json automatically
	enableAutoFix := "--run_output_interactive=false"
	args := append([]string{profilePath, enableAutoFix, "build"}, ib.targets...)
	return exec.CommandContext(ctx, "ibazel", args...)
}

// returns a runner to interact with ibazel.
func NewIBazel(cmds []BazelCommand, dir string) (*IBazel, error) {
	eventsDir, err := os.MkdirTemp("", "ibazel-events")
	if err != nil {
		return nil, err
	}
	eventsFile, err := os.Create(profileEventsFilePath(eventsDir))
	if err != nil {
		return nil, err
	}
	if err = eventsFile.Close(); err != nil {
		return nil, err
	}

	targets := make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		if !slices.Contains(targets, cmd.Target) {
			targets = append(targets, cmd.Target)
		}
	}

	return &IBazel{
		targets:   targets,
		events:    make(chan iBazelEvent),
		eventsDir: eventsDir,
		dir:       dir,
	}, nil
}

func (ib *IBazel) profileEventsFilePath() string {
	return profileEventsFilePath(ib.eventsDir)
}

func profileEventsFilePath(eventsDir string) string {
	return path.Join(eventsDir, "profile.json")
}

// Watch opens the provided profile.json and reads it as it is continuously written by iBazel
// Each time it sees a iBazel event log, it parses it and puts it on the events channel
func (ib *IBazel) Watch(ctx context.Context) error {
	tail, err := tail.TailFile(ib.profileEventsFilePath(), tail.Config{Follow: true, ReOpen: true})
	if err != nil {
		return err
	}
	for line := range tail.Lines {
		var event iBazelEvent
		if err := json.Unmarshal([]byte(line.Text), &event); err != nil {
			return errors.Newf("failed to unmarshal event json: %s", err)
		}
		ib.events <- event
	}
	return nil
}

func (ib *IBazel) WaitForInitialBuild(ctx context.Context) error {
	for event := range ib.events {
		if event.Type == buildDone {
			return nil
		}
		if event.Type == buildFailed {
			return errors.Newf("initial ibazel build failed")
		}
	}

	return nil
}

func (ib *IBazel) getCommandOptions(ctx context.Context) commandOptions {
	return commandOptions{
		name: "iBazel",
		exec: ib.GetExec(ctx),
		dir:  ib.dir,
		// Don't output iBazel logs until initial build is complete
		// as it will break the progress bar
		bufferOutput: true,
	}
}

// Build starts an ibazel process to build the targets provided in the constructor
// It runs perpetually, watching for file changes
func (ib *IBazel) Build(ctx context.Context) (err error) {
	ib.proc, err = startCmd(ctx, ib.getCommandOptions(ctx))
	return err
}

func (ib *IBazel) StartOutput() error {
	return ib.proc.StartOutput()
}

func (ib *IBazel) Stop() {
	os.RemoveAll(ib.eventsDir)
	ib.proc.cancel()
}

// Schema information at https://github.com/bazelbuild/bazel-watcher?tab=readme-ov-file#profiler-events
type iBazelEvent struct {
	// common
	Type      string   `json:"type"`
	Iteration string   `json:"iteration"`
	Time      int64    `json:"time"`
	Targets   []string `json:"targets,omitempty"`
	Elapsed   int64    `json:"elapsed,omitempty"`

	// start event
	IBazelVersion     string `json:"iBazelVersion,omitempty"`
	BazelVersion      string `json:"bazelVersion,omitempty"`
	MaxHeapSize       string `json:"maxHeapSize,omitempty"`
	CommittedHeapSize string `json:"committedHeapSize,omitempty"`

	// change event
	Change string `json:"change,omitempty"`

	// build & reload event
	Changes []string `json:"changes,omitempty"`

	// browser event
	RemoteType    string `json:"remoteType,omitempty"`
	RemoteTime    int64  `json:"remoteTime,omitempty"`
	RemoteElapsed int64  `json:"remoteElapsed,omitempty"`
	RemoteData    string `json:"remoteData,omitempty"`
}

const (
	buildDone   = "BUILD_DONE"
	buildFailed = "BUILD_FAILED"
)
