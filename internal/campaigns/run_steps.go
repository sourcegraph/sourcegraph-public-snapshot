package campaigns

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

func runSteps(ctx context.Context, rf RepoFetcher, wc WorkspaceCreator, repo *graphql.Repository, steps []Step, logger *TaskLogger, tempDir string, reportProgress func(string)) ([]byte, error) {
	reportProgress("Downloading archive")
	zip, err := rf.Fetch(ctx, repo)
	if err != nil {
		return nil, errors.Wrap(err, "fetching repo")
	}
	defer zip.Close()

	reportProgress("Initializing workspace")
	workspace, err := wc.Create(ctx, repo, zip.Path())
	if err != nil {
		return nil, errors.Wrap(err, "creating workspace")
	}
	defer workspace.Close(ctx)

	results := make([]StepResult, len(steps))

	for i, step := range steps {
		reportProgress(fmt.Sprintf("Preparing step %d", i+1))

		stepContext := StepContext{Repository: *repo}
		if i > 0 {
			stepContext.PreviousStep = results[i-1]
		}

		// Find a location that we can use for a cidfile, which will contain the
		// container ID that is used below. We can then use this to remove the
		// container on a successful run, rather than leaving it dangling.
		cidFile, err := ioutil.TempFile(tempDir, repo.Slug()+"-container-id")
		if err != nil {
			return nil, errors.Wrap(err, "Creating a CID file failed")
		}

		// However, Docker will fail if the cidfile actually exists, so we need
		// to remove it. Because Windows can't remove open files, we'll first
		// close it, even though that's unnecessary elsewhere.
		cidFile.Close()
		if err = os.Remove(cidFile.Name()); err != nil {
			return nil, errors.Wrap(err, "removing cidfile")
		}

		// Since we went to all that effort, we can now defer a function that
		// uses the cidfile to clean up after this function is done.
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
		runScriptFile, err := ioutil.TempFile(tempDir, "")
		if err != nil {
			return nil, errors.Wrap(err, "creating temporary file")
		}
		defer os.Remove(runScriptFile.Name())

		// Parse step.Run as a template and render it into a buffer and the
		// temp file we just created.
		var runScript bytes.Buffer
		out := io.MultiWriter(&runScript, runScriptFile)
		if err := renderTemplate("step-run", step.Run, out, &stepContext); err != nil {
			return nil, errors.Wrap(err, "parsing step run")
		}

		if err := runScriptFile.Close(); err != nil {
			return nil, errors.Wrap(err, "closing temporary file")
		}

		// This file needs to be readable within the container regardless of the
		// user the container is running as, so we'll set the appropriate group
		// and other bits to make it so.
		//
		// A fun note: although os.File exposes a Chmod() method, we can't
		// unconditionally use it because Windows cannot change the attributes
		// of an open file. Rather than going to the trouble of having
		// conditionally compiled files here, instead we'll just wait until the
		// file is closed to twiddle the permission bits. Which is now!
		if err := os.Chmod(runScriptFile.Name(), 0644); err != nil {
			return nil, errors.Wrap(err, "setting permissions on the temporary file")
		}

		// Parse and render the step.Files.
		files, err := renderMap(step.Files, &stepContext)
		if err != nil {
			return nil, errors.Wrap(err, "parsing step files")
		}

		// Create temp files with the rendered content of step.Files so that we
		// can mount them into the container.
		filesToMount := make(map[string]*os.File, len(files))
		for name, content := range files {
			fp, err := ioutil.TempFile(tempDir, "")
			if err != nil {
				return nil, errors.Wrap(err, "creating temporary file")
			}
			defer os.Remove(fp.Name())

			if _, err := fp.WriteString(content); err != nil {
				return nil, errors.Wrap(err, "writing to temporary file")
			}

			if err := fp.Close(); err != nil {
				return nil, errors.Wrap(err, "closing temporary file")
			}

			filesToMount[name] = fp
		}

		// Resolve step.Env given the current environment.
		stepEnv, err := step.Env.Resolve(os.Environ())
		if err != nil {
			return nil, errors.Wrap(err, "resolving step environment")
		}

		// Render the step.Env variables as templates.
		env, err := renderMap(stepEnv, &stepContext)
		if err != nil {
			return nil, errors.Wrap(err, "parsing step environment")
		}

		reportProgress(runScript.String())
		const workDir = "/work"
		workspaceOpts, err := workspace.DockerRunOpts(ctx, workDir)
		if err != nil {
			return nil, errors.Wrap(err, "getting Docker options for workspace")
		}
		args := append([]string{
			"run",
			"--rm",
			"--init",
			"--cidfile", cidFile.Name(),
			"--workdir", workDir,
			"--mount", fmt.Sprintf("type=bind,source=%s,target=%s,ro", runScriptFile.Name(), containerTemp),
		}, workspaceOpts...)
		for target, source := range filesToMount {
			args = append(args, "--mount", fmt.Sprintf("type=bind,source=%s,target=%s,ro", source.Name(), target))
		}

		for k, v := range env {
			args = append(args, "-e", k+"="+v)
		}

		args = append(args, "--entrypoint", shell)

		cmd := exec.CommandContext(ctx, "docker", args...)
		cmd.Args = append(cmd.Args, "--", step.image, containerTemp)
		if dir := workspace.WorkDir(); dir != nil {
			cmd.Dir = *dir
		}

		var stdoutBuffer, stderrBuffer bytes.Buffer
		cmd.Stdout = io.MultiWriter(&stdoutBuffer, logger.PrefixWriter("stdout"))
		cmd.Stderr = io.MultiWriter(&stderrBuffer, logger.PrefixWriter("stderr"))

		logger.Logf("[Step %d] run: %q, container: %q", i+1, step.Run, step.Container)
		logger.Logf("[Step %d] full command: %q", i+1, strings.Join(cmd.Args, " "))

		t0 := time.Now()
		err = cmd.Run()
		elapsed := time.Since(t0).Round(time.Millisecond)
		if err != nil {
			logger.Logf("[Step %d] took %s; error running Docker container: %+v", i+1, elapsed, err)

			return nil, stepFailedErr{
				Err:         err,
				Args:        cmd.Args,
				Run:         runScript.String(),
				Container:   step.Container,
				TmpFilename: containerTemp,
				Stdout:      strings.TrimSpace(stdoutBuffer.String()),
				Stderr:      strings.TrimSpace(stderrBuffer.String()),
			}
		}

		logger.Logf("[Step %d] complete in %s", i+1, elapsed)

		changes, err := workspace.Changes(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "getting changed files in step")
		}

		results[i] = StepResult{files: changes, Stdout: &stdoutBuffer, Stderr: &stderrBuffer}
	}

	reportProgress("Calculating diff")
	diffOut, err := workspace.Diff(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "git diff failed")
	}

	return diffOut, err
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

