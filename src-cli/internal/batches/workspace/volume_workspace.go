package workspace

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"

	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/repozip"
	"github.com/sourcegraph/src-cli/internal/exec"
	"github.com/sourcegraph/src-cli/internal/version"
)

type imageEnsurer func(ctx context.Context, image string) (docker.Image, error)

type dockerVolumeWorkspaceCreator struct {
	tempDir     string
	EnsureImage imageEnsurer
}

var _ Creator = &dockerVolumeWorkspaceCreator{}

func (wc *dockerVolumeWorkspaceCreator) Create(ctx context.Context, repo *graphql.Repository,
	steps []batcheslib.Step, archive repozip.Archive) (ws Workspace, err error) {
	volume, err := wc.createVolume(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "creating Docker volume")
	}

	defer func() {
		if err != nil {
			deleteVolume(ctx, volume)
		}
	}()

	// Figure out the user that containers will be run as.
	ug := docker.UIDGID{}
	if len(steps) > 0 {
		img, err := wc.EnsureImage(ctx, steps[0].Container)
		if err != nil {
			return nil, err
		}
		if ug, err = img.UIDGID(ctx); err != nil {
			return nil, errors.Wrap(err, "getting container UID and GID")
		}
	}

	w := &dockerVolumeWorkspace{
		tempDir: wc.tempDir,
		volume:  volume,
		uidGid:  ug,
	}
	if err := wc.unzipRepoIntoVolume(ctx, w, archive.Path()); err != nil {
		return nil, errors.Wrap(err, "unzipping repo into workspace")
	}

	if err := wc.copyFilesIntoVolumes(ctx, w, archive.AdditionalFilePaths()); err != nil {
		return nil, errors.Wrap(err, "copying additional files into workspace")
	}

	return w, errors.Wrap(wc.prepareGitRepo(ctx, w), "preparing local git repo")
}

func (*dockerVolumeWorkspaceCreator) createVolume(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "docker", "volume", "create").CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(out)), nil
}

func deleteVolume(ctx context.Context, volume string) error {
	return exec.CommandContext(ctx, "docker", "volume", "rm", volume).Run()
}

func (*dockerVolumeWorkspaceCreator) prepareGitRepo(ctx context.Context, w *dockerVolumeWorkspace) error {
	script := `#!/bin/sh
	
set -e
set -x

git init

# Note that we don't actually use these anywhere, since we're not creating the
# real commits in this container, but we do need _something_ set to avoid Git
# erroring out.
git config user.name 'Sourcegraph Batch Changes'
git config user.email batch-changes@sourcegraph.com

# --force because we want previously "gitignored" files in the repository
git add --force --all
git commit --quiet --all --allow-empty -m src-action-exec
`

	if _, err := w.runScript(ctx, "/work", script); err != nil {
		return errors.Wrap(err, "preparing workspace")
	}
	return nil
}

func (wc *dockerVolumeWorkspaceCreator) unzipRepoIntoVolume(ctx context.Context, w *dockerVolumeWorkspace, zip string) error {
	// We want to mount that temporary file into a Docker container that has the
	// workspace volume attached, and unzip it into the volume.

	// We need to keep a temporary file in the volume before unzipping for the
	// permissions to persist because... reasons. Rather than reading the
	// potentially large ZIP file, we'll cheat a bit and just assume that if we
	// create a file with an appropriately namespaced and random name, it's
	// _probably_ OK. If you manage to reliably trigger an archive that has this
	// file in it, we'll send you a hoodie or something.
	randToken := make([]byte, 16)
	if _, err := rand.Read(randToken); err != nil {
		return errors.Wrap(err, "generating randomness")
	}
	dummy := fmt.Sprintf(".batch-change-workspace-placeholder-%s", hex.EncodeToString(randToken))

	// So, let's use that to set up the volume.
	//
	// Theoretically, we could combine this `docker run` and the following one
	// into one invocation. Doing so, however, is tricky: we'd have to su within
	// the script being run, and Alpine requires a real user account and group;
	// just having numeric IDs is insufficient. The logic to make this work is
	// complicated enough that it feels brittle, and beyond what should be
	// encoded in this function. Running `docker run` twice isn't ideal, but
	// should be quick enough in general that it's not a huge concern.
	opts := append([]string{
		"run",
		"--rm",
		"--init",
		"--workdir", "/work",
	}, w.dockerRunOptsWithUser(docker.Root, "/work")...)
	opts = append(
		opts,
		DockerVolumeWorkspaceImage,
		"sh", "-c",
		fmt.Sprintf("touch /work/%s; chown -R %s /work", dummy, w.uidGid.String()),
	)

	if out, err := exec.CommandContext(ctx, "docker", opts...).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "chown output:\n\n%s\n\n", string(out))
	}

	// Now we can unzip the archive as the user and clean up the temporary file.
	opts = append([]string{
		"run",
		"--rm",
		"--init",
		"--workdir", "/work",
		"--mount", "type=bind,source=" + zip + ",target=/tmp/zip,ro",
	}, w.dockerRunOptsWithUser(w.uidGid, "/work")...)
	opts = append(
		opts,
		DockerVolumeWorkspaceImage,
		"sh", "-c",
		fmt.Sprintf("unzip /tmp/zip; rm /work/%s", dummy),
	)

	if out, err := exec.CommandContext(ctx, "docker", opts...).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "unzip output:\n\n%s\n\n", string(out))
	}

	return nil
}

