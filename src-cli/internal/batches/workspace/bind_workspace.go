package workspace

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/repozip"
	"github.com/sourcegraph/src-cli/internal/batches/util"
)

type dockerBindWorkspaceCreator struct {
	Dir string
}

var _ Creator = &dockerBindWorkspaceCreator{}

func (wc *dockerBindWorkspaceCreator) Create(ctx context.Context, repo *graphql.Repository, steps []batcheslib.Step, archive repozip.Archive) (Workspace, error) {
	w, err := wc.unzipToWorkspace(ctx, repo, archive.Path())
	if err != nil {
		return nil, errors.Wrap(err, "unzipping the repository")
	}

	if err := wc.copyToWorkspace(ctx, w, archive.AdditionalFilePaths()); err != nil {
		return nil, errors.Wrap(err, "copying additional files into workspace")
	}

	return w, errors.Wrap(wc.prepareGitRepo(ctx, w), "preparing local git repo")
}

func (*dockerBindWorkspaceCreator) prepareGitRepo(ctx context.Context, w *dockerBindWorkspace) error {
	if _, err := runGitCmd(ctx, w.dir, "init"); err != nil {
		return errors.Wrap(err, "git init failed")
	}

	// --force because we want previously "gitignored" files in the repository
	if _, err := runGitCmd(ctx, w.dir, "add", "--force", "--all"); err != nil {
		return errors.Wrap(err, "git add failed")
	}
	if _, err := runGitCmd(ctx, w.dir, "commit", "--quiet", "--all", "--allow-empty", "-m", "src-action-exec"); err != nil {
		return errors.Wrap(err, "git commit failed")
	}

	return nil
}

func (wc *dockerBindWorkspaceCreator) unzipToWorkspace(ctx context.Context, repo *graphql.Repository, zip string) (*dockerBindWorkspace, error) {
	prefix := "workspace-" + util.SlugForRepo(repo.Name, repo.Rev())
	workspace, err := unzipToTempDir(ctx, zip, wc.Dir, prefix)
	if err != nil {
		return nil, errors.Wrap(err, "unzipping the ZIP archive")
	}

	return &dockerBindWorkspace{tempDir: wc.Dir, dir: workspace}, nil
}

func (wc *dockerBindWorkspaceCreator) copyToWorkspace(_ context.Context, w *dockerBindWorkspace, files map[string]string) error {
	for name, src := range files {
		srcStat, err := os.Stat(src)
		if err != nil {
			return err
		}

		if !srcStat.Mode().IsRegular() {
			return errors.Newf("%s is not a regular file", src)
		}

		srcFile, err := os.Open(src)
		if err != nil {
			return err
		}

		destPath := path.Join(w.dir, name)

		destFile, err := prepareCopyDestinationFile(srcStat, destPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return err
		}

		if cerr := destFile.Close(); cerr != nil {
			return errors.Wrap(cerr, "closing destination file failed")
		}
		if cerr := srcFile.Close(); cerr != nil {
			return errors.Wrap(cerr, "closing source file failed")
		}
	}

	return nil
}

// dockerBindWorkspace implements a workspace that operates on the host FS
// and is mounted into the docker containers using a bind mount in the end.
type dockerBindWorkspace struct {
	// tempDir is a temporary directory that will be used for ephemeral files
	// that are needed throughout the process.
	tempDir string
	// dir is the directory where the repo archive is unzipped to.
	// This is also the path that is directly mounted into the docker
	// containers.
	dir string
}

var _ Workspace = &dockerBindWorkspace{}

func (w *dockerBindWorkspace) Close(ctx context.Context) error {
	return os.RemoveAll(w.dir)
}

func (w *dockerBindWorkspace) DockerRunOpts(ctx context.Context, target string) ([]string, error) {
	return []string{
		"--mount",
		fmt.Sprintf("type=bind,source=%s,target=%s", w.dir, target),
	}, nil
}

func (w *dockerBindWorkspace) WorkDir() *string { return &w.dir }

func (w *dockerBindWorkspace) Diff(ctx context.Context) ([]byte, error) {
	if _, err := runGitCmd(ctx, w.dir, "add", "--all"); err != nil {
		return nil, errors.Wrap(err, "git add failed")
	}

	// As of Sourcegraph 3.14 we only support unified diff format.
	// That means we need to strip away the `a/` and `/b` prefixes with `--no-prefix`.
	// See: https://github.com/sourcegraph/sourcegraph/blob/82d5e7e1562fef6be5c0b17f18631040fd330835/enterprise/internal/campaigns/service.go#L324-L329
	//
	// Also, we need to add --binary so binary file changes are inlined in the patch.
	//
	// ATTENTION: When you change the options here, be sure to also update the
	// ApplyDiff method accordingly.
	return runGitCmd(ctx, w.dir, "diff", "--cached", "--no-prefix", "--binary")
}

