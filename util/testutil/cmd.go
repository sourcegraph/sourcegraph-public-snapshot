package testutil

import (
	"bytes"
	"fmt"
	"os/exec"
)

// RunCmd runs cmd in dir and wraps non-nil errors with contextual
// information.
func RunCmd(cmd *exec.Cmd, dir string) error {
	cmd.Dir = dir

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return wrapExecErr(err, cmd, buf.String())
	}
	return nil
}

func wrapExecErr(err error, cmd *exec.Cmd, output string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("exec %v in %s failed (%s); output follows\n\n%s", cmd.Args, cmd.Dir, err, output)
}
