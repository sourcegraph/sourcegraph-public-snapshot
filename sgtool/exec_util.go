package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func execCmd(cmd *exec.Cmd) error {
	printCmd("%s", strings.Join(cmd.Args, " "))
	start := time.Now()

	var out io.Writer
	var buf *bytes.Buffer
	log.Println()
	if globalOpts.Verbose {
		out = os.Stderr
	} else {
		buf = new(bytes.Buffer)
		out = buf
	}
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Run(); err != nil {
		log.Println()
		if buf == nil {
			printFailure("command failed (%s) with output (see above)", err)
		} else {
			printFailure("command failed (%s) with output:\n%s\n", err, buf.Bytes())
		}
		return err
	}

	if globalOpts.Verbose {
		log.Println()
	}
	log.Printf(fade("[%s]\n"), time.Since(start)/time.Millisecond*time.Millisecond)
	return nil
}

func cmdOutput(prog string, arg ...string) (string, error) {
	cmd := exec.Command(prog, arg...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("command %q failed: %s", cmd.Args, err)
	}
	return string(out), nil
}

func printCmd(format string, a ...interface{}) {
	log.Printf(green("▶ ")+format, a...)
}

func printFailure(format string, a ...interface{}) {
	log.Printf(red("▶ ")+format, a...)
}

// overrideEnv copies all of the current environment variables to cmd,
// except for the named variable. If present, it overwrites its value
// with the provided value; otherwise it sets to the provided value.
func overrideEnv(cmd *exec.Cmd, name, value string) {
	for _, s := range os.Environ() {
		if !strings.HasPrefix(s, name+"=") {
			cmd.Env = append(cmd.Env, s)
		}
	}
	cmd.Env = append(cmd.Env, name+"="+value)
}

func green(s string) string {
	return "\x1b[32m" + s + "\x1b[0m"
}

func red(s string) string {
	return "\x1b[31m" + s + "\x1b[0m"
}

func fade(s string) string {
	return "\x1b[30;1m" + s + "\x1b[0m"
}

// requireCmds returns an error if any of the dependencies are unavailable.
// A dependency is a binary name that must be found in PATH.
func requireCmds(cmds ...string) error {
	for _, cmd := range cmds {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("missing dependency %q: %v", cmd, err)
		}
	}
	return nil
}
