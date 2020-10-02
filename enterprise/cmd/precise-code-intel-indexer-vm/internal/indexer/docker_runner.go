package indexer

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type dockerRunner struct {
	commander Commander
	options   HandlerOptions
	//
	repoDir string
	index   store.Index
}

var _ Runner = &dockerRunner{}

func NewDockerRunner(
	commander Commander,
	options HandlerOptions,
	repoDir string,
	index store.Index,
) Runner {
	return &dockerRunner{
		commander: commander,
		options:   options,
		repoDir:   repoDir,
		index:     index,
	}
}

const dockerDataDir = "/data"

func (r *dockerRunner) Startup(ctx context.Context) error {
	return nil
}

func (r *dockerRunner) Teardown(ctx context.Context) error {
	return nil
}

func (r *dockerRunner) Invoke(ctx context.Context, image string, cs *CommandSpec) error {
	return r.commander.Run(ctx, r.MakeArgs(ctx, image, cs, r.repoDir)...)
}

func (r *dockerRunner) MakeArgs(ctx context.Context, image string, cs *CommandSpec, mountPoint string) []string {
	return concatAll(
		"docker", "run", "--rm",
		r.resourceFlags(),
		r.volumeMountFlags(mountPoint),
		r.workingDirectoryFlags(),
		r.envFlags(cs),
		image,
		cs.command,
	)
}

func (r *dockerRunner) resourceFlags() []string {
	return []string{"--cpus", strconv.Itoa(r.options.FirecrackerNumCPUs), "--memory", r.options.FirecrackerMemory}
}

func (r *dockerRunner) volumeMountFlags(mountPoint string) []string {
	return []string{"-v", fmt.Sprintf("%s:%s", mountPoint, dockerDataDir)}
}

func (r *dockerRunner) workingDirectoryFlags() []string {
	return []string{"-w", filepath.Join(dockerDataDir, r.index.Root)}
}

func (r *dockerRunner) envFlags(cs *CommandSpec) []string {
	var envFlags []string
	for _, k := range sortKeys(cs.env) {
		envFlags = append(envFlags, fmt.Sprintf("%s=%s", k, cs.env[k]))
	}

	return prefixKeys("-e", envFlags)
}
