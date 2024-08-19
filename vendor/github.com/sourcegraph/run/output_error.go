package run

import (
	"io"

	"go.bobheadxi.dev/streamline/pipeline"
)

type errorOutput struct{ err error }

// NewErrorOutput creates an Output that just returns error. Useful for allowing function
// that help run Commands and want to just return an Output even if errors can happen
// before command execution.
func NewErrorOutput(err error) Output { return &errorOutput{err: err} }

func (o *errorOutput) StdErr() Output                    { return o }
func (o *errorOutput) StdOut() Output                    { return o }
func (o *errorOutput) Map(LineMap) Output                { return o }
func (o *errorOutput) Pipeline(pipeline.Pipeline) Output { return o }

func (o *errorOutput) Stream(io.Writer) error           { return o.err }
func (o *errorOutput) StreamLines(func(string)) error   { return o.err }
func (o *errorOutput) Lines() ([]string, error)         { return nil, o.err }
func (o *errorOutput) String() (string, error)          { return "", o.err }
func (o *errorOutput) JQ(string) ([]byte, error)        { return nil, o.err }
func (o *errorOutput) Read([]byte) (int, error)         { return 0, o.err }
func (o *errorOutput) WriteTo(io.Writer) (int64, error) { return 0, o.err }

func (o *errorOutput) Wait() error { return o.err }
