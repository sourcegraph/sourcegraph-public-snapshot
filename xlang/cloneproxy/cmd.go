package main

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

func mkStdIoLSConn(ctx context.Context, name string, arg ...string) (io.ReadWriteCloser, error) {
	cmd := exec.CommandContext(ctx, name, arg...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stdin pipe for language server")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create stdout pipe for language server")
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "failed to start cmd for language server")
	}

	return &cmdRWCloser{
		Cmd:    cmd,
		Reader: stdout,
		Writer: stdin,
	}, nil
}

type cmdRWCloser struct {
	*exec.Cmd

	// Reader and Writer do not need to be Closers since they are StdoutPipe
	// and StdinPipe respectively. Both of those will be closed by Cmd.Wait.
	io.Reader
	io.Writer
}

func (c *cmdRWCloser) Close() error {
	if err := c.Cmd.Process.Kill(); err != nil {
		return errors.Wrap(err, "unable to kill process during cmdRWCloser.Close()")
	}

	if err := c.Cmd.Wait(); err != nil {
		return errors.Wrap(err, "unable to wait on cmd to finish during cmdRWCloser.Close()")
	}

	return nil
}
