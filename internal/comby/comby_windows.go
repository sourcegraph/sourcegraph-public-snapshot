package comby

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
)

// Comby is not supported on Windows

func Outputs(ctx context.Context, args Args) (string, error) {
	return "", errors.New("Comby is not supported on Windows")
}

func Replacements(ctx context.Context, args Args) ([]*FileReplacement, error) {
	return nil, errors.New("Comby is not supported on Windows")
}

func SetupCmdWithPipes(ctx context.Context, args Args) (cmd *exec.Cmd, stdin io.WriteCloser, stdout io.ReadCloser, stderr *bytes.Buffer, err error) {
	return nil, nil, nil, nil, errors.New("Comby is not supported on Windows")
}

func ToCombyFileMatchWithChunks(b []byte) (Result, error) {
	return nil, errors.New("Comby is not supported on Windows")
}

func InterpretCombyError(err error, stderr *bytes.Buffer) error {
	return errors.New("Comby is not supported on Windows")
}

func ToFileMatch(b []byte) (Result, error) {
	return nil, errors.New("Comby is not supported on Windows")
}
