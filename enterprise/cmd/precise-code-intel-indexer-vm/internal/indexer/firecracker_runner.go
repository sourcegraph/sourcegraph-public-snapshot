package indexer

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

const firecrackerMountPoint = "/repo-dir"

type firecrackerRunner struct {
	commander Commander
	options   HandlerOptions
	//
	repoDir      string
	name         string
	index        store.Index
	images       map[string]string
	dockerRunner Runner
}

var _ Runner = &firecrackerRunner{}

func NewFirecrackerRunner(
	commander Commander,
	options HandlerOptions,
	repoDir string,
	name string,
	index store.Index,
) Runner {
	images := requiredImages(index)
	dockerRunner := NewDockerRunner(commander, options, repoDir, index)

	return &firecrackerRunner{
		commander:    commander,
		options:      options,
		repoDir:      repoDir,
		name:         name,
		index:        index,
		images:       images,
		dockerRunner: dockerRunner,
	}
}

const srcCliImage = "sourcegraph/src-cli:latest"

func requiredImages(index store.Index) map[string]string {
	images := map[string]string{
		"src-cli": srcCliImage,
		"indexer": index.Indexer,
	}
	for i, step := range index.DockerSteps {
		// TODO - deduplicate
		images[fmt.Sprintf("%d", i)] = step.Image
	}

	return images
}

var firecrackerCommonFlags = []string{
	"--runtime", "docker",
	"--network-plugin", "docker-bridge",
}

func (r *firecrackerRunner) Startup(ctx context.Context) error {
	if err := r.ensureTarfilesOnHost(ctx, r.images); err != nil {
		return err
	}

	runArgs := concatAll(
		"ignite", "run", "--name", r.name, "--ssh",
		firecrackerCommonFlags,
		r.resourceFlags(),
		r.copyFilesFlags(ctx),
		sanitizeImage(r.options.FirecrackerImage),
	)
	if err := r.commander.Run(ctx, runArgs...); err != nil {
		return errors.Wrap(err, "failed to start firecracker vm")
	}

	for _, key := range sortKeys(r.images) {
		copyArgs := concatAll(
			"ignite", "exec", r.name, "--",
			"docker", "load",
			"-i", r.tarfilePathOnVirtualMachine(key),
		)

		if err := r.commander.Run(ctx, copyArgs...); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to load %s", r.images[key]))
		}
	}

	return r.dockerRunner.Startup(ctx)
}

func (r *firecrackerRunner) Teardown(ctx context.Context) error {
	stopArgs := concatAll(
		"ignite", "stop",
		firecrackerCommonFlags,
		r.name,
	)
	if err := r.commander.Run(ctx, stopArgs...); err != nil {
		log15.Warn("failed to stop firecracker vm", "name", r.name, "err", err)
	}

	removeArgs := concatAll(
		"ignite", "rm", "-f",
		firecrackerCommonFlags,
		r.name,
	)
	if err := r.commander.Run(ctx, removeArgs...); err != nil {
		log15.Warn("failed to remove firecracker vm", "name", r.name, "err", err)
	}

	return r.dockerRunner.Teardown(ctx)
}

func (r *firecrackerRunner) Invoke(ctx context.Context, image string, cs *CommandSpec) error {
	return r.commander.Run(ctx, r.MakeArgs(ctx, image, cs, firecrackerMountPoint)...)
}

func (r *firecrackerRunner) MakeArgs(ctx context.Context, image string, cs *CommandSpec, mountPoint string) []string {
	return concatAll(
		"ignite", "exec", r.name, "--",
		r.dockerRunner.MakeArgs(ctx, image, cs, mountPoint),
	)
}

func (r *firecrackerRunner) ensureTarfilesOnHost(ctx context.Context, images map[string]string) error {
	for _, key := range sortKeys(images) {
		if ok, err := fileExists(r.tarfilePathOnHost(key)); err != nil {
			return err
		} else if ok {
			continue
		}

		pullArgs := concatAll(
			"docker", "pull",
			images[key],
		)
		if err := r.commander.Run(ctx, pullArgs...); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to pull %s", images[key]))
		}

		saveArgs := concatAll(
			"docker", "save",
			"-o", r.tarfilePathOnHost(key),
			images[key],
		)
		if err := r.commander.Run(ctx, saveArgs...); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to save %s", images[key]))
		}
	}

	return nil
}

func (r *firecrackerRunner) resourceFlags() []string {
	return []string{"--cpus", strconv.Itoa(r.options.FirecrackerNumCPUs), "--memory", r.options.FirecrackerMemory}
}

func (r *firecrackerRunner) copyFilesFlags(ctx context.Context) []string {
	var copyfiles []string
	for _, key := range sortKeys(r.images) {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", r.tarfilePathOnHost(key), r.tarfilePathOnVirtualMachine(key)))
	}

	return prefixKeys("--copy-files", append(
		[]string{fmt.Sprintf("%s:%s", r.repoDir, firecrackerMountPoint)},
		copyfiles...,
	))
}

func (r *firecrackerRunner) tarfilePathOnHost(key string) string {
	return filepath.Join(r.options.ImageArchivePath, fmt.Sprintf("%s.tar", key))
}

func (r *firecrackerRunner) tarfilePathOnVirtualMachine(key string) string {
	return fmt.Sprintf("/%s.tar", key)
}
