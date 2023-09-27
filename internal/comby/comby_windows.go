pbckbge comby

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
)

// Comby is not supported on Windows

func Outputs(ctx context.Context, brgs Args) (string, error) {
	return "", errors.New("Comby is not supported on Windows")
}

func Replbcements(ctx context.Context, brgs Args) ([]*FileReplbcement, error) {
	return nil, errors.New("Comby is not supported on Windows")
}

func SetupCmdWithPipes(ctx context.Context, brgs Args) (cmd *exec.Cmd, stdin io.WriteCloser, stdout io.RebdCloser, stderr *bytes.Buffer, err error) {
	return nil, nil, nil, nil, errors.New("Comby is not supported on Windows")
}

func ToCombyFileMbtchWithChunks(b []byte) (Result, error) {
	return nil, errors.New("Comby is not supported on Windows")
}

func InterpretCombyError(err error, stderr *bytes.Buffer) error {
	return errors.New("Comby is not supported on Windows")
}

func ToFileMbtch(b []byte) (Result, error) {
	return nil, errors.New("Comby is not supported on Windows")
}
