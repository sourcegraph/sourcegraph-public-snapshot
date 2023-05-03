package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

type AppCmdState struct {
	Cmd     *exec.Cmd
	OutPipe io.ReadCloser
	ErrPipe io.ReadCloser
}

var appCmd *AppCmdState = nil

func startApp(path string) (*AppCmdState, error) {
	args := "--cacheDir cache --configDir config"
	cmd := exec.Command(path, strings.Split(args, " ")...)
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &AppCmdState{
		Cmd:     cmd,
		OutPipe: outPipe,
		ErrPipe: errPipe,
	}, nil
}

func curl() (string, error) {
	cmd := exec.Command("curl", "-v", "http://localhost:3080")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func appIsAvailable(t *testing.T, timeout time.Duration) error {
	t.Helper()
	ticker := time.NewTicker(1 * time.Second)
	timer := time.NewTimer(timeout)

	var lastErr error = nil
	for {
		select {
		case <-ticker.C:
			{
				t.Logf("Checking if app is up")
				resp, err := http.Get("http://localhost:3080/")

				if err == nil && resp.StatusCode == 200 {
					return nil
				} else if err != nil {
					t.Logf("check err: %s", err)
					lastErr = err
				}
				if resp != nil {
					t.Logf("status code: %d", resp.StatusCode)
				}
				if o, err := curl(); err != nil {
					t.Logf("curl err: %s - %s", err, o)
				} else {
					t.Logf("curl: %s", o)
				}
			}
		case <-timer.C:
			{
				ticker.Stop()
				timer.Stop()
				return lastErr
			}
		}
	}

}

func TestMain(m *testing.M) {
	if len(os.Args) < 1 {
		fmt.Fprintln(os.Stderr, "Sourcegraph App binary path required as first argument to this test")
		os.Exit(0)
	}
	appPath := os.Args[1]

	// If we're running in bazel we should resolve the arg to a binary path
	if os.Getenv("BAZEL") == "true" {
		if p, found := bazel.FindBinary(appPath, "sourcegraph"); !found {
			fmt.Fprintf(os.Stderr, "failed to find 'sourcegraph' binary inside bazel from path: %s", appPath)
			os.Exit(1)
		} else {
			appPath = p
		}
	}
	cmd, err := startApp(appPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to sourcegraph binary: %s", err)
		os.Exit(1)
	}
	appCmd = cmd
	exit := m.Run()
	fmt.Printf("Wait: %s", cmd.Cmd.Wait())
	os.Exit(exit)
}

func TestTauriSigninURL(t *testing.T) {

	if err := appIsAvailable(t, 10*time.Second); err != nil {
		t.Logf("app failed to start up in 30 seconds: %v", err)
	}
	// now that the app is up, the tauri url should have been printed
	data, err := io.ReadAll(appCmd.ErrPipe)
	if err != nil {
		t.Fatal("failed to read app output", err)
	}

	var signInURL string
	r := bufio.NewReader(bytes.NewBuffer(data))
	for {
		line, err := r.ReadString('\n')
		t.Logf("Line=%s", line)
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("failed while reading the output of app: %s", err)
		}
		if strings.Contains(line, "tauri:sign-in-url") {
			signInURL = line
		}
	}

	t.Fatalf("Found signin url: %s", signInURL)
}
