package campaigns

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

func runSteps(ctx context.Context, client api.Client, repo *graphql.Repository, steps []Step, logger *TaskLogger, tempDir string) ([]byte, error) {
	zipFile, err := fetchRepositoryArchive(ctx, client, repo, tempDir)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching ZIP archive failed")
	}
	defer os.Remove(zipFile.Name())

	prefix := "changeset-" + repo.Slug()
	volumeDir, err := unzipToTempDir(ctx, zipFile.Name(), tempDir, prefix)
	if err != nil {
		return nil, errors.Wrap(err, "Unzipping the ZIP archive failed")
	}
	defer os.RemoveAll(volumeDir)

	runGitCmd := func(args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = volumeDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, errors.Wrapf(err, "'git %s' failed: %s", strings.Join(args, " "), out)
		}
		return out, nil
	}

	if _, err := runGitCmd("init"); err != nil {
		return nil, errors.Wrap(err, "git init failed")
	}

	// Set user.name and user.email in the local repository. The user name and
	// e-mail will eventually be ignored anyway, since we're just using the Git
	// repository to generate diffs, but we don't want git to generate alarming
	// looking warnings.
	if _, err := runGitCmd("config", "--local", "user.name", "Sourcegraph"); err != nil {
		return nil, errors.Wrap(err, "git config user.name failed")
	}
	if _, err := runGitCmd("config", "--local", "user.email", "campaigns@sourcegraph.com"); err != nil {
		return nil, errors.Wrap(err, "git config user.email failed")
	}

	// --force because we want previously "gitignored" files in the repository
	if _, err := runGitCmd("add", "--force", "--all"); err != nil {
		return nil, errors.Wrap(err, "git add failed")
	}
	if _, err := runGitCmd("commit", "--quiet", "--all", "-m", "src-action-exec"); err != nil {
		return nil, errors.Wrap(err, "git commit failed")
	}

	for i, step := range steps {
		logger.Logf("[Step %d] docker run %s", i+1, step.image)

		cidFile, err := ioutil.TempFile(tempDir, prefix+"-container-id")
		if err != nil {
			return nil, errors.Wrap(err, "Creating a CID file failed")
		}
		_ = os.Remove(cidFile.Name()) // docker exits if this file exists upon `docker run` starting
		defer func() {
			cid, err := ioutil.ReadFile(cidFile.Name())
			_ = os.Remove(cidFile.Name())
			if err == nil {
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()
				_ = exec.CommandContext(ctx, "docker", "rm", "-f", "--", string(cid)).Run()
			}
		}()

		// For now, we only support shell scripts provided via the Run field.
		shell, containerTemp, err := probeImageForShell(ctx, step.image)
		if err != nil {
			return nil, errors.Wrapf(err, "probing image %q for shell", step.image)
		}

		// Set up a temporary file on the host filesystem to contain the
		// script.
		fp, err := ioutil.TempFile(tempDir, "")
		if err != nil {
			return nil, errors.Wrap(err, "creating temporary file")
		}
		hostTemp := fp.Name()
		defer os.Remove(hostTemp)
		if _, err := fp.WriteString(step.Run); err != nil {
			return nil, errors.Wrapf(err, "writing to temporary file %q", hostTemp)
		}
		fp.Close()

		const workDir = "/work"
		cmd := exec.CommandContext(ctx, "docker", "run",
			"--rm",
			"--cidfile", cidFile.Name(),
			"--workdir", workDir,
			"--mount", fmt.Sprintf("type=bind,source=%s,target=%s", volumeDir, workDir),
			"--mount", fmt.Sprintf("type=bind,source=%s,target=%s,ro", hostTemp, containerTemp),
			"--entrypoint", shell,
		)
		for k, v := range step.Env {
			cmd.Args = append(cmd.Args, "-e", k+"="+v)
		}
		cmd.Args = append(cmd.Args, "--", step.image, containerTemp)
		cmd.Dir = volumeDir
		cmd.Stdout = logger.PrefixWriter("stdout")
		cmd.Stderr = logger.PrefixWriter("stderr")

		a, err := json.Marshal(cmd.Args)
		if err != nil {
			panic(err)
		}
		logger.Log(string(a))

		t0 := time.Now()
		err = cmd.Run()
		elapsed := time.Since(t0).Round(time.Millisecond)
		if err != nil {
			logger.Logf("[Step %d] took %s; error running Docker container: %+v", i+1, elapsed, err)
			return nil, errors.Wrapf(err, "Running Docker container for image %q failed", step.image)
		}
		logger.Logf("[Step %d] complete in %s", i+1, elapsed)

	}

	if _, err := runGitCmd("add", "--all"); err != nil {
		return nil, errors.Wrap(err, "git add failed")
	}

	// As of Sourcegraph 3.14 we only support unified diff format.
	// That means we need to strip away the `a/` and `/b` prefixes with `--no-prefix`.
	// See: https://github.com/sourcegraph/sourcegraph/blob/82d5e7e1562fef6be5c0b17f18631040fd330835/enterprise/internal/campaigns/service.go#L324-L329
	//
	// Also, we need to add --binary so binary file changes are inlined in the patch.
	//
	diffOut, err := runGitCmd("diff", "--cached", "--no-prefix", "--binary")
	if err != nil {
		return nil, errors.Wrap(err, "git diff failed")
	}

	return diffOut, err
}