func (wc *dockerVolumeWorkspaceCreator) copyFilesIntoVolumes(ctx context.Context, w *dockerVolumeWorkspace, files map[string]string) error {
	if len(files) == 0 {
		return nil
	}

	opts := append([]string{
		"run",
		"--rm",
		"--init",
		"--workdir", "/work",
	}, w.dockerRunOptsWithUser(w.uidGid, "/work")...)

	// We sort these so our tests don't break. Sorry.
	var names []string
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	var copyCmds []string
	for _, name := range names {
		localPath := files[name]
		opts = append(opts, []string{
			"--mount", "type=bind,source=" + localPath + ",target=/tmp/" + name + ",ro",
		}...)

		copyCmds = append(copyCmds, "cp /tmp/"+name+" /work/"+name)
	}

	opts = append(
		opts,
		DockerVolumeWorkspaceImage,
		"sh", "-c",
		strings.Join(copyCmds, " && ")+";",
	)

	if out, err := exec.CommandContext(ctx, "docker", opts...).CombinedOutput(); err != nil {
		return errors.Wrapf(err, "unzip output:\n\n%s\n\n", string(out))
	}
	return nil
}

// dockerVolumeWorkspace workspaces are placed on Docker volumes (surprise!),
// and are therefore transparent to the host filesystem. This has performance
// advantages if bind mounts are slow, such as on Docker for Mac, but could make
// debugging harder and is slower when it's time to actually retrieve the diff.
type dockerVolumeWorkspace struct {
	tempDir string
	volume  string
	uidGid  docker.UIDGID
}

var _ Workspace = &dockerVolumeWorkspace{}

func (w *dockerVolumeWorkspace) Close(ctx context.Context) error {
	// Cleanup here is easy: we just get rid of the Docker volume.
	return deleteVolume(ctx, w.volume)
}

func (w *dockerVolumeWorkspace) DockerRunOpts(ctx context.Context, target string) ([]string, error) {
	return w.dockerRunOptsWithUser(w.uidGid, target), nil
}

func (w *dockerVolumeWorkspace) WorkDir() *string { return nil }

func (w *dockerVolumeWorkspace) Diff(ctx context.Context) ([]byte, error) {
	// As of Sourcegraph 3.14 we only support unified diff format.
	// That means we need to strip away the `a/` and `/b` prefixes with `--no-prefix`.
	// See: https://github.com/sourcegraph/sourcegraph/blob/82d5e7e1562fef6be5c0b17f18631040fd330835/enterprise/internal/campaigns/service.go#L324-L329
	//
	// Also, we need to add --binary so binary file changes are inlined in the patch.
	//
	// ATTENTION: When you change the options here, be sure to also update the
	// ApplyDiff method accordingly.
	script := `#!/bin/sh

set -e
# No set -x here, since we're going to parse the git status output.

git add --all > /dev/null
exec git diff --cached --no-prefix --binary
`

	out, err := w.runScript(ctx, "/work", script)
	if err != nil {
		return nil, errors.Wrapf(err, "git diff:\n\n%s", string(out))
	}

	return out, nil
}

func (w *dockerVolumeWorkspace) ApplyDiff(ctx context.Context, diff []byte) error {
	script := fmt.Sprintf(`#!/bin/sh

set -e

cat <<'EOF' | exec git apply -p0 -
%s
EOF

git add --all > /dev/null
`, string(diff))

	out, err := w.runScript(ctx, "/work", script)
	if err != nil {
		return errors.Wrapf(err, "git apply diff:\n\n%s", string(out))
	}

	return nil
}

// DockerVolumeWorkspaceImage is the Docker image we'll run our unzip and git
// commands in. This needs to match the name defined in
// .github/workflows/docker.yml.
var DockerVolumeWorkspaceImage = "sourcegraph/src-batch-change-volume-workspace"

func init() {
	dockerTag := version.BuildTag
	if version.BuildTag == version.DefaultBuildTag {
		dockerTag = "latest"
	}

	DockerVolumeWorkspaceImage = DockerVolumeWorkspaceImage + ":" + dockerTag
}

// runScript is a utility function to mount the given shell script into a Docker
// container started from the dockerWorkspaceImage, then run it and return the
// output.
func (w *dockerVolumeWorkspace) runScript(ctx context.Context, target, script string) ([]byte, error) {
	f, err := os.CreateTemp(w.tempDir, "src-run-*")
	if err != nil {
		return nil, errors.Wrap(err, "creating run script")
	}
	name := f.Name()
	defer os.Remove(name)

	if _, err := f.WriteString(script); err != nil {
		return nil, errors.Wrap(err, "writing run script")
	}
	if err := f.Close(); err != nil {
		return nil, errors.Wrap(err, "closing run script")
	}

	// Sidestep any umask issues on the temporary file by always making it
	// executable by everyone.
	if err := os.Chmod(name, 0755); err != nil {
		return nil, errors.Wrap(err, "chmodding run script")
	}

	common, err := w.DockerRunOpts(ctx, target)
	if err != nil {
		return nil, errors.Wrap(err, "generating run options")
	}

	opts := append([]string{
		"run",
		"--rm",
		"--init",
		"--workdir", target,
		"--mount", "type=bind,source=" + name + ",target=/run.sh,ro",
	}, common...)
	opts = append(opts, DockerVolumeWorkspaceImage, "sh", "/run.sh")

	out, err := exec.CommandContext(ctx, "docker", opts...).CombinedOutput()
	if err != nil {
		return out, errors.Wrapf(err, "Docker output:\n\n%s\n\n", string(out))
	}

	return out, nil
}

func (w *dockerVolumeWorkspace) dockerRunOptsWithUser(ug docker.UIDGID, target string) []string {
	return []string{
		"--user", ug.String(),
		"--mount", "type=volume,source=" + w.volume + ",target=" + target,
	}
}
