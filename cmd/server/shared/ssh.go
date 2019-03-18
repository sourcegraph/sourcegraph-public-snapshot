package shared

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// copySSH will copy the files at /etc/sourcegraph/ssh and put them into
// ~/.ssh
func copySSH() error {
	from := filepath.Join(os.Getenv("CONFIG_DIR"), "ssh")
	fi, err := os.Stat(from)
	if err != nil {
		if os.IsNotExist(err) {
			if verbose {
				log.Printf("%s does not exist, so only repos that do not require SSH will be accessible.", from)
			}
			return nil
		}
		return errors.Wrap(err, "failed to setup SSH auth")
	}
	if !fi.IsDir() {
		return errors.Errorf("%s is not a directory", from)
	}

	// Easiest way to recursive copy and update perm is via shell
	to := os.ExpandEnv("$HOME/.ssh")
	e := execer{}
	e.Command("cp", "-r", from+"/", to)
	e.Command("find", to, "-type", "f", "-exec", "chmod", "600", "{}", ";")
	e.Command("find", to, "-type", "d", "-exec", "chmod", "700", "{}", ";")
	return e.Error()
}

// execer wraps exec.Command, but acts like "set -x". If a command fails, all
// future commands will return the original error.
type execer struct {
	// Out if set will write the command, stdout and stderr to it
	Out io.Writer
	err error
}

// Command creates an exec.Command connected to stdout/stderr and runs it.
func (e *execer) Command(name string, arg ...string) {
	if e.err != nil {
		return
	}
	cmd := exec.Command(name, arg...)
	if verbose {
		log.Printf("$ %s %s", name, strings.Join(arg, " "))
	}
	if e.Out != nil {
		e.Out.Write([]byte(fmt.Sprintf("\n$ %s %s\n", name, strings.Join(arg, " "))))
		cmd.Stdout = e.Out
		cmd.Stderr = e.Out
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	e.err = cmd.Run()
}

// Error returns the first error encountered.
func (e execer) Error() error {
	return e.err
}
