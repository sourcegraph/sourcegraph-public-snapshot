package gitutil

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func checkArgSafety(rev string) error {
	if strings.HasPrefix(rev, "-") {
		return fmt.Errorf("invalid git revision (can't start with '-'): %q", rev)
	}
	return nil
}

func execGitCommand(dir string, args ...string) (out string, err error) {
	outBytes, err := execCustomGitCommand(dir, nil, args...)
	return string(bytes.TrimSpace(outBytes)), err
}

func execCustomGitCommand(dir string, f func(*exec.Cmd), args ...string) (out []byte, err error) {
	// t0 := time.Now()
	// if *trace {
	// 	defer func() {
	// 		d := time.Since(t0)
	// 		b := uint(len(out))
	// 		fmt.Fprintf(os.Stderr, "# git %s -> %s, %d bytes\n", strings.Join(args, " "), d, b)
	// 		recordExecInfo("git "+strings.Join(args, " "), d, b)
	// 		recordExecInfo("git "+args[0]+" (ALL)", d, b)
	// 	}()
	// }

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stderr = &buf
	if f != nil {
		f(cmd)
	}
	out, err = cmd.Output()
	stderr := buf.Bytes()
	if err != nil {
		return nil, gitError(args, err, stderr)
	}
	if len(stderr) != 0 {
		return nil, gitError(args, errors.New("unexpected output on stderr"), stderr)
	}
	return out, nil
}

func gitError(args []string, err error, stderr []byte) error {
	return fmt.Errorf("command failed: git %s: %s\n%s", strings.Join(args, " "), err, stderr)
}
