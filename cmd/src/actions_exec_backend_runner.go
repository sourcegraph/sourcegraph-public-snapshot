package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"golang.org/x/net/context/ctxhttp"
)

type ActionRepoStatus struct {
	Cached bool

	LogFile    string
	EnqueuedAt time.Time
	StartedAt  time.Time
	FinishedAt time.Time

	Patch CampaignPlanPatch
	Err   error
}

func (x *actionExecutor) do(ctx context.Context, repo ActionRepo) (err error) {
	// Check if cached.
	cacheKey := actionExecutionCacheKey{Repo: repo, Runs: x.action.Steps}
	if result, ok, err := x.opt.cache.get(ctx, cacheKey); err != nil {
		return errors.Wrapf(err, "checking cache for %s", repo.Name)
	} else if ok {
		x.updateRepoStatus(repo, ActionRepoStatus{
			Cached: true,
			Patch:  result,
		})
		return nil
	}

	prefix := "action-" + strings.Replace(strings.Replace(repo.Name, "/", "-", -1), "github.com-", "", -1)

	// TODO(sqs): better cleanup of old log files
	logFile, err := ioutil.TempFile(tempDirPrefix, prefix+"-log")
	if err != nil {
		return err
	}
	if !x.opt.keepLogs {
		defer func() {
			if err == nil {
				os.Remove(logFile.Name())
			}
		}()
	}
	logWriter := io.Writer(logFile)
	if *verbose {
		logWriter = io.MultiWriter(logWriter, os.Stderr)
	}

	x.updateRepoStatus(repo, ActionRepoStatus{
		LogFile:   logFile.Name(),
		StartedAt: time.Now(),
	})

	runCtx, cancel := context.WithTimeout(ctx, x.opt.timeout)
	defer cancel()

	patch, err := runAction(runCtx, prefix, repo.ID, repo.Name, repo.Rev, x.action.Steps, logWriter)
	status := ActionRepoStatus{
		FinishedAt: time.Now(),
	}
	if len(patch) > 0 {
		status.Patch = CampaignPlanPatch{
			Repository:   repo.ID,
			BaseRevision: repo.Rev,
			Patch:        string(patch),
		}
	}
	if err != nil {
		if errors.Cause(err) == context.DeadlineExceeded {
			err = &errTimeoutReached{timeout: x.opt.timeout}
		}
		status.Err = err
		fmt.Fprintf(logWriter, "# ERROR: %s\n", err)
	}
	x.updateRepoStatus(repo, status)

	// Add to cache if successful.
	if err == nil {
		// We don't use runCtx here because we want to write to the cache even
		// if we've now reached the timeout
		if err := x.opt.cache.set(ctx, cacheKey, status.Patch); err != nil {
			return errors.Wrapf(err, "caching result for %s", repo.Name)
		}
	}

	return err
}

type errTimeoutReached struct{ timeout time.Duration }

func (e *errTimeoutReached) Error() string {
	return fmt.Sprintf("Timeout reached. Execution took longer than %s.", e.timeout)
}

