package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/batches/log"
	"github.com/sourcegraph/src-cli/internal/batches/util"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"

	"github.com/sourcegraph/sourcegraph/lib/process"

	yamlv3 "gopkg.in/yaml.v3"
)

type runStepsOpts struct {
	// wc is the workspace creator to use. It will be called at the beginning of
	// runSteps to prepare the workspace for execution.
	wc workspace.Creator
	// ensureImage is called in runSteps to make sure the used image has been
	// pulled from the registry.
	ensureImage imageEnsurer
	// task is the definition of the workspace execution.
	task *Task
	// tempDir points to where temporary files of the execution should live at.
	tempDir string

	logger log.TaskLogger

	ui StepsExecutionUI

	// isRemote toggles server-side execution mode. This disables file mounts using
	// step.mounts.
	isRemote bool

	// globalEnv is the os.Environ() for the execution. We don't read from os.Environ()
	// directly to allow injecting variables and hiding others.
	globalEnv []string
}

func runSteps(ctx context.Context, opts *runStepsOpts) (stepResults []execution.AfterStepResult, err error) {
	opts.ui.ArchiveDownloadStarted()
	err = opts.task.RepoArchive.Ensure(ctx)
	opts.ui.ArchiveDownloadFinished(err)
	if err != nil {
		return nil, errors.Wrap(err, "fetching repo")
	}
	defer opts.task.RepoArchive.Close()

	opts.ui.WorkspaceInitializationStarted()
	ws, err := opts.wc.Create(ctx, opts.task.Repository, opts.task.Steps, opts.task.RepoArchive)
	if err != nil {
		return nil, errors.Wrap(err, "creating workspace")
	}
	defer ws.Close(ctx)
	opts.ui.WorkspaceInitializationFinished()

	var (
		// lastOutputs are the step.outputs from the previously run step.
		// Outputs are additive, so we will only ever append to this map,
		// but never clear it.
		lastOutputs        = make(map[string]any)
		previousStepResult execution.AfterStepResult
		// startStep is the index of the step we start executing with. Set when
		// there's a partial cache result for the task.
		startStep int
	)

	if opts.task.CachedStepResultFound {
		// Set the Outputs to the cached outputs
		lastOutputs = opts.task.CachedStepResult.Outputs
		previousStepResult = opts.task.CachedStepResult
		lastStep := opts.task.CachedStepResult.StepIndex
		// Add the cached result to the result set. It may be required for building
		// the changeset specs in the end if we don't need to execute any other steps
		// due to dynamic skipping.
		// We remove it at the end of this function if it isn't needed, so we don't
		// write it to the cache again.
		stepResults = append(stepResults, previousStepResult)

		// If we have cached results and don't need to execute any more steps,
		// we can quit.
		// TODO: This doesn't consider skipped steps, but it's not _that_
		// bad, we will find out about that further down and correctly finish
		// execution early.
		if lastStep == len(opts.task.Steps)-1 {
			return stepResults, nil
		}

		// If the previous steps made any modifications to the workspace yet,
		// apply them.
		if opts.task.CachedStepResult.Diff != "" {
			if err := ws.ApplyDiff(ctx, []byte(opts.task.CachedStepResult.Diff)); err != nil {
				return nil, errors.Wrap(err, "applying diff of cache result")
			}
		}

		startStep = lastStep + 1

		opts.ui.SkippingStepsUpto(
			// UI is 1-indexed.
			startStep + 1,
		)
	}

	for i := startStep; i < len(opts.task.Steps); i++ {
		step := opts.task.Steps[i]

		stepContext := template.StepContext{
			BatchChange: *opts.task.BatchChangeAttributes,
			Repository:  util.NewTemplatingRepo(opts.task.Repository.Name, opts.task.Repository.FileMatches),
			Outputs:     lastOutputs,
			Steps: template.StepsContext{
				Path:    opts.task.Path,
				Changes: previousStepResult.ChangedFiles,
			},
			PreviousStep: previousStepResult,
		}

		// Check if the step needs to be skipped.
		cond, err := template.EvalStepCondition(step.IfCondition(), &stepContext)
		if err != nil {
			return nil, errors.Wrap(err, "evaluating step condition")
		}
		if !cond {
			opts.ui.StepSkipped(i + 1)
			continue
		}

		// We need to grab the digest for the exact image we're using.
		img, err := opts.ensureImage(ctx, step.Container)
		if err != nil {
			return nil, err
		}
		digest, err := img.Digest(ctx)
		if err != nil {
			return nil, err
		}

		stdoutBuffer, stderrBuffer, err := executeSingleStep(ctx, opts, ws, i, step, digest, &stepContext)
		defer func() {
			if err != nil {
				exitCode := -1
				sfe := &stepFailedErr{}
				if errors.As(err, sfe) {
					exitCode = sfe.ExitCode
				}
				opts.ui.StepFailed(i+1, err, exitCode)
			}
		}()
		if err != nil {
			return stepResults, err
		}

		changes, err := ws.Changes(ctx)
		if err != nil {
			return stepResults, errors.Wrap(err, "getting changed files in step")
		}

		// Get the current diff and store that away as the per-step result.
		stepDiff, err := ws.Diff(ctx)
		if err != nil {
			return stepResults, errors.Wrap(err, "getting diff produced by step")
		}

		stepResult := execution.AfterStepResult{
			ChangedFiles: changes,
			Stdout:       stdoutBuffer.String(),
			Stderr:       stderrBuffer.String(),
			StepIndex:    i,
			Diff:         string(stepDiff),
			// Those will be set below.
			Outputs: make(map[string]interface{}),
		}

		// Set stepContext.Step to current step's results before rendering outputs.
		stepContext.Step = stepResult
		// Render and evaluate outputs.
		if err := setOutputs(step.Outputs, lastOutputs, &stepContext); err != nil {
			return stepResults, errors.Wrap(err, "setting step outputs")
		}
		for k, v := range lastOutputs {
			stepResult.Outputs[k] = v
		}
		stepResults = append(stepResults, stepResult)
		previousStepResult = stepResult

		opts.ui.StepFinished(i+1, stepResult.Diff, stepResult.ChangedFiles, stepResult.Outputs)
	}

	// If we ended up executing at least one step in addition to the cached step,
	// we can remove it from the result set. Otherwise we write it to the cache
	// again which is unnecessary. We only need to return it when we want to build changeset
	// specs from it.
	if len(stepResults) > 1 && opts.task.CachedStepResultFound {
		stepResults = stepResults[1:]
	}

	return stepResults, err
}

