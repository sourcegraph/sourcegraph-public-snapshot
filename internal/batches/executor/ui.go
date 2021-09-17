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
	ArchiveDownloadFinished()

	WorkspaceInitializationStarted()
	WorkspaceInitializationFinished()

	SkippingStepsUpto(int)

	StepSkipped(int)

	StepPreparing(int)
	StepStarted(int, string)

	StepOutputWriter(context.Context, *Task, int) StepOutputWriter

	CalculatingDiffStarted()
	CalculatingDiffFinished()

	StepFinished(idx int, diff []byte, changes *git.Changes, outputs map[string]interface{})
}

// NoopStepsExecUI is an implementation of StepsExecutionUI that does nothing.
type NoopStepsExecUI struct{}

func (noop NoopStepsExecUI) ArchiveDownloadStarted()                {}
func (noop NoopStepsExecUI) ArchiveDownloadFinished()               {}
func (noop NoopStepsExecUI) WorkspaceInitializationStarted()        {}
func (noop NoopStepsExecUI) WorkspaceInitializationFinished()       {}
func (noop NoopStepsExecUI) SkippingStepsUpto(startStep int)        {}
func (noop NoopStepsExecUI) StepSkipped(step int)                   {}
func (noop NoopStepsExecUI) StepPreparing(step int)                 {}
func (noop NoopStepsExecUI) StepStarted(step int, runScript string) {}
func (noop NoopStepsExecUI) StepOutputWriter(ctx context.Context, task *Task, step int) StepOutputWriter {
	return NoopStepOutputWriter{}
}
func (noop NoopStepsExecUI) CalculatingDiffStarted()  {}
func (noop NoopStepsExecUI) CalculatingDiffFinished() {}
func (noop NoopStepsExecUI) StepFinished(idx int, diff []byte, changes *git.Changes, outputs map[string]interface{}) {
}

type NoopStepOutputWriter struct{}

func (noop NoopStepOutputWriter) StdoutWriter() io.Writer { return io.Discard }
func (noop NoopStepOutputWriter) StderrWriter() io.Writer { return io.Discard }
func (noop NoopStepOutputWriter) Close() error            { return nil }
