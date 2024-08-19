package run

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/djherbis/nio/v3"
	"go.bobheadxi.dev/streamline"
	"go.bobheadxi.dev/streamline/pipeline"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Output configures output and aggregation from a command.
//
// It is behind an interface to more easily enable mock outputs and build different types
// of outputs, such as multi-outputs and error-only outputs, without complicating the core
// commandOutput implementation.
type Output interface {
	// Map adds a LineMap function to be applied to this Output.
	//
	// Deprecated: Use Pipeline instead.
	Map(f LineMap) Output
	// Pipeline is similar to Map, but uses a new interface that provides more flexible
	// ways of manipulating output on the stream. It is only applied at aggregation time
	// using e.g. Stream, Lines, and so on. Multiple Pipelines are applied sequentially,
	// with the result of previous Pipelines propagated to subsequent Pipelines.
	//
	// For more details, refer to the pipeline.Pipeline documentation.
	Pipeline(p pipeline.Pipeline) Output

	// TODO wishlist functionality
	// Mode(mode OutputMode) Output

	// Stream writes mapped output from the command to the destination writer until
	// command completion.
	Stream(dst io.Writer) error
	// StreamLines writes mapped output from the command and sends it line by line to the
	// destination callback until command completion.
	StreamLines(dst func(line string)) error
	// Lines waits for command completion and aggregates mapped output from the command as
	// a slice of lines.
	Lines() ([]string, error)
	// Lines waits for command completion and aggregates mapped output from the command as
	// a combined string.
	String() (string, error)
	// JQ waits for command completion executes a JQ query against the entire output.
	//
	// Refer to https://github.com/itchyny/gojq for the specifics of supported syntax.
	JQ(query string) ([]byte, error)
	// Reader is implemented so that Output can be provided directly to another Command
	// using Input().
	io.Reader
	// WriterTo is implemented for convenience when chaining commands in LineMap.
	io.WriterTo

	// Wait waits for command completion and returns.
	Wait() error
}

// commandOutput is the core Output implementation, designed to be attached to an exec.Cmd.
//
// It only handles piping output and configuration - aggregation is handled by the embedded
// aggregator.
//
// All aggregation functions should take care that all output collected is returned,
// regardless of whether read operations return errors.
type commandOutput struct {
	ctx context.Context

	// stream is the underlying output aggregation implementation. It reads from a
	// read side of a pipe which receives output from a command.
	stream *streamline.Stream

	// waitAndCloseFunc should only be called via doWaitOnce(). It should wait for command
	// exit and handle setting an error such that once reads from reader are complete, the
	// reader should return the error from the command.
	waitAndCloseFunc func() error
	waitAndCloseOnce sync.Once
}

var _ Output = &commandOutput{}

type attachedOutput int

const (
	attachCombined   attachedOutput = 0
	attachOnlyStdOut attachedOutput = 1
	attachOnlyStdErr attachedOutput = 2
)