const workDir = "/work"

func executeSingleStep(
	ctx context.Context,
	opts *runStepsOpts,
	workspace workspace.Workspace,
	stepIdx int,
	step batcheslib.Step,
	imageDigest string,
	stepContext *template.StepContext,
) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	// ----------
	// PREPARATION
	// ----------
	opts.ui.StepPreparingStart(stepIdx + 1)

	cidFile, cleanup, err := createCidFile(ctx, opts.tempDir, util.SlugForRepo(opts.task.Repository.Name, opts.task.Repository.Rev()))
	if err != nil {
		opts.ui.StepPreparingFailed(stepIdx+1, err)
		return bytes.Buffer{}, bytes.Buffer{}, err
	}
	defer cleanup()

	// For now, we only support shell scripts provided via the Run field.
	shell, containerTemp, err := probeImageForShell(ctx, imageDigest)
	if err != nil {
		err = errors.Wrapf(err, "probing image %q for shell", step.Container)
		opts.ui.StepPreparingFailed(stepIdx+1, err)
		return bytes.Buffer{}, bytes.Buffer{}, err
	}

	runScriptFile, runScript, cleanup, err := createRunScriptFile(ctx, opts.tempDir, step.Run, stepContext)
	if err != nil {
		opts.ui.StepPreparingFailed(stepIdx+1, err)
		return bytes.Buffer{}, bytes.Buffer{}, err
	}
	defer cleanup()

	// Parse and render the step.Files.
	filesToMount, cleanup, err := createFilesToMount(opts.tempDir, step, stepContext)
	if err != nil {
		opts.ui.StepPreparingFailed(stepIdx+1, err)
		return bytes.Buffer{}, bytes.Buffer{}, err
	}
	defer cleanup()

	// Resolve step.Env given the current environment.
	stepEnv, err := step.Env.Resolve(opts.globalEnv)
	if err != nil {
		err = errors.Wrap(err, "resolving step environment")
		opts.ui.StepPreparingFailed(stepIdx+1, err)
		return bytes.Buffer{}, bytes.Buffer{}, err
	}
	// Render the step.Env variables as templates.
	env, err := template.RenderStepMap(stepEnv, stepContext)
	if err != nil {
		err = errors.Wrap(err, "parsing step environment")
		opts.ui.StepPreparingFailed(stepIdx+1, err)
		return bytes.Buffer{}, bytes.Buffer{}, err
	}

	opts.ui.StepPreparingSuccess(stepIdx + 1)

	// ----------
	// EXECUTION
	// ----------
	opts.ui.StepStarted(stepIdx+1, runScript, env)

	workspaceOpts, err := workspace.DockerRunOpts(ctx, workDir)
	if err != nil {
		return bytes.Buffer{}, bytes.Buffer{}, errors.Wrap(err, "getting Docker options for workspace")
	}

	// Where should we execute the steps.run script?
	scriptWorkDir := workDir
	if opts.task.Path != "" {
		scriptWorkDir = workDir + "/" + opts.task.Path
	}

	args := append([]string{
		"run",
		"--rm",
		"--init",
		"--cidfile", cidFile,
		"--workdir", scriptWorkDir,
		"--mount", fmt.Sprintf("type=bind,source=%s,target=%s,ro", runScriptFile, containerTemp),
	}, workspaceOpts...)

	for target, source := range filesToMount {
		args = append(args, "--mount", fmt.Sprintf("type=bind,source=%s,target=%s,ro", source.Name(), target))
	}

	// Temporarily add a guard to prevent a path to mount path for server-side processing.
	if !opts.isRemote {
		// Mount any paths on the local system to the docker container. The paths have already been validated during parsing.
		for _, mount := range step.Mount {
			args = append(args, "--mount", fmt.Sprintf("type=bind,source=%s,target=%s,ro", mount.Path, mount.Mountpoint))
		}
	}

	for k, v := range env {
		args = append(args, "-e", k+"="+v)
	}

	args = append(args, "--entrypoint", shell)

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Args = append(cmd.Args, "--", imageDigest, containerTemp)
	if dir := workspace.WorkDir(); dir != nil {
		cmd.Dir = *dir
	}

	writerCtx, writerCancel := context.WithCancel(ctx)
	defer writerCancel()
	outputWriter := opts.ui.StepOutputWriter(writerCtx, opts.task, stepIdx+1)
	defer func() {
		outputWriter.Close()
	}()

	stdoutWriter := io.MultiWriter(&stdout, outputWriter.StdoutWriter(), opts.logger.PrefixWriter("stdout"))
	stderrWriter := io.MultiWriter(&stderr, outputWriter.StderrWriter(), opts.logger.PrefixWriter("stderr"))

	// Setup readers that pipe the output into the given buffers
	wg, err := process.PipeOutput(ctx, cmd, stdoutWriter, stderrWriter)
	if err != nil {
		return stdout, stderr, errors.Wrap(err, "piping process output")
	}

	newStepFailedErr := func(wrappedErr error) stepFailedErr {
		exitCode := -1
		exitErr := &exec.ExitError{}
		if errors.As(wrappedErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return stepFailedErr{
			Err:         wrappedErr,
			ExitCode:    exitCode,
			Args:        cmd.Args,
			Run:         runScript,
			Container:   step.Container,
			TmpFilename: containerTemp,
			Stdout:      strings.TrimSpace(stdout.String()),
			Stderr:      strings.TrimSpace(stderr.String()),
		}
	}

	opts.logger.Logf("[Step %d] run: %q, container: %q", stepIdx+1, step.Run, step.Container)
	opts.logger.Logf("[Step %d] full command: %q", stepIdx+1, strings.Join(cmd.Args, " "))

	// Start the command.
	t0 := time.Now()
	if err := cmd.Start(); err != nil {
		opts.logger.Logf("[Step %d] error starting Docker container: %+v", stepIdx+1, err)
		return stdout, stderr, newStepFailedErr(err)
	}

	// Wait for the readers, because the pipes used by PipeOutput under the
	// hood are closed when the command exits.
	wg.Wait()

	// Now wait for the command.
	err = cmd.Wait()
	elapsed := time.Since(t0).Round(time.Millisecond)
	if err != nil {
		opts.logger.Logf("[Step %d] took %s; error running Docker container: %+v", stepIdx+1, elapsed, err)
		return stdout, stderr, newStepFailedErr(err)
	}

	opts.logger.Logf("[Step %d] complete in %s", stepIdx+1, elapsed)
	return stdout, stderr, nil
}

