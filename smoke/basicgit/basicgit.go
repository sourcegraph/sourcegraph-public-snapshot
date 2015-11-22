package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
)

var (
	username = flag.String("username", "testuser", "the username to use for src operations")
	password = flag.String("password", "24a6d249fa9c0a280cfabf9d9b90eec33914e546", "the password to use for src operations")

	registeredClient            = "testserver"
	registeredClientURL         = "http://localhost:3080"
	registeredClientRedirectURL = "http://localhost:3080/login/oauth/receive"

	verbose = true

	asyncChildProcs   []*exec.Cmd
	asyncChildProcsMu sync.Mutex
)

func main() {
	flag.Parse()
	if username == "" || password == "" {
		fmt.Fprintf(os.Stderr, "username or password not specified\n")
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		for s := range c {
			fmt.Fprintf(os.Stderr, "signal %s received, killing %d child processes\n", s.String(), len(asyncChildProcs))
			for _, child := range asyncChildProcs {
				fmt.Fprintf(os.Stderr, "  killing process %d\n", child.Process.Pid)
				child.Process.Signal(s)
			}
		}
	}()

	if err := main_(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main_() error {
	c(`src login --endpoint=https://sourcegraph.com -u %s -p %s`, username, password)

	server, err := async(`src serve --allow-all-logins`)
	if err != nil {
		return err
	}
	defer server.Process.Signal(os.Interrupt)

	c(`src registered-clients create --client-name=%s --client-uri='%s' --redirect-uri='%s'`, registeredClient, registeredClientURL, registeredClientRedirectURL)

	c(`rm -rf ~/.sourcegraph/repos/testrepo`)
	c(`src --endpoint=http://localhost:3080 repo create testrepo`)

	testrepoDir, err := ioutil.TempDir("", "testrepo")
	must(err)
	defer os.RemoveAll(testrepoDir)
	func() {
		// Create and push git repository.
		origPwd := os.Getenv("PWD")
		os.Setenv("PWD", testrepoDir)
		defer os.Setenv("PWD", origPwd)

		c(`git init`)
		c(`touch emptyfile`)
		c(`git add -A .`)
		c(`git commit -m'add emptyfile'`)
		c(`git remote add origin http://%s:%s@localhost:3080/testrepo`, *username, *password)
		c(`git push -u origin master`)
	}()

	cloneDir, err := ioutil.TempDir("", "testrepo")
	log.Printf("cloning to %s", cloneDir)
	must(err)
	defer os.RemoveAll(cloneDir)
	func() {
		// Clone git repository
		origPwd := os.Getenv("PWD")
		os.Setenv("PWD", cloneDir)
		defer os.Setenv("PWD", origPwd)

		c(`git clone http://%s:%s@localhost:3080/testrepo`, *username, *password)
		noErr(`stat testrepo/emptyfile`)
	}()

	return nil
}

// must panics if the last of the arguments is an error.
func must(rets ...interface{}) {
	if len(rets) > 0 {
		if err, isErr := rets[len(rets)-1].(error); isErr {
			panic(fmt.Sprintf("fatal error: %s\n", err))
		}
	}
}

// c runs the specified command, passing through stdout and stderr.
// If the exit code is non-zero, then it will panic.
func c(cmdStr string, args ...interface{}) {
	expCmdStr := fmt.Sprintf(cmdStr, args...)
	if verbose {
		fmt.Printf("$ %s\n", expCmdStr)
	}
	if err := newCmd(cmdStr, args...).Run(); err != nil {
		panic(fmt.Sprintf("Error running %q: %s\n", expCmdStr, err))
	}
}

// noErr runs the specified command. If the error code is non-zero,
// then it will print the combined stdout and stderr of the process
// and panic.
func noErr(cmdStr string, args ...interface{}) {
	cmd := newCmd(cmdStr, args...)
	var stdouterr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdouterr, &stdouterr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "unexpected error running %s, output: %s", cmdStr, stdouterr)
		must(err)
	}
}

// newCmd create a new exec.Cmd that invokes the passed command
// (specified as format string and arguments) as a shell command.
func newCmd(cmdStr string, args ...interface{}) *exec.Cmd {
	if len(args) > 0 {
		cmdStr = fmt.Sprintf(cmdStr, args...)
	}
	cmd := exec.Command("sh", "-c", cmdStr)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if os.Getenv("PWD") != "" {
		cmd.Dir = os.Getenv("PWD")
	}
	return cmd
}

// async calls a command asynchronously. It adds the command to the
// list of asynchronous processes, so these can be killed upon
// receiving a SIGINT or SIGKILL.
func async(cmdStr string, args ...interface{}) (*exec.Cmd, error) {
	asyncChildProcsMu.Lock()
	defer asyncChildProcsMu.Unlock()

	cmd := newCmdDirect(cmdStr, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	asyncChildProcs = append(asyncChildProcs, cmd)

	return cmd, nil
}

// newCmdDirect invokes a command directly (i.e., not via `sh`). This
// is convenient in cases where you don't want an extra layer above
// the actual command you want to invoke (e.g., async process where
// you want to be able to kill the process directly, without futzing
// about with process group IDs).
func newCmdDirect(cmdStr string, args ...interface{}) *exec.Cmd {
	if len(args) > 0 {
		cmdStr = fmt.Sprintf(cmdStr, args...)
	}
	cmdStr = os.ExpandEnv(cmdStr)
	fields := strings.Fields(cmdStr)

	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if os.Getenv("PWD") != "" {
		cmd.Dir = os.Getenv("PWD")
	}
	return cmd
}
