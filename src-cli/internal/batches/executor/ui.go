package executor

import (
	"context"
	"io"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
)

type TaskExecutionUI interface {
	Start([]*Task)
	Success()
	Failed(err error)

	TaskStarted(*Task)
	TaskFinished(*Task, error)

	TaskChangesetSpecsBuilt(*Task, []*batcheslib.ChangesetSpec)

	StepsExecutionUI(*Task) StepsExecutionUI
}

type StepOutputWriter interface {
	StdoutWriter() io.Writer
	StderrWriter() io.Writer
	Close() error
}

type StepsExecutionUI interface {
	ArchiveDownloadStarted()
	ArchiveDownloadFinished(error)

	WorkspaceInitializationStarted()
	WorkspaceInitializationFinished()

	SkippingStepsUpto(int)

	StepSkipped(int)

	StepPreparingStart(int)
	StepPreparingSuccess(int)
	StepPreparingFailed(int, error)
	StepStarted(stepIdx int, runScript string, env map[string]string)

	StepOutputWriter(context.Context, *Task, int) StepOutputWriter

	StepFinished(idx int, diff []byte, changes git.Changes, outputs map[string]interface{})
	StepFailed(idx int, err error, exitCode int)
}

// NoopStepsExecUI is an implementation of StepsExecutionUI that does nothing.
type NoopStepsExecUI struct{}

func (noop NoopStepsExecUI) ArchiveDownloadStarted()                                       {}
func (noop NoopStepsExecUI) ArchiveDownloadFinished(error)                                 {}
func (noop NoopStepsExecUI) WorkspaceInitializationStarted()                               {}
func (noop NoopStepsExecUI) WorkspaceInitializationFinished()                              {}
func (noop NoopStepsExecUI) SkippingStepsUpto(startStep int)                               {}
func (noop NoopStepsExecUI) StepSkipped(step int)                                          {}
func (noop NoopStepsExecUI) StepPreparingStart(step int)                                   {}
func (noop NoopStepsExecUI) StepPreparingSuccess(step int)                                 {}
func (noop NoopStepsExecUI) StepPreparingFailed(step int, err error)                       {}
func (noop NoopStepsExecUI) StepStarted(step int, runScript string, env map[string]string) {}
func (noop NoopStepsExecUI) StepOutputWriter(ctx context.Context, task *Task, step int) StepOutputWriter {
	return NoopStepOutputWriter{}
}
func (noop NoopStepsExecUI) StepFinished(idx int, diff []byte, changes git.Changes, outputs map[string]interface{}) {
}
func (noop NoopStepsExecUI) StepFailed(idx int, err error, exitCode int) {
}

type NoopStepOutputWriter struct{}

func (noop NoopStepOutputWriter) StdoutWriter() io.Writer { return io.Discard }
func (noop NoopStepOutputWriter) StderrWriter() io.Writer { return io.Discard }
func (noop NoopStepOutputWriter) Close() error            { return nil }
