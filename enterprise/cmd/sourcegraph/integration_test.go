package main

import (
	"bufio"
	"bytes"
	"context"
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

var AppBinaryPath string

type App struct {
	Cmd    *exec.Cmd
	stdOut io.ReadCloser
	stdErr io.ReadCloser
}

func (app *App) StdOut() (string, error) {
	data, err := io.ReadAll(app.stdOut)
	return string(data), err
}

func (app *App) StdErr() (string, error) {
	data, err := io.ReadAll(app.stdErr)
	return string(data), err
}

var appCmd *App = nil

func startApp(ctx context.Context, path string) (*App, error) {
	testDir := bazel.TestTmpDir()

	args := fmt.Sprintf("--cacheDir %s --configDir %s", testDir, testDir)
	cmd := exec.CommandContext(ctx, path, strings.Split(args, " ")...)
	cmd.Env = append(cmd.Env, "USE_EMBEDDED_POSTGRESQL=1")
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

	return &App{
		Cmd:    cmd,
		stdOut: outPipe,
		stdErr: errPipe,
	}, nil
}

func debugf(t *testing.T, msg string, parts ...any) {
	if os.Getenv("DEBUG") != "1" {
		return
	}
	t.Logf(msg, parts...)
}

func waitTillAvailable(t *testing.T, timeout time.Duration) error {
	t.Helper()
	ticker := time.NewTicker(1 * time.Second)
	timer := time.NewTimer(timeout)

	var lastErr error = nil
	for {
		select {
		case <-ticker.C:
			{
				debugf(t, "Checking if app is up")
				resp, err := http.Get("http://localhost:3080/")

				if err == nil && resp.StatusCode == 200 {
					debugf(t, "status code 200 app is up")
					return nil
				} else if err != nil {
					debugf(t, "check err: %s", err)
					lastErr = err
				}
				if resp != nil {
					debugf(t, "status code: %d", resp.StatusCode)
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

	AppBinaryPath = appPath
	exit := m.Run()
	os.Exit(exit)
}

func TestTauriSigninURL(t *testing.T) {
	cmd, err := startApp(context.Background(), AppBinaryPath)
	if err != nil {
		t.Fatalf("failed to start sourcegraph binary: %s", err)
	}
	if err := waitTillAvailable(t, 30*time.Second); err != nil {
		debugf(t, "app failed to start up in 30 seconds: %v", err)
	}
	// Kill the app since we don't want it to continually output things
	cmd.Cmd.Process.Kill()
	// Since the app was up, the tauri signin url should have been printed
	stdOut, err := cmd.StdOut()
	if err != nil {
		t.Fatalf("failed to read command standard out: %s", err)
	}
	stdErr, err := cmd.StdErr()
	if err != nil {
		t.Fatalf("failed to read command standard err: %s", err)
	}

	var signInURL string
	r := bufio.NewReader(bytes.NewBufferString(stdErr))
	for {
		line, err := r.ReadString('\n')
		debugf(t, "Out=%s", line)
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("failed while reading the output of app: %s", err)
			break
		}
		if strings.Contains(line, "tauri:sign-in-url") {
			signInURL = line
			break
		}
	}

	if signInURL == "" {
		t.Fatalf("tauri:sign-in-url not found after 30 secs of app startup\nStdOut:%s", stdOut)
	}
}
