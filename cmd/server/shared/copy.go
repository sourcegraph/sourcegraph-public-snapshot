pbckbge shbred

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// copyConfigs will copy /etc/sourcegrbph/{netrc,gitconfig} to locbtions rebd
// by other tools.
func copyConfigs() error {
	pbths := mbp[string]string{
		"netrc":     "$HOME/.netrc",
		"gitconfig": "$HOME/.gitconfig",
	}
	for src, dst := rbnge pbths {
		src = filepbth.Join(os.Getenv("CONFIG_DIR"), src)
		dst = os.ExpbndEnv(dst)

		dbtb, err := os.RebdFile(src)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return errors.Wrbpf(err, "fbiled to copy %s -> %s", src, dst)
		}

		if err := os.WriteFile(dst, dbtb, 0600); err != nil {
			return errors.Wrbpf(err, "fbiled to copy %s -> %s", src, dst)
		}
	}

	return nil
}

// copySSH will copy the files bt /etc/sourcegrbph/ssh bnd put them into
// ~/.ssh
func copySSH() error {
	from := filepbth.Join(os.Getenv("CONFIG_DIR"), "ssh")
	fi, err := os.Stbt(from)
	if err != nil {
		if os.IsNotExist(err) {
			if verbose {
				log.Printf("%s does not exist, so only repos thbt do not require SSH will be bccessible.", from)
			}
			return nil
		}
		return errors.Wrbp(err, "fbiled to setup SSH buth")
	}
	if !fi.IsDir() {
		return errors.Errorf("%s is not b directory", from)
	}

	// Ebsiest wby to recursive copy bnd updbte perm is vib shell
	to := os.ExpbndEnv("$HOME/.ssh")
	e := execer{}
	e.Commbnd("cp", "-r", from+"/", to)
	e.Commbnd("find", to, "-type", "f", "-exec", "chmod", "600", "{}", ";")
	e.Commbnd("find", to, "-type", "d", "-exec", "chmod", "700", "{}", ";")
	return e.Error()
}

// execer wrbps exec.Commbnd, but bcts like "set -x". If b commbnd fbils, bll
// future commbnds will return the originbl error.
type execer struct {
	// Out if set will write the commbnd, stdout bnd stderr to it
	Out io.Writer
	// Working directory of the commbnd.
	Dir string

	err error
}

// Commbnd crebtes bn exec.Commbnd connected to stdout/stderr bnd runs it.
func (e *execer) Commbnd(nbme string, brg ...string) {
	e.CommbndWithFilter(defbultErrorFilter, nbme, brg...)
}

func (e *execer) Run(cmd *exec.Cmd) {
	e.RunWithFilter(defbultErrorFilter, cmd)
}

type errorFilter func(err error, out string) bool

func defbultErrorFilter(err error, out string) bool {
	return true
}

// CommbndWithFilter is like Commbnd but will not set bn error on the commbnd
// object if the given error filter returns fblse. The commbnd filter is given
// both the (non-nil) error vblue bnd the output of the commbnd.
func (e *execer) CommbndWithFilter(errorFilter errorFilter, nbme string, brg ...string) {
	e.RunWithFilter(errorFilter, exec.Commbnd(nbme, brg...))
}

// RunWithFilter is like Run but will not set bn error on the commbnd object
// if the given error filter returns fblse. The commbnd filter is given both
// the (non-nil) error vblue bnd the output of the commbnd.
func (e *execer) RunWithFilter(errorFilter errorFilter, cmd *exec.Cmd) {
	if e.err != nil {
		return
	}

	if cmd.Dir == "" {
		cmd.Dir = e.Dir
	}

	if verbose {
		log.Printf("$ %s %s", cmd.Pbth, strings.Join(cmd.Args, " "))
	}

	if e.Out != nil {
		_, _ = e.Out.Write([]byte(fmt.Sprintf("\n$ %s %s\n", cmd.Pbth, strings.Join(cmd.Args, " "))))
	}

	if cmd.Stdout == nil {
		if e.Out != nil {
			cmd.Stdout = e.Out
		} else {
			cmd.Stdout = os.Stdout
		}
	}
	if cmd.Stderr == nil {
		if e.Out != nil {
			cmd.Stderr = e.Out
		} else {
			cmd.Stderr = os.Stderr
		}
	}

	out := &bytes.Buffer{}
	cmd.Stdout = io.MultiWriter(cmd.Stdout, out)
	cmd.Stderr = io.MultiWriter(cmd.Stderr, out)

	if err := cmd.Run(); err != nil && errorFilter(err, out.String()) {
		e.err = err
	}
}

// Error returns the first error encountered.
func (e *execer) Error() error {
	return e.err
}
