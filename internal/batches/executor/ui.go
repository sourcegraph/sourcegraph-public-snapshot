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

type StepsExecutionUI interface {
	ArchiveDownloadStarted()
	ArchiveDownloadFinished()

	WorkspaceInitializationStarted()
	WorkspaceInitializationFinished()

	SkippingStepsUpto(int)

	StepSkipped(int)

	StepPreparing(int)
	StepStarted(int, string)

	StepStdoutWriter(context.Context, *Task, int) io.WriteCloser
	StepStderrWriter(context.Context, *Task, int) io.WriteCloser

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
func (noop NoopStepsExecUI) StepStdoutWriter(ctx context.Context, task *Task, step int) io.WriteCloser {
	return discardCloser{io.Discard}
}
func (noop NoopStepsExecUI) StepStderrWriter(ctx context.Context, task *Task, step int) io.WriteCloser {
	return discardCloser{io.Discard}
}
func (noop NoopStepsExecUI) CalculatingDiffStarted()  {}
func (noop NoopStepsExecUI) CalculatingDiffFinished() {}
func (noop NoopStepsExecUI) StepFinished(idx int, diff []byte, changes *git.Changes, outputs map[string]interface{}) {
}

type discardCloser struct {
	io.Writer
}

func (discardCloser) Close() error { return nil }