type stepFailedErr struct {
	Run       string
	Container string

	TmpFilename string

	Args   []string
	Stdout string
	Stderr string

	Err error
}

func (e stepFailedErr) Cause() error { return e.Err }

func (e stepFailedErr) Error() string {
	var out strings.Builder

	fmtRun := func(run string) string {
		lines := strings.Split(run, "\n")
		if len(lines) == 1 {
			return lines[0]
		}
		return lines[0] + fmt.Sprintf("\n\t(... and %d more lines)", len(lines)-1)
	}

	out.WriteString(fmt.Sprintf("run: %s\ncontainer: %s\n", fmtRun(e.Run), e.Container))

	printOutput := func(output string) {
		for _, line := range strings.Split(output, "\n") {
			if e.TmpFilename != "" {
				line = strings.ReplaceAll(line, e.TmpFilename+": ", "")
			}
			out.WriteString("\t" + line + "\n")
		}
	}

	if len(e.Stdout) > 0 {
		out.WriteString("\nstandard out:\n")
		printOutput(e.Stdout)
	}

	if len(e.Stderr) > 0 {
		out.WriteString("\nstandard error:\n")
		printOutput(e.Stderr)
	}

	if exitErr, ok := e.Err.(*exec.ExitError); ok {
		out.WriteString(fmt.Sprintf("\nCommand failed with exit code %d.", exitErr.ExitCode()))
	} else {
		out.WriteString(fmt.Sprintf("\nCommand failed: %s", e.Err))
	}

	return out.String()
}

func (e stepFailedErr) SingleLineError() string {
	out := e.Err.Error()
	if len(e.Stderr) > 0 {
		out = e.Stderr
	}

	return strings.Split(out, "\n")[0]
}