func (w *dockerBindWorkspace) ApplyDiff(ctx context.Context, diff []byte) error {
	// Write the diff to a temp file so we can pass it to `git apply`
	tmp, err := os.CreateTemp(w.tempDir, "bind-workspace-test-*")
	if err != nil {
		return errors.Wrap(err, "saving cached diff to temporary file")
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(diff); err != nil {
		return errors.Wrap(err, "writing to temporary file")
	}

	if err := tmp.Close(); err != nil {
		return errors.Wrap(err, "closing temporary file")
	}

	// Apply diff
	if out, err := runGitCmd(ctx, w.dir, "apply", "-p0", tmp.Name()); err != nil {
		return errors.Wrapf(err, "applying cached diff: %s", string(out))
	}

	// Add all files to index
	_, err = runGitCmd(ctx, w.dir, "add", "--all")
	return err
}

func unzipToTempDir(ctx context.Context, zipFile, tempDir, tempFilePrefix string) (string, error) {
	volumeDir, err := os.MkdirTemp(tempDir, tempFilePrefix)
	if err != nil {
		return "", err
	}

	if err := os.Chmod(volumeDir, 0777); err != nil {
		return "", err
	}

	return volumeDir, unzip(ctx, zipFile, volumeDir)
}

func unzip(ctx context.Context, zipFile, dest string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	outputBase := filepath.Clean(dest) + string(os.PathSeparator)

	for _, f := range r.File {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: https://snyk.io/research/zip-slip-vulnerability#go
		if !strings.HasPrefix(fpath, outputBase) {
			return errors.Newf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			if err := mkdirAll(dest, f.Name, 0777); err != nil {
				return err
			}
			continue
		}

		outFile, err := prepareCopyDestinationFile(f.FileInfo(), fpath)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		cerr := outFile.Close()
		// Now we have safely closed everything that needs it, and can check errors
		if err != nil {
			return errors.Wrapf(err, "copying %q failed", f.Name)
		}
		if cerr != nil {
			return errors.Wrap(err, "closing output file failed")
		}

	}

	return nil
}

type moder interface {
	Mode() os.FileMode
}

func prepareCopyDestinationFile(sourceInfo moder, dest string) (*os.File, error) {
	if err := mkdirAll(filepath.Dir(dest), "", 0777); err != nil {
		return nil, err
	}

	outFile, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return nil, errors.Wrap(err, "opening destination file failed")
	}

	// Since the container might not run as the same user, we need to ensure
	// that the file is globally writable. If the execute bit is normally
	// set on the zipped up file, let's ensure we propagate that to the
	// group and other permission bits too.
	if sourceInfo.Mode()&0111 != 0 {
		err = os.Chmod(outFile.Name(), 0777)
	} else {
		err = os.Chmod(outFile.Name(), 0666)
	}
	if err != nil {
		outFile.Close()
		return nil, err
	}

	return outFile, nil
}

// Technically, this is a misnomer, since it might be a socket or block special,
// but errPathExistsAsNonDir is just ugly for an internal type.
type errPathExistsAsFile string

var _ error = errPathExistsAsFile("")

func (e errPathExistsAsFile) Error() string {
	return fmt.Sprintf("path already exists, but not as a directory: %s", string(e))
}

// mkdirAll is essentially os.MkdirAll(filepath.Join(base, path), perm), but
// applies the given permission regardless of the user's umask.
func mkdirAll(base, path string, perm os.FileMode) error {
	abs := filepath.Join(base, path)

	// Create the directory if it doesn't exist.
	st, err := os.Stat(abs)
	if err != nil {
		// It's expected that we'll get an error if the directory doesn't exist,
		// so let's check that it's of the type we expect.
		if !os.IsNotExist(err) {
			return err
		}

		// Now we're clear to create the directory.
		if err := os.MkdirAll(abs, perm); err != nil {
			return err
		}
	} else if !st.IsDir() {
		// The file/socket/whatever exists, but it's not a directory. That's
		// definitely going to be an issue.
		return errPathExistsAsFile(abs)
	}

	// If os.MkdirAll() was invoked earlier, then the permissions it set were
	// subject to the umask. Let's walk the directories we may or may not have
	// created and ensure their permissions look how we want.
	return ensureAll(base, path, perm)
}

// ensureAll ensures that all directories under path have the expected
// permissions.
func ensureAll(base, path string, perm os.FileMode) error {
	var errs errors.MultiError

	// In plain English: for each directory in the path parameter, we should
	// chmod that path to the permissions that are expected.
	acc := []string{base}
	for _, element := range strings.Split(path, string(os.PathSeparator)) {
		acc = append(acc, element)
		if err := os.Chmod(filepath.Join(acc...), perm); err != nil {
			errs = errors.Append(errs, err)
		}
	}

	return errs
}
