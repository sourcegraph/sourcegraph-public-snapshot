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
	"text/template"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/campaigns/graphql"
)

func runSteps(ctx context.Context, wc *WorkspaceCreator, repo *graphql.Repository, steps []Step, logger *TaskLogger, tempDir string, reportProgress func(string)) ([]byte, error) {
	reportProgress("Downloading archive")

	volumeDir, err := wc.Create(ctx, repo)
	if err != nil {
		return nil, errors.Wrap(err, "creating workspace")
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

	reportProgress("Initializing workspace")
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

	results := make([]StepResult, len(steps))

	for i, step := range steps {
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

		// Parse step.Run as a template...
		tmpl, err := parseAsTemplate("step-run", step.Run, &stepContext)
		if err != nil {
			return nil, errors.Wrap(err, "parsing step run")
		}

		// ... and render it into a buffer and the temp file we just created.
		var runScript bytes.Buffer
		if err := tmpl.Execute(io.MultiWriter(&runScript, runScriptFile), stepContext); err != nil {
			return nil, errors.Wrap(err, "executing template")
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
		files, err := renderStepFiles(step.Files, &stepContext)
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

			if _, err := io.Copy(fp, content); err != nil {
				return nil, errors.Wrap(err, "writing to temporary file")
			}

			if err := fp.Close(); err != nil {
				return nil, errors.Wrap(err, "closing temporary file")
			}

			filesToMount[name] = fp
		}

		// Render the step.Env variables as templates.
		env, err := renderStepEnv(step.Env, &stepContext)
		if err != nil {
			return nil, errors.Wrap(err, "parsing step files")
		}

		reportProgress(runScript.String())
		const workDir = "/work"
		args := []string{
			"run",
			"--rm",
			"--init",
			"--cidfile", cidFile.Name(),
			"--workdir", workDir,
			"--mount", fmt.Sprintf("type=bind,source=%s,target=%s", volumeDir, workDir),
			"--mount", fmt.Sprintf("type=bind,source=%s,target=%s,ro", runScriptFile.Name(), containerTemp),
		}
		for target, source := range filesToMount {
			args = append(args, "--mount", fmt.Sprintf("type=bind,source=%s,target=%s,ro", source.Name(), target))
		}

		for k, v := range env {
			args = append(args, "-e", k+"="+v)
		}

		args = append(args, "--entrypoint", shell)

		cmd := exec.CommandContext(ctx, "docker", args...)
		cmd.Args = append(cmd.Args, "--", step.image, containerTemp)
		cmd.Dir = volumeDir

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

		if _, err := runGitCmd("add", "--all"); err != nil {
			return nil, errors.Wrap(err, "git add failed")
		}

		statusOut, err := runGitCmd("status", "--porcelain")
		if err != nil {
			return nil, errors.Wrap(err, "git status failed")
		}

		changes, err := parseGitStatus(statusOut)
		if err != nil {
			return nil, errors.Wrap(err, "parsing git status output")
		}

		results[i] = StepResult{Files: changes, Stdout: &stdoutBuffer, Stderr: &stderrBuffer}
	}

	reportProgress("Calculating diff")
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

func parseAsTemplate(name, input string, stepCtx *StepContext) (*template.Template, error) {
	return template.New(name).Delims("${{", "}}").Funcs(stepCtx.ToFuncMap()).Parse(input)
}

func renderStepFiles(files map[string]string, stepCtx *StepContext) (map[string]io.Reader, error) {
	containerFiles := make(map[string]io.Reader, len(files))

	for fileName, fileRaw := range files {
		// We treat the file contents as a template and render it
		// into a buffer that we then mount into the code host.
		var out bytes.Buffer

		tmpl, err := parseAsTemplate(fileName, fileRaw, stepCtx)
		if err != nil {
			return containerFiles, err
		}

		if err := tmpl.Execute(&out, stepCtx); err != nil {
			return containerFiles, err
		}

		containerFiles[fileName] = &out
	}

	return containerFiles, nil
}

func renderStepEnv(env map[string]string, stepCtx *StepContext) (map[string]string, error) {
	parsedEnv := make(map[string]string, len(env))

	fnMap := stepCtx.ToFuncMap()

	for k, v := range env {
		// We treat the file contents as a template and render it
		// into a buffer that we then mount into the code host.
		var out bytes.Buffer

		tmpl, err := template.New(k).Delims("${{", "}}").Funcs(fnMap).Parse(v)
		if err != nil {
			return parsedEnv, err
		}

		if err := tmpl.Execute(&out, stepCtx); err != nil {
			return parsedEnv, err
		}

		parsedEnv[k] = out.String()
	}

	return parsedEnv, nil
}

// StepContext represents the contextual information available when executing a
// step that's defined in a campaign spec.
type StepContext struct {
	PreviousStep StepResult
	Repository   graphql.Repository
}

// ToFuncMap returns a template.FuncMap to access fields on the StepContext in a
// text/template.
func (stepCtx *StepContext) ToFuncMap() template.FuncMap {
	return template.FuncMap{
		"join": func(list []string, sep string) string {
			return strings.Join(list, sep)
		},
		"split": func(s string, sep string) []string {
			return strings.Split(s, sep)
		},
		"previous_step": func() map[string]interface{} {
			result := map[string]interface{}{
				"modified_files": stepCtx.PreviousStep.ModifiedFiles(),
				"added_files":    stepCtx.PreviousStep.AddedFiles(),
				"deleted_files":  stepCtx.PreviousStep.DeletedFiles(),
				"renamed_files":  stepCtx.PreviousStep.RenamedFiles(),
			}

			if stepCtx.PreviousStep.Stdout != nil {
				result["stdout"] = stepCtx.PreviousStep.Stdout.String()
			} else {
				result["stdout"] = ""
			}

			if stepCtx.PreviousStep.Stderr != nil {
				result["stderr"] = stepCtx.PreviousStep.Stderr.String()
			} else {
				result["stderr"] = ""
			}

			return result
		},
		"repository": func() map[string]interface{} {
			return map[string]interface{}{
				"search_result_paths": stepCtx.Repository.SearchResultPaths(),
				"name":                stepCtx.Repository.Name,
			}
		},
	}
}

// StepResult represents the result of a previously executed step.
type StepResult struct {
	// Files are the changes made to files by the step.
	Files StepChanges

	// Stdout is the output produced by the step on standard out.
	Stdout *bytes.Buffer
	// Stderr is the output produced by the step on standard error.
	Stderr *bytes.Buffer
}

// StepChanges are the changes made to files by a previous step in a repository.
type StepChanges struct {
	Modified []string
	Added    []string
	Deleted  []string
	Renamed  []string
}

// ModifiedFiles returns the files modified by a step.
func (r StepResult) ModifiedFiles() []string { return r.Files.Modified }

// AddedFiles returns the files added by a step.
func (r StepResult) AddedFiles() []string { return r.Files.Added }

// DeletedFiles returns the files deleted by a step.
func (r StepResult) DeletedFiles() []string { return r.Files.Deleted }

// RenamedFiles returns the new name of files that have been renamed by a step.
func (r StepResult) RenamedFiles() []string { return r.Files.Renamed }

func parseGitStatus(out []byte) (StepChanges, error) {
	result := StepChanges{}

	stripped := strings.TrimSpace(string(out))
	if len(stripped) == 0 {
		return result, nil
	}

	for _, line := range strings.Split(stripped, "\n") {
		if len(line) < 4 {
			return result, fmt.Errorf("git status line has unrecognized format: %q", line)
		}

		file := line[3:]

		switch line[0] {
		case 'M':
			result.Modified = append(result.Modified, file)
		case 'A':
			result.Added = append(result.Added, file)
		case 'D':
			result.Deleted = append(result.Deleted, file)
		case 'R':
			files := strings.Split(file, " -> ")
			newFile := files[len(files)-1]
			result.Renamed = append(result.Renamed, newFile)
		}
	}

	return result, nil
}
