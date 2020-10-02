package indexer

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
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

	runner, err := h.createRunner(repoDir, index)
	if err != nil {
		return err
	}

	if err := runner.Startup(ctx); err != nil {
		return err
	}
	defer func() {
		if teardownErr := runner.Teardown(ctx); teardownErr != nil {
			err = multierror.Append(err, teardownErr)
		}
	}()

	for _, dockerStep := range index.DockerSteps {
		if err := h.runDockerStep(ctx, runner, index, dockerStep); err != nil {
			return err
		}
	}

	if err := h.index(ctx, runner, index); err != nil {
		return err
	}
	if err := h.upload(ctx, runner, index); err != nil {
		return err
	}

	return nil
}

func (h *Handler) createRunner(repoDir string, index store.Index) (Runner, error) {
	if h.options.UseFirecracker {
		name, err := h.uuidGenerator()
		if err != nil {
			return nil, err
		}

		return NewFirecrackerRunner(h.commander, h.options, repoDir, name.String(), index), nil
	}

	return NewDockerRunner(h.commander, h.options, repoDir, index), nil
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

	commands := [][]string{
		{"git", "-C", tempDir, "init"},
		{"git", "-C", tempDir, "-c", "protocol.version=2", "fetch", cloneURL.String(), commit},
		{"git", "-C", tempDir, "checkout", commit},
	}

	for _, args := range commands {
		if err := h.commander.Run(ctx, args...); err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed `%s`", strings.Join(args, " ")))
		}
	}

	return tempDir, nil
}

func (h *Handler) runDockerStep(ctx context.Context, runner Runner, index store.Index, dockerStep store.DockerStep) error {
	// TODO - root
	if err := runner.Invoke(ctx, dockerStep.Image, FromArgs(dockerStep.Commands)); err != nil {
		return err
	}

	return nil
}

func (h *Handler) index(ctx context.Context, runner Runner, index store.Index) error {
	// TODO - root
	if err := runner.Invoke(ctx, index.Indexer, FromArgs(index.IndexerArgs)); err != nil {
		return err
	}

	return nil
}

const uploadImage = "sourcegraph/src-cli:latest"

func (h *Handler) upload(ctx context.Context, runner Runner, index store.Index) error {
	uploadURL, err := makeUploadURL(h.options.FrontendURLFromDocker, h.options.AuthToken)
	if err != nil {
		return err
	}

	cs := FromArgs([]string{
		"lsif", "upload",
		"-no-progress",
		"-repo", index.RepositoryName,
		"-commit", index.Commit,
		"-upload-route", "/.internal-code-intel/lsif/upload",
	}).AddEnv("SRC_ENDPOINT", uploadURL.String())

	if err := runner.Invoke(ctx, uploadImage, cs); err != nil {
		return err
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
