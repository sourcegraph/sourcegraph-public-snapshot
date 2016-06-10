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
	"time"
)

var (
	username = flag.String("username", "testuser1", "the username to use for src operations")
	password = flag.String("password", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "the password to use for src operations")
	email    = flag.String("email", "email@email.email", "the email to use for src operations")

	verbose = true

	asyncChildProcs   []*exec.Cmd
	asyncChildProcsMu sync.Mutex
)

func main() {
	flag.Parse()
	if *username == "" || *password == "" {
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
	// Set up temporary dirs/files
	if sgpathDir, err := ioutil.TempDir("", "sgpath"); err == nil {
		os.Setenv("SGPATH", sgpathDir)
		defer os.RemoveAll(sgpathDir)
	} else {
		return err
	}
	if srcAuthFile, err := ioutil.TempFile("", "src-auth"); err == nil {
		if err := srcAuthFile.Close(); err != nil {
			return err
		}
		if err := os.Remove(srcAuthFile.Name()); err != nil {
			return err
		}
		os.Setenv("SRC_AUTH_FILE", srcAuthFile.Name())
		defer os.Remove(srcAuthFile.Name())
	} else {
		return err
	}

	os.Setenv("SG_USERNAME", *username)
	os.Setenv("SG_PASSWORD", *password)
	os.Setenv("SG_EMAIL", *email)

	// launch local server
	server, err := async(`src serve --id-key=$SGPATH/id.pem`)
	if err != nil {
		return err
	}
	defer server.Process.Signal(os.Interrupt)
	start := time.Now()
	for {
		if _, err := os.Stat(os.ExpandEnv("$SGPATH/id.pem")); err == nil {
			break
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
		if time.Now().Sub(start) > 5*time.Second {
			return fmt.Errorf("timeout after %v waiting for $SGPATH/id.pem to be created", 5*time.Second)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// create user on instance (will be granted admin access)
	if err := ce(`src --endpoint=http://localhost:3080 user create $SG_USERNAME $SG_PASSWORD $SG_EMAIL`); err != nil {
		return err
	}

	// authenticate with local instance
	if err := ce(`src --endpoint=http://localhost:3080 login`); err != nil {
		return err
	}

	c(`rm -rf $SGPATH/repos/testrepo`)
	c(`src repo create testrepo`)

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
	if err := ce(cmdStr, args...); err != nil {
		panic(fmt.Sprintf("Error running %q: %s\n", expCmdStr, err))
	}
}

func ce(cmdStr string, args ...interface{}) error {
	expCmdStr := fmt.Sprintf(cmdStr, args...)
	if verbose {
		fmt.Printf("$ %s\n", expCmdStr)
	}
	return newCmd(cmdStr, args...).Run()
}

// noErr runs the specified command. If the error code is non-zero,
// then it will print the combined stdout and stderr of the process
// and panic.
func noErr(cmdStr string, args ...interface{}) {
	cmd := newCmd(cmdStr, args...)
	var stdouterr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdouterr, &stdouterr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "unexpected error running %s, output: %s", cmdStr, stdouterr.String())
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