func unzipToTempDir(ctx context.Context, zipFile, tempDir, tempFilePrefix string) (string, error) {
	volumeDir, err := ioutil.TempDir(tempDir, tempFilePrefix)
	if err != nil {
		return "", err
	}
	return volumeDir, unzip(zipFile, volumeDir)
}

func fetchRepositoryArchive(ctx context.Context, client api.Client, repo *graphql.Repository, tempDir string) (*os.File, error) {
	req, err := client.NewHTTPRequest(ctx, "GET", repositoryZipArchivePath(repo), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/zip")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to fetch archive (HTTP %d from %s)", resp.StatusCode, req.URL.String())
	}

	f, err := ioutil.TempFile(tempDir, strings.Replace(repo.Name, "/", "-", -1)+".zip")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return nil, err
	}
	return f, nil
}

func repositoryZipArchivePath(repo *graphql.Repository) string {
	return path.Join("", repo.Name+"@"+repo.DefaultBranch.Name, "-", "raw")
}

func unzip(zipFile, dest string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	outputBase := filepath.Clean(dest) + string(os.PathSeparator)

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: https://snyk.io/research/zip-slip-vulnerability#go
		if !strings.HasPrefix(fpath, outputBase) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
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

func probeImageForShell(ctx context.Context, image string) (shell, tempfile string, err error) {
	// We need to know two things to be able to run a shell script:
	//
	// 1. Which shell is available. We're going to look for /bin/bash and then
	//    /bin/sh, in that order. (Sorry, tcsh users.)
	// 2. Where to put the shell script in the container so that we don't
	//    clobber any actual user data.
	//
	// We can do these together: although it's not part of POSIX proper, every
	// *nix made in the last decade or more has mktemp(1) available. We know
	// that mktemp will give us a file name that doesn't exist in the image if
	// we run it as part of the command. We can also probe for the shell at the
	// same time by trying to run /bin/bash -c mktemp,
	// followed by /bin/sh -c mktemp.

	// We'll also set up our error.
	err = new(multierror.Error)

	// Now we can iterate through our shell options and try to run mktemp with
	// them.
	for _, shell = range []string{"/bin/bash", "/bin/sh"} {
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)

		args := []string{"run", "--rm", "--entrypoint", shell, image, "-c", "mktemp"}

		cmd := exec.CommandContext(ctx, "docker", args...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		if runErr := cmd.Run(); runErr != nil {
			err = multierror.Append(err, errors.Wrapf(runErr, "probing shell %q:\n%s", shell, stderr.String()))
		} else {
			// Even if there were previous errors, we can now ignore them.
			err = nil
			tempfile = strings.TrimSpace(stdout.String())
			return
		}
	}

	// If we got here, then all the attempts to probe the shell failed. Let's
	// admit defeat and return. At least err is already in place.
	return
}