func runAction(ctx context.Context, prefix, repoID, repoName, rev string, steps []*ActionStep, logFile io.Writer) ([]byte, error) {
	fmt.Fprintf(logFile, "# Repository %s @ %s (%d steps)\n", repoName, rev, len(steps))

	zipFile, err := fetchRepositoryArchive(ctx, repoName, rev)
	if err != nil {
		return nil, errors.Wrap(err, "Fetching ZIP archive failed")
	}
	defer os.Remove(zipFile.Name())

	volumeDir, err := unzipToTempDir(ctx, zipFile.Name(), prefix)
	if err != nil {
		return nil, errors.Wrap(err, "Unzipping the ZIP archive failed")
	}
	defer os.RemoveAll(volumeDir)

	for i, step := range steps {
		if i != 0 {
			fmt.Fprintln(logFile)
		}

		logPrefix := fmt.Sprintf("Step %d", i)

		switch step.Type {
		case "command":
			fmt.Fprintf(logFile, "# %s: command %v\n", logPrefix, step.Args)

			cmd := exec.CommandContext(ctx, step.Args[0], step.Args[1:]...)
			cmd.Dir = volumeDir
			cmd.Stdout = logFile
			cmd.Stderr = logFile
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(logFile, "# %s: error: %s.\n", logPrefix, err)
				return nil, errors.Wrap(err, "run command")
			}
			fmt.Fprintf(logFile, "# %s: done.\n", logPrefix)

		case "docker":
			var fromDockerfile string
			if step.Dockerfile != "" {
				fromDockerfile = " (built from inline Dockerfile)"
			}
			fmt.Fprintf(logFile, "# %s: docker run %v%s\n", logPrefix, step.Image, fromDockerfile)

			cidFile, err := ioutil.TempFile(tempDirPrefix, prefix+"-container-id")
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

			const workDir = "/work"
			cmd := exec.CommandContext(ctx, "docker", "run",
				"--rm",
				"--cidfile", cidFile.Name(),
				"--workdir", workDir,
				"--mount", fmt.Sprintf("type=bind,source=%s,target=%s", volumeDir, workDir),
			)
			for _, cacheDir := range step.CacheDirs {
				// persistentCacheDir returns a host directory that persists across runs of this
				// action for this repository. It is useful for (e.g.) yarn and npm caches.
				persistentCacheDir := func(containerDir string) (string, error) {
					baseCacheDir, err := userCacheDir()
					if err != nil {
						return "", err
					}
					b := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", step.Image, repoName, rev)))
					return filepath.Join(baseCacheDir, "action-exec-cache-dir",
						base64.RawURLEncoding.EncodeToString(b[:16]),
						strings.TrimPrefix(cacheDir, string(os.PathSeparator))), nil
				}

				hostDir, err := persistentCacheDir(cacheDir)
				if err != nil {
					return nil, err
				}
				if err := os.MkdirAll(hostDir, 0700); err != nil {
					return nil, err
				}
				cmd.Args = append(cmd.Args, "--mount", fmt.Sprintf("type=bind,source=%s,target=%s", hostDir, cacheDir))
			}
			cmd.Args = append(cmd.Args, "--", step.Image)
			cmd.Args = append(cmd.Args, step.Args...)
			cmd.Dir = volumeDir
			cmd.Stdout = logFile
			cmd.Stderr = logFile
			t0 := time.Now()
			err = cmd.Run()
			elapsed := time.Since(t0).Round(time.Millisecond)
			if err != nil {
				fmt.Fprintf(logFile, "# %s: error: %s. (%s)\n", logPrefix, err, elapsed)
				return nil, errors.Wrap(err, "run docker container")
			}
			fmt.Fprintf(logFile, "# %s: done. (%s)\n", logPrefix, elapsed)

		default:
			return nil, fmt.Errorf("unrecognized run type %q", step.Type)
		}
	}

	// Compute diff.
	oldDir, err := unzipToTempDir(ctx, zipFile.Name(), prefix)
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(oldDir)

	diffOut, err := diffDirs(ctx, oldDir, volumeDir)
	if err != nil {
		return nil, errors.Wrap(err, "Generating a diff failed")
	}

	// Strip temp dir prefixes from diff.
	fileDiffs, err := diff.ParseMultiFileDiff(diffOut)
	if err != nil {
		return nil, err
	}
	for _, fileDiff := range fileDiffs {
		for i := range fileDiff.Extended {
			fileDiff.Extended[i] = strings.Replace(fileDiff.Extended[i], oldDir+string(os.PathSeparator), "", -1)
			fileDiff.Extended[i] = strings.Replace(fileDiff.Extended[i], volumeDir+string(os.PathSeparator), "", -1)
		}
		fileDiff.OrigName = strings.TrimPrefix(fileDiff.OrigName, oldDir+string(os.PathSeparator))
		fileDiff.NewName = strings.TrimPrefix(fileDiff.NewName, volumeDir+string(os.PathSeparator))
	}
	return diff.PrintMultiFileDiff(fileDiffs)
}

func diffDirs(ctx context.Context, oldDir, newDir string) ([]byte, error) {
	args := []string{"--unified", "--new-file", "--recursive"}

	if diffSupportsNoDereference {
		args = append(args, "--no-dereference")
	}

	if diffSupportsColor {
		args = append(args, "--color=never")
	}

	args = append(args, oldDir, newDir)
	cmd := exec.CommandContext(ctx, "diff", args...)

	out, err := cmd.CombinedOutput()
	// 1 just means files differ, not error
	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		outputSummary := string(out)
		if max := 250; len(outputSummary) >= max {
			outputSummary = outputSummary[:max] + "..."
		}
		return nil, errors.Wrapf(err, "diff (output was: %q)", outputSummary)
	}

	return out, nil
}

func diffSupportsFlag(ctx context.Context, flag string) (bool, error) {
	cmd := exec.CommandContext(ctx, "diff", flag)
	out, err := cmd.CombinedOutput()
	// diff 2.8.1 returns exit code 2 when printing "unrecognized option" message
	if err != nil && cmd.ProcessState.ExitCode() != 2 {
		return false, errors.Wrapf(err, "Checking whether diff supports %q failed", flag)
	}
	return !strings.Contains(string(out), "unrecognized option `"+flag), nil
}

// We use an explicit prefix for our temp directories, because otherwise Go
// would use $TMPDIR, which is set to `/var/folders` per default on macOS. But
// Docker for Mac doesn't have `/var/folders` in its default set of shared
// folders, but it does have `/tmp` in there.
const tempDirPrefix = "/tmp"

func unzipToTempDir(ctx context.Context, zipFile, prefix string) (string, error) {
	volumeDir, err := ioutil.TempDir(tempDirPrefix, prefix)
	if err != nil {
		return "", err
	}
	unzipCmd := exec.CommandContext(ctx, "unzip", "-qq", zipFile, "-d", volumeDir)
	if out, err := unzipCmd.CombinedOutput(); err != nil {
		os.RemoveAll(volumeDir)
		return "", fmt.Errorf("unzip failed: %s: %s", err, out)
	}
	return volumeDir, nil
}

func fetchRepositoryArchive(ctx context.Context, repoName, rev string) (*os.File, error) {
	zipURL, err := repositoryZipArchiveURL(repoName, rev, "")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", zipURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/zip")
	if cfg.AccessToken != "" {
		req.Header.Set("Authorization", "token "+cfg.AccessToken)
	}
	resp, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to fetch archive (HTTP %d from %s)", resp.StatusCode, zipURL)
	}

	f, err := ioutil.TempFile(tempDirPrefix, strings.Replace(repoName, "/", "-", -1)+".zip")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return nil, err
	}
	return f, nil
}

func repositoryZipArchiveURL(repoName, rev, token string) (*url.URL, error) {
	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	if token != "" {
		u.User = url.User(token)
	}
	u.Path = path.Join(u.Path, repoName+"@"+rev, "-", "raw")
	return u, nil
}