func setOutputs(stepOutputs batcheslib.Outputs, global map[string]interface{}, stepCtx *template.StepContext) error {
	for name, output := range stepOutputs {
		var value bytes.Buffer

		if err := template.RenderStepTemplate("outputs-"+name, output.Value, &value, stepCtx); err != nil {
			return errors.Wrap(err, "parsing step run")
		}

		switch output.Format {
		case "yaml":
			var out interface{}
			// We use yamlv3 here, because it unmarshals YAML into
			// map[string]interface{} which we need to serialize it back to
			// JSON when we cache the results.
			// See https://github.com/go-yaml/yaml/issues/139 for context
			if err := yamlv3.NewDecoder(&value).Decode(&out); err != nil {
				return err
			}
			global[name] = out
		case "json":
			var out interface{}
			if err := json.NewDecoder(&value).Decode(&out); err != nil {
				return err
			}
			global[name] = out
		default:
			global[name] = value.String()
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

	// We can iterate through our shell options and try to run mktemp with them.
	for _, shell = range []string{"/bin/bash", "/bin/sh"} {
		stdout := new(bytes.Buffer)
		stderr := new(bytes.Buffer)

		args := []string{"run", "--rm", "--entrypoint", shell, image, "-c", "mktemp"}

		cmd := exec.CommandContext(ctx, "docker", args...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		if runErr := cmd.Run(); runErr != nil {
			err = errors.Append(err, errors.Wrapf(runErr, "probing shell %q:\n%s", shell, stderr.String()))
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

// createFilesToMount creates temporary files with the contents of Step.Files
// that are to be mounted into the container that executes the step.
func createFilesToMount(tempDir string, step batcheslib.Step, stepContext *template.StepContext) (map[string]*os.File, func(), error) {
	// Parse and render the step.Files.
	files, err := template.RenderStepMap(step.Files, stepContext)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parsing step files")
	}

	var toCleanup []string
	cleanup := func() {
		for _, fname := range toCleanup {
			os.Remove(fname)
		}
	}

	// Create temp files with the rendered content of step.Files so that we
	// can mount them into the container.
	filesToMount := make(map[string]*os.File, len(files))
	for name, content := range files {
		fp, err := os.CreateTemp(tempDir, "")
		if err != nil {
			return nil, cleanup, errors.Wrap(err, "creating temporary file")
		}
		toCleanup = append(toCleanup, fp.Name())

		if _, err := fp.WriteString(content); err != nil {
			return nil, cleanup, errors.Wrap(err, "writing to temporary file")
		}

		if err := fp.Close(); err != nil {
			return nil, cleanup, errors.Wrap(err, "closing temporary file")
		}

		filesToMount[name] = fp
	}

	return filesToMount, cleanup, nil
}

// createRunScriptFile creates a temporary file and renders stepRun into it.
//
// It returns the location of the file, its content, a function to cleanup the file and possible errors.
func createRunScriptFile(ctx context.Context, tempDir string, stepRun string, stepCtx *template.StepContext) (string, string, func(), error) {
	// Set up a temporary file on the host filesystem to contain the
	// script.
	runScriptFile, err := os.CreateTemp(tempDir, "")
	if err != nil {
		return "", "", nil, errors.Wrap(err, "creating temporary file")
	}
	cleanup := func() { os.Remove(runScriptFile.Name()) }

	// Parse step.Run as a template and render it into a buffer and the
	// temp file we just created.
	var runScript bytes.Buffer
	out := io.MultiWriter(&runScript, runScriptFile)
	if err := template.RenderStepTemplate("step-run", stepRun, out, stepCtx); err != nil {
		return "", "", nil, errors.Wrap(err, "parsing step run")
	}

	if err := runScriptFile.Close(); err != nil {
		return "", "", nil, errors.Wrap(err, "closing temporary file")
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
		return "", "", nil, errors.Wrap(err, "setting permissions on the temporary file")
	}

	return runScriptFile.Name(), runScript.String(), cleanup, nil
}

// createCidFile creates a temporary file that will contain the container ID
// when executing steps.
// It returns the location of the file and a function that cleans up the
// file.
func createCidFile(ctx context.Context, tempDir string, repoSlug string) (string, func(), error) {
	// Find a location that we can use for a cidfile, which will contain the
	// container ID that is used below. We can then use this to remove the
	// container on a successful run, rather than leaving it dangling.
	cidFile, err := os.CreateTemp(tempDir, repoSlug+"-container-id")
	if err != nil {
		return "", nil, errors.Wrap(err, "Creating a CID file failed")
	}

	// However, Docker will fail if the cidfile actually exists, so we need
	// to remove it. Because Windows can't remove open files, we'll first
	// close it, even though that's unnecessary elsewhere.
	cidFile.Close()
	if err = os.Remove(cidFile.Name()); err != nil {
		return "", nil, errors.Wrap(err, "removing cidfile")
	}

	// Since we went to all that effort, we can now defer a function that
	// uses the cidfile to clean up after this function is done.
	cleanup := func() {
		cid, err := os.ReadFile(cidFile.Name())
		_ = os.Remove(cidFile.Name())
		if err == nil {
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			_ = exec.CommandContext(ctx, "docker", "rm", "-f", "--", string(cid)).Run()
		}
	}

	return cidFile.Name(), cleanup, nil
}

type stepFailedErr struct {
	Run       string
	Container string

	TmpFilename string

	Args   []string
	Stdout string
	Stderr string

	// ExitCode of the command, or -1 if a non-command error occured.
	ExitCode int
	Err      error
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

	if e.ExitCode != -1 {
		out.WriteString(fmt.Sprintf("\nCommand failed with exit code %d.", e.ExitCode))
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