// attachOutputAndRun is called by (*Command).Run() to start command execution and collect
// command output.
func attachAndRun(
	ctx context.Context,
	attachOutput attachedOutput,
	attachInput io.Reader,
	executedCmd ExecutedCommand,
) Output {
	// Set up command
	cmd := exec.CommandContext(ctx, executedCmd.Args[0], executedCmd.Args[1:]...)
	cmd.Dir = executedCmd.Dir
	cmd.Env = executedCmd.Environ
	cmd.Stdin = attachInput

	// Prepare tracing
	tracer, attrs := getTracer(ctx)
	// span should manually be ended in error scenarios - make sure each code path that
	// should end the span appropriately ends the span before returning.
	var span trace.Span
	ctx, span = tracer.Start(ctx, "Run "+cmd.Path, trace.WithAttributes(attrs(executedCmd)...))

	// Set up buffers for output and errors - we need to retain a copy of stderr for error
	// creation.
	var outputBuffer, stderrCopy = makeUnboundedBuffer(), makeUnboundedBuffer()

	// We use this buffered pipe from github.com/djherbis/nio that allows async read and
	// write operations to the reader and writer portions of the pipe respectively.
	outputReader, outputWriter := nio.Pipe(outputBuffer)

	// Set up output hooks
	switch attachOutput {
	case attachCombined:
		cmd.Stdout = outputWriter
		cmd.Stderr = io.MultiWriter(stderrCopy, outputWriter)

	case attachOnlyStdOut:
		cmd.Stdout = outputWriter
		cmd.Stderr = stderrCopy

	case attachOnlyStdErr:
		cmd.Stdout = nil // discard
		cmd.Stderr = io.MultiWriter(stderrCopy, outputWriter)

	default:
		err := fmt.Errorf("unexpected attach type %d", attachOutput)
		span.RecordError(err)
		span.SetStatus(codes.Error, "")
		span.End()
		return NewErrorOutput(err)
	}

	// Log and start command execution
	if log := getLogger(ctx); log != nil {
		log(executedCmd)
	}
	if err := cmd.Start(); err != nil {
		err := fmt.Errorf("failed to start command: %w", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, "")
		span.End()
		return NewErrorOutput(err)
	}

	output := &commandOutput{
		ctx:    ctx,
		stream: streamline.New(outputReader),
	}

	output.waitAndCloseFunc = func() error {
		// In the happy case, this is where we end the span - when the command finishes
		// and all resources are closed.
		defer span.End()

		err := newError(cmd.Wait(), stderrCopy)
		span.AddEvent("Done") // add done event because some time may elapse before span end
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "")
		}

		// CloseWithError makes it so that when all output has been consumed from the
		// reader, the given error is returned.
		outputWriter.CloseWithError(err)

		return err
	}

	return output
}

func (o *commandOutput) Map(f LineMap) Output {
	return o.Pipeline(&lineMapPipelineAdapter{
		ctx:     o.ctx,
		buffer:  &bytes.Buffer{},
		lineMap: f,
	})
}

func (o *commandOutput) Pipeline(p pipeline.Pipeline) Output {
	o.stream = o.stream.WithPipeline(p)
	return o
}

func (o *commandOutput) Stream(dst io.Writer) error {
	trace.SpanFromContext(o.ctx).AddEvent("Stream")

	_, err := o.WriteTo(dst)
	return err
}

func (o *commandOutput) StreamLines(dst func(line string)) error {
	trace.SpanFromContext(o.ctx).AddEvent("StreamLines")

	go o.waitAndClose()

	return o.stream.Stream(dst)
}

func (o *commandOutput) Lines() ([]string, error) {
	trace.SpanFromContext(o.ctx).AddEvent("Lines")

	go o.waitAndClose()

	return o.stream.Lines()
}

func (o *commandOutput) JQ(query string) ([]byte, error) {
	trace.SpanFromContext(o.ctx).AddEvent("JQ")

	jqCode, err := buildJQ(query)
	if err != nil {
		// Record this error because it is not related to reading/writing
		trace.SpanFromContext(o.ctx).RecordError(err)
		return nil, err
	}

	return execJQ(o.ctx, jqCode, o)
}

func (o *commandOutput) String() (string, error) {
	trace.SpanFromContext(o.ctx).AddEvent("String")

	go o.waitAndClose()

	return o.stream.String()
}

func (o *commandOutput) Read(p []byte) (int, error) {
	trace.SpanFromContext(o.ctx).AddEvent("Read")

	go o.waitAndClose()

	return o.stream.Read(p)
}

// WriteTo implements io.WriterTo, and returns int64 instead of int because of:
// https://stackoverflow.com/questions/29658892/why-does-io-writertos-writeto-method-return-an-int64-rather-than-an-int
func (o *commandOutput) WriteTo(dst io.Writer) (int64, error) {
	trace.SpanFromContext(o.ctx).AddEvent("WriteTo")

	go o.waitAndClose()

	return o.stream.WriteTo(dst)
}

func (o *commandOutput) Wait() error {
	trace.SpanFromContext(o.ctx).AddEvent("Wait")

	return o.waitAndClose()
}

// waitAndClose waits for command completion and closes the write half of the reader. Most
// callers do not need to use the returned error - operations that read from o.reader
// should return the error from that instead, which in most cases should be the same error.
func (o *commandOutput) waitAndClose() error {
	// If err is not reset by waitAndCloseOnce.Do, then output has already been consumed,
	// and we raise this default error.
	err := fmt.Errorf("output has already been consumed")
	o.waitAndCloseOnce.Do(func() {
		err = o.waitAndCloseFunc()
	})
	return err
}
