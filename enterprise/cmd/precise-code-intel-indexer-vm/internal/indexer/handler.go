package indexer

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer-vm/internal/index_manager"
	queue "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/queue/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type Handler struct {
	queueClient   queue.Client
	indexManager  *indexmanager.Manager
	commander     Commander
	options       HandlerOptions
	uuidGenerator func() (uuid.UUID, error)
}

var _ workerutil.Handler = &Handler{}

type HandlerOptions struct {
	FrontendURL           string
	FrontendURLFromDocker string
	AuthToken             string
	FirecrackerImage      string
	UseFirecracker        bool
	FirecrackerNumCPUs    int
	FirecrackerMemory     string
	ImageArchivePath      string
}

// Handle clones the target code into a temporary directory, invokes the target indexer in a fresh
// docker container, and uploads the results to the external frontend API.
func (h *Handler) Handle(ctx context.Context, _ workerutil.Store, record workerutil.Record) error {
	index := record.(store.Index)

	h.indexManager.AddID(index.ID)
	defer h.indexManager.RemoveID(index.ID)

	repoDir, err := h.fetchRepository(ctx, index.RepositoryName, index.Commit)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(repoDir)
	}()

	uploadURL, err := makeUploadURL(h.options.FrontendURLFromDocker, h.options.AuthToken)
	if err != nil {
		return err
	}

	name, err := h.uuidGenerator()
	if err != nil {
		return err
	}

	mountPoint := repoDir
	if h.options.UseFirecracker {
		mountPoint = "/repo-dir"

		images := map[string]string{
			"lsif-go": "sourcegraph/lsif-go:latest",
			"src-cli": "sourcegraph/src-cli:latest",
		}

		copyfiles := []string{}
		for _, key := range orderedKeys(images) {
			copyfiles = append(copyfiles, "--copy-files", fmt.Sprintf("%s:%s", h.tarfilePathOnHost(key), h.tarfilePathInVM(key)))
		}

		for _, key := range orderedKeys(images) {
			if _, err := os.Stat(h.tarfilePathOnHost(key)); err == nil {
				continue
			} else if !os.IsNotExist(err) {
				return err
			}

			if err := h.saveDockerImage(ctx, key, images[key]); err != nil {
				return err
			}
		}

		startCommand := flatten(
			"ignite", "run",
			"--runtime", "docker",
			"--network-plugin", "docker-bridge",
			"--cpus", strconv.Itoa(h.options.FirecrackerNumCPUs),
			"--memory", h.options.FirecrackerMemory,
			"--copy-files", fmt.Sprintf("%s:%s", repoDir, mountPoint),
			copyfiles,
			"--ssh",
			"--name", name.String(),
			sanitizeImage(h.options.FirecrackerImage),
		)
		if err := h.commander.Run(ctx, startCommand...); err != nil {
			return errors.Wrap(err, "failed to start firecracker vm")
		}
		defer func() {
			stopCommand := flatten(
				"ignite", "stop",
				"--runtime", "docker",
				"--network-plugin", "docker-bridge",
				name.String(),
			)
			if err := h.commander.Run(ctx, stopCommand...); err != nil {
				log15.Warn("failed to stop firecracker vm", "name", name.String(), "err", err)
			}

			removeCommand := flatten(
				"ignite", "rm", "-f",
				"--runtime", "docker",
				"--network-plugin", "docker-bridge",
				name.String(),
			)
			if err := h.commander.Run(ctx, removeCommand...); err != nil {
				log15.Warn("failed to remove firecracker vm", "name", name.String(), "err", err)
			}
		}()

		for _, key := range orderedKeys(images) {
			loadCommand := flatten(
				"ignite", "exec", name.String(), "--",
				"docker", "load",
				"-i", h.tarfilePathInVM(key),
			)
			if err := h.commander.Run(ctx, loadCommand...); err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to load %s", images[key]))
			}
		}
	}

	indexCommand := flatten(
		"docker", "run", "--rm",
		"--cpus", strconv.Itoa(h.options.FirecrackerNumCPUs),
		"--memory", h.options.FirecrackerMemory,
		"-v", fmt.Sprintf("%s:/data", mountPoint),
		"-w", "/data",
		"sourcegraph/lsif-go:latest",
		"lsif-go",
		"--no-animation",
	)
	if h.options.UseFirecracker {
		indexCommand = flatten(
			"ignite", "exec", name.String(), "--",
			indexCommand,
		)
	}
	if err := h.commander.Run(ctx, indexCommand...); err != nil {
		return errors.Wrap(err, "failed to index repository")
	}

	uploadCommand := flatten(
		"docker", "run", "--rm",
		"--cpus", strconv.Itoa(h.options.FirecrackerNumCPUs),
		"--memory", h.options.FirecrackerMemory,
		"-v", fmt.Sprintf("%s:/data", mountPoint),
		"-w", "/data",
		"-e", fmt.Sprintf("SRC_ENDPOINT=%s", uploadURL.String()),
		"sourcegraph/src-cli:latest",
		"lsif", "upload",
		"-no-progress",
		"-repo", index.RepositoryName,
		"-commit", index.Commit,
		"-upload-route", "/.internal-code-intel/lsif/upload",
	)
	if h.options.UseFirecracker {
		uploadCommand = flatten(
			"ignite", "exec", name.String(), "--",
			uploadCommand,
		)
	}
	if err := h.commander.Run(ctx, uploadCommand...); err != nil {
		return errors.Wrap(err, "failed to upload index")
	}

	return nil
}

