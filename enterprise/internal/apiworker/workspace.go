package apiworker

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/command"
)

// prepareWorkspace creates and returns a temporary director in which acts the workspace
// while processing a single job. It is up to the caller to ensure that this directory is
// removed after the job has finished processing. If a repository name is supplied, then
// that repository will be cloned (through the frontend API) into the workspace.
func (h *handler) prepareWorkspace(ctx context.Context, commandRunner command.Runner, repositoryName, commit string) (_ string, err error) {
	// TMPDIR is set in the dev Procfile to avoid requiring developers to explicitly
	// allow bind mounts of the host's /tmp. If this directory doesn't exist,
	// ioutil.TempDir below will fail.
	if tmpdir := os.Getenv("TMPDIR"); tmpdir != "" {
		if err := os.MkdirAll(tmpdir, os.ModePerm); err != nil {
			return "", err
		}
	}

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
		}
	}()

	if repositoryName == "" {
		return tempDir, nil
	}

	cloneURL, err := makeURL(
		h.options.ClientOptions.EndpointOptions.URL,
		h.options.ClientOptions.EndpointOptions.Username,
		h.options.ClientOptions.EndpointOptions.Password,
		h.options.GitServicePath,
		repositoryName,
	)
	if err != nil {
		return "", err
	}

	gitCommands := []command.CommandSpec{
		{Key: "setup.git.init", Commands: []string{"git", "-C", tempDir, "init"}},
		{Key: "setup.git.fetch", Commands: []string{"git", "-C", tempDir, "-c", "protocol.version=2", "fetch", cloneURL.String(), commit}},
		{Key: "setup.git.checkout", Commands: []string{"git", "-C", tempDir, "checkout", commit}},
	}
	for _, spec := range gitCommands {
		if err := commandRunner.Run(ctx, spec); err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("failed %s", spec.Key))
		}
	}

	return tempDir, nil
}

func makeURL(base, username, password string, path ...string) (*url.URL, error) {
	u, err := makeRelativeURL(base, path...)
	if err != nil {
		return nil, err
	}

	u.User = url.UserPassword(username, password)
	return u, nil
}

func makeRelativeURL(base string, path ...string) (*url.URL, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	return baseURL.ResolveReference(&url.URL{Path: filepath.Join(path...)}), nil
}
