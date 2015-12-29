package sgx

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	droneexec "github.com/drone/drone-exec/exec"
	droneyaml "github.com/drone/drone-exec/yaml"
	"github.com/drone/drone-plugin-go/plugin"
	"github.com/drone/drone/yaml/matrix"
	"golang.org/x/tools/godoc/vfs"

	"strings"

	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/worker/plan"
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
}

func (c *checkCmd) Execute(args []string) error {
	var err error
	c.Dir, err = filepath.Abs(c.Dir)
	if err != nil {
		return err
	}

	config, axes, err := plan.CreateLocal(cli.Ctx, vfs.OS(c.Dir))
	if err != nil {
		return err
	}

	if c.Debug {
		yamlBytes, err := yaml.Marshal(config)
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "# .drone.yml file:")
		fmt.Fprintln(os.Stderr, string(yamlBytes))
		fmt.Fprintln(os.Stderr, "# Build matrix:")
		for _, axis := range axes {
			fmt.Fprintf(os.Stderr, " - %s\n", axis)
		}
		fmt.Fprintln(os.Stderr)
	}

	success := true
	for i, axis := range axes {
		if len(axis) > 0 {
			fmt.Printf(cyan("# %v\n"), axis)
		}
		if err := c.execAxis(config, axis); err != nil {
			fmt.Println(red("FAIL"))
			success = false
		} else {
			fmt.Println(green("PASS"))
		}
		if i != len(axes)-1 {
			fmt.Println()
		}
	}

	if !success {
		if len(axes) > 1 {
			fmt.Println()
			fmt.Println(red("OVERALL: FAIL"))
		}
		os.Exit(1)
	}
	return nil
}

func (c *checkCmd) execAxis(config *droneyaml.Config, axis matrix.Axis) error {
	opt := droneexec.Options{
		Cache:  c.Cache,
		Build:  true,
		Deploy: c.Deploy,
		Notify: c.Notify,
		Debug:  c.Debug,
		Mount:  c.Dir,
	}

	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	payload := droneexec.Payload{
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
			Status:      plugin.StateRunning,
			Environment: axis,
		},
		System: &plugin.System{
			Globals: []string{},
			Plugins: []string{"plugins/*", "*/*"},
		},
		Yaml: string(yamlBytes),
	}

	return droneexec.Exec(payload, opt)
}

func guessPath(dir string) string {
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