// makeTempDir is a wrapper around ioutil.TempDir that can be replaced during unit tests.
var makeTempDir = func() (string, error) {
	// TMPDIR is set in the dev Procfile to avoid requiring developers to explicitly
	// allow bind mounts of the host's /tmp. If this directory doesn't exist, ioutil.TempDir
	// below will fail.
	if tmpdir := os.Getenv("TMPDIR"); tmpdir != "" {
		if err := os.MkdirAll(tmpdir, os.ModePerm); err != nil {
			return "", err
		}
	}

	return ioutil.TempDir("", "")
}

// fetchRepository creates a temporary directory and performs a git checkout with the given repository
// and commit. If there is an error, the temporary directory is removed.
func (h *Handler) fetchRepository(ctx context.Context, repositoryName, commit string) (string, error) {
	tempDir, err := makeTempDir()
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
		}
	}()

	cloneURL, err := makeCloneURL(h.options.FrontendURL, h.options.AuthToken, repositoryName)
	if err != nil {
		return "", err
	}

	gitCommands := [][]string{
		{"git", "-C", tempDir, "init"},
		{"git", "-C", tempDir, "-c", "protocol.version=2", "fetch", cloneURL.String(), commit},
		{"git", "-C", tempDir, "checkout", commit},
	}
	for _, gitCommand := range gitCommands {
		if err := h.commander.Run(ctx, gitCommand...); err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed `git %s`", strings.Join(gitCommand, " ")))
		}
	}

	return tempDir, nil
}

func (h *Handler) saveDockerImage(ctx context.Context, key, image string) error {
	pullCommand := flatten(
		"docker", "pull",
		image,
	)
	if err := h.commander.Run(ctx, pullCommand...); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to pull %s", image))
	}

	saveCommand := flatten(
		"docker", "save",
		"-o", h.tarfilePathOnHost(key),
		image,
	)
	if err := h.commander.Run(ctx, saveCommand...); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to save %s", image))
	}

	return nil
}

func (h *Handler) tarfilePathOnHost(key string) string {
	return filepath.Join(h.options.ImageArchivePath, fmt.Sprintf("%s.tar", key))
}

func (h *Handler) tarfilePathInVM(key string) string {
	return fmt.Sprintf("/%s.tar", key)
}

func makeCloneURL(baseURL, authToken, repositoryName string) (*url.URL, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	base.User = url.UserPassword("indexer", authToken)

	return base.ResolveReference(&url.URL{Path: path.Join(".internal-code-intel", "git", repositoryName)}), nil
}

func makeUploadURL(baseURL, authToken string) (*url.URL, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	base.User = url.UserPassword("indexer", authToken)

	return base, nil
}

func flatten(values ...interface{}) (union []string) {
	for _, value := range values {
		switch v := value.(type) {
		case string:
			union = append(union, v)
		case []string:
			union = append(union, v...)
		}
	}

	return union
}

func orderedKeys(m map[string]string) (keys []string) {
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
