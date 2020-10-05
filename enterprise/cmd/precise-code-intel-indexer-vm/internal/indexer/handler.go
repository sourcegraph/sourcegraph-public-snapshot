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

const uploadImage = "sourcegraph/src-cli:latest"
const uploadRoute = "/.internal-code-intel/lsif/upload"

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

	commandFormatter, err := h.makeCommandFormatter(repoDir)
	if err != nil {
		return err
	}

	images := []string{
		uploadImage,
	}
	for _, dockerStep := range index.DockerSteps {
		images = append(images, dockerStep.Image)
	}
	if index.Indexer != "" {
		images = append(images, index.Indexer)
	}

	if err := commandFormatter.Setup(ctx, h.commander, images); err != nil {
		return err
	}
	defer func() {
		if teardownErr := commandFormatter.Teardown(ctx, h.commander); teardownErr != nil {
			err = multierror.Append(err, teardownErr)
		}
	}()

	for _, dockerStep := range index.DockerSteps {
		dockerStepCommand := NewCmd(dockerStep.Image, dockerStep.Commands...).SetWd(dockerStep.Root)

		if err := h.commander.Run(ctx, commandFormatter.FormatCommand(dockerStepCommand)...); err != nil {
			return errors.Wrap(err, "failed to perform docker step")
		}
	}

	if index.Indexer != "" {
		indexCommand := NewCmd(index.Indexer, index.IndexerArgs...).SetWd(index.Root)

		if err := h.commander.Run(ctx, commandFormatter.FormatCommand(indexCommand)...); err != nil {
			return errors.Wrap(err, "failed to index repository")
		}
	}

	uploadURL, err := makeUploadURL(h.options.FrontendURLFromDocker, h.options.AuthToken)
	if err != nil {
		return err
	}

	outfile := "dump.lsif"
	if index.Outfile != "" {
		outfile = index.Outfile
	}

	args := flatten(
		"lsif", "upload",
		"-no-progress",
		"-repo", index.RepositoryName,
		"-commit", index.Commit,
		"-upload-route", uploadRoute,
		"-file", outfile,
	)

	uploadCommand := NewCmd(uploadImage, args...).
		SetWd(index.Root).
		AddEnv("SRC_ENDPOINT", uploadURL.String())

	if err := h.commander.Run(ctx, commandFormatter.FormatCommand(uploadCommand)...); err != nil {
		return errors.Wrap(err, "failed to upload index")
	}

	return nil
}

func (h *Handler) makeCommandFormatter(repoDir string) (CommandFormatter, error) {
	if !h.options.UseFirecracker {
		return NewDockerCommandFormatter(repoDir, h.options), nil
	}

	name, err := h.uuidGenerator()
	if err != nil {
		return nil, err
	}

	return NewFirecrackerCommandFormatter(name.String(), repoDir, h.options), nil
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
