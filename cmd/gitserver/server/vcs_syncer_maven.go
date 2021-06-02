package server

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/maven/coursier"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
	"github.com/sourcegraph/sourcegraph/schema"
)

type MavenArtifactSyncer struct {
	Config *schema.MavenConnection
}

var _ VCSSyncer = &MavenArtifactSyncer{}

func (s MavenArtifactSyncer) Type() string {
	return "maven"
}

// IsCloneable checks to see if the given URL points to a Java Maven package
// that has available sources. Any non-nil error indicates the sources do not
// exist for this package.
func (s MavenArtifactSyncer) IsCloneable(ctx context.Context, remoteURL *vcs.URL) error {
	dependency := reposource.DecomposeMavenPath(remoteURL.Path)
	sources, err := coursier.FetchSources(ctx, s.Config, dependency)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return errors.Errorf("no sources.jar for dependency %s", dependency)
	}
	return nil
}

// CloneCommand returns a dummy command that has no side-effect. The actual cloning happens
// inside this method because there is no separate command-line tool that performs all the steps
// to proxy the sources of a Maven package.
func (s MavenArtifactSyncer) CloneCommand(ctx context.Context, remoteURL *vcs.URL, bareGitDir string) (*exec.Cmd, error) {
	// Create a temporary directory where we initialize a non-bare git repo,
	// unzip all the sources and create a git commit. This directory gets
	// removed after we have pushed the contents to the bare git directory.
	tmpDir, err := ioutil.TempDir("", "maven")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	dependency := reposource.DecomposeMavenPath(remoteURL.Path)

	paths, err := coursier.FetchSources(ctx, s.Config, dependency)
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return nil, errors.Errorf("no sources.jar for dependency %s", dependency)
	}

	path := paths[0]

	initCmd := exec.CommandContext(ctx, "git", "init", "--initial-branch=main")
	initCmd.Dir = tmpDir
	if output, err := runWith(ctx, initCmd, false, nil); err != nil {
		return nil, errors.Wrapf(err, "command %s failed with output %q", initCmd.Args, string(output))
	}

	err = s.commitJar(ctx, tmpDir, dependency, path)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(bareGitDir, 0755)
	if err != nil {
		return nil, err
	}

	initBareCmd := exec.CommandContext(ctx, "git", "--bare", "init")
	initBareCmd.Dir = bareGitDir
	if output, err := runWith(ctx, initBareCmd, false, nil); err != nil {
		return nil, errors.Wrapf(err, "command %s failed with output %q", initBareCmd.Args, string(output))
	}

	remoteAddCmd := exec.CommandContext(ctx, "git", "remote", "add", "origin", bareGitDir)
	remoteAddCmd.Dir = tmpDir
	if output, err := runWith(ctx, remoteAddCmd, false, nil); err != nil {
		return nil, errors.Wrapf(err, "command %s failed with output %q", remoteAddCmd.Args, string(output))
	}

	gitPushCmd := exec.CommandContext(ctx, "git", "push", "origin", "main")
	gitPushCmd.Dir = tmpDir
	if output, err := runWith(ctx, gitPushCmd, false, nil); err != nil {
		return nil, errors.Wrapf(err, "command %s failed with output %q", gitPushCmd.Args, string(output))
	}

	return exec.CommandContext(ctx, "git", "--version"), nil
}

// Fetch does nothing for Maven packages because they are immutable and cannot be updated after publishing.
func (s MavenArtifactSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, dir GitDir) error {
	return nil
}

// RemoteShowCommand returns the command to be executed for showing remote.
func (s MavenArtifactSyncer) RemoteShowCommand(ctx context.Context, remoteURL *vcs.URL) (cmd *exec.Cmd, err error) {
	return exec.CommandContext(ctx, "git", "remote", "show", "./"), nil
}

func (s MavenArtifactSyncer) commitJar(ctx context.Context, workingDirectory, dependency, path string) error {
	cmd := exec.CommandContext(ctx, "unzip", path, "-d", "./")
	cmd.Dir = workingDirectory
	if output, err := runWith(ctx, cmd, false, nil); err != nil {
		return errors.Wrapf(err, "failed to unzip with output %q", string(output))
	}

	file, err := os.Create(filepath.Join(workingDirectory, "lsif-java.json"))
	if err != nil {
		return err
	}
	defer file.Close()

	jsonContents, err := json.Marshal(&lsifJavaJson{
		Kind:         "maven",
		Jvm:          "8",
		Dependencies: []string{dependency},
	})
	if err != nil {
		return err
	}

	_, err = file.Write(jsonContents)
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = workingDirectory
	if output, err := runWith(ctx, cmd, false, nil); err != nil {
		return errors.Wrapf(err, "failed to git add with output %q", string(output))
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", dependency)
	cmd.Dir = workingDirectory
	if output, err := runWith(ctx, cmd, false, nil); err != nil {
		return errors.Wrapf(err, "failed to git commit with output %q", string(output))
	}

	return nil
}

type lsifJavaJson struct {
	Kind         string   `json:"kind"`
	Jvm          string   `json:"jvm"`
	Dependencies []string `json:"dependencies"`
}
