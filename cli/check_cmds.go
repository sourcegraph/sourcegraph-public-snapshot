package cli

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/inconshreveable/log15.v2"

	droneexec "github.com/drone/drone-exec/exec"
	"github.com/drone/drone-plugin-go/plugin"
	"golang.org/x/net/context"
	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vfsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/worker/builder"
)

func init() {
	_, err := cli.CLI.AddCommand("check",
		"check code (run CI locally)",
		"The check subcommand checks code. It lets you run CI locally using virtually the same configuration and (Docker) environments that the server uses.",
		&checkCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type checkCmd struct {
	Cache  bool   `long:"cache" description:"execute cache steps"`
	Deploy bool   `long:"deploy" description:"execute publish and deploy steps"`
	Notify bool   `long:"notify" description:"execute notify steps"`
	Debug  bool   `long:"debug" description:"execute the build in debug mode"`
	Dir    string `long:"dir" description:"directory tree to check" default:"."`

	ShowConfig bool `long:"show-config" description:"show .sg-drone.yml config file (with inference & srclib additions) and exit"`
}

func (c *checkCmd) Execute(args []string) error {
	var err error
	c.Dir, err = filepath.Abs(c.Dir)
	if err != nil {
		return err
	}

	builder, err := c.configureBuilder(cliContext)
	if err != nil {
		return err
	}

	if err := builder.Exec(cliContext); err != nil {
		fmt.Println(red("FAIL"))
		return err
	}
	fmt.Println(green("PASS"))
	return nil
}

func (c *checkCmd) configureBuilder(ctx context.Context) (*builder.Builder, error) {
	var b builder.Builder
	fs := vfs.OS(c.Dir)

	guessPath := func(dir string) string {
		path := filepath.Join(os.Getenv("GOPATH"), "src")
		if filepath.HasPrefix(dir, path) {
			return strings.TrimPrefix(dir, path+string(os.PathSeparator))
		}

		parts := strings.Split(dir, string(os.PathSeparator)+"src"+string(os.PathSeparator))
		if len(parts) > 1 {
			path := parts[len(parts)-1]
			return strings.TrimPrefix(path, string(os.PathSeparator))
		}
		return filepath.Base(dir)
	}

	// Drone payload
	b.Payload = droneexec.Payload{
		Repo: &plugin.Repo{
			IsPrivate: true,
			IsTrusted: true,
			Link:      "https://" + guessPath(c.Dir),
		},
		Workspace: &plugin.Workspace{},
		Build: &plugin.Build{
			Status: plugin.StateRunning,
			Commit: "0000000000",
		},
		Job: &plugin.Job{
			Status: plugin.StateRunning,
		},
		System: &plugin.System{
			Globals: []string{},
			Plugins: []string{"plugins/*", "*/*"},
		},
	}

	// .sg-drone.yml
	yamlBytes, err := vfs.ReadFile(fs, ".sg-drone.yml")
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if err == nil {
		b.Payload.Yaml = string(yamlBytes)
		b.DroneYMLFileExists = true
	}

	// Drone options
	b.Options = droneexec.Options{
		Cache:  c.Cache,
		Build:  true,
		Deploy: c.Deploy,
		Notify: c.Notify,
		Debug:  c.Debug,
		Mount:  c.Dir,
	}

	// Inventory
	b.Inventory = func(ctx context.Context) (*inventory.Inventory, error) {
		return inventory.Scan(ctx, vfsutil.Walkable(fs, filepath.Join))
	}

	// CreateTasks
	b.CreateTasks = func(ctx context.Context, labels []string) ([]builder.TaskState, error) {
		states := make([]builder.TaskState, len(labels))
		for i, label := range labels {
			states[i] = &taskState{
				label: label,
				log:   nopWriteCloser{os.Stderr},
			}
		}
		return states, nil
	}

	// FinalBuildConfig
	b.FinalBuildConfig = func(ctx context.Context, configYAML string) error {
		if c.ShowConfig {
			os.Exit(0)
		}

		return nil
	}

	return &b, nil
}

// taskState manages the state of a task running via `src check`. It
// implements builder.TaskState.
type taskState struct {
	label string

	// log is where task logs are written. Internal errors
	// encountered by the builder are not written to w but are
	// returned as errors from its methods.
	log io.WriteCloser
}

// Start implements builder.TaskState.
func (s taskState) Start(ctx context.Context) error {
	log15.Info("START", "task", s)
	return nil
}

// Skip implements builder.TaskState.
func (s taskState) Skip(ctx context.Context) error {
	log15.Info("SKIP", "task", s)
	return nil
}

// Warnings implements builder.TaskState.
func (s taskState) Warnings(ctx context.Context) error {
	log15.Warn("WARNINGS", "task", s)
	return nil
}

// End implements builder.TaskState.
func (s taskState) End(ctx context.Context, execErr error) error {
	defer s.log.Close()

	if execErr == nil {
		log15.Info("PASS", "task", s)
	} else {
		log15.Error("FAIL", "task", s, "err", execErr)
	}
	return nil
}

// CreateSubtask implements builder.TaskState.
func (s taskState) CreateSubtask(ctx context.Context, label string) (builder.TaskState, error) {
	return &taskState{
		label: label,
		log:   s.log,
	}, nil
}

func (s taskState) Log() io.Writer { return s.log }

func (s taskState) String() string { return s.label }

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }
