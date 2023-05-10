package main

import (
	"bytes"
	"context"
	"flag"
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
var AppCmd *App

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

func (app *App) Shutdown() error {
	return app.Cmd.Process.Kill()
}

func (app *App) DumpOutput(w io.Writer) error {
	fmt.Fprintln(w, "Stdout:")
	if _, err := io.Copy(w, app.stdOut); err != nil {
		return err
	}
	fmt.Fprintln(w, "Stderr:")
	if _, err := io.Copy(w, app.stdErr); err != nil {
		return err
	}
	return nil
}

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
	t.Logf(msg, parts...)
}

func waitTillAvailable(t *testing.T, timeout time.Duration) error {
	t.Helper()
	ticker := time.NewTicker(5 * time.Second)
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
	flag.StringVar(&AppBinaryPath, "appBinaryPath", "", "Path to Sourcegraph App binary when not running inside Bazel")
	flag.Parse()
	// If we're running in bazel we should resolve the arg to a binary path
	if os.Getenv("BAZEL") == "true" {
		if p, found := bazel.FindBinary("", "sourcegraph"); !found {
			fmt.Fprintf(os.Stderr, "failed to find 'sourcegraph' binary inside bazel from path")
			os.Exit(1)
		} else {
			AppBinaryPath = p
		}
	}
	cmd, err := startApp(context.Background(), AppBinaryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start sourcegraph binary: %s", err)
		os.Exit(1)
	}
	AppCmd = cmd
	exit := m.Run()
	if exit != 0 {
		AppCmd.DumpOutput(os.Stdout)
	}
	if err := AppCmd.Shutdown(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to shutdown sourcegraph binary: %s", err)
	}
	os.Exit(exit)
}

func TestTauriSigninURL(t *testing.T) {
	if err := waitTillAvailable(t, 60*time.Second); err != nil {
		t.Fatalf("app failed to start up in 30 seconds: %v", err)
	}

	var buf = bytes.NewBuffer(nil)
	t.Logf("dumping app output")
	AppCmd.DumpOutput(buf)

	signinURL := ""
	for _, line := range strings.Split(buf.String(), "\n") {
		if strings.HasPrefix(line, "tauri:sign-in-url") {
			parts := strings.Split(line, " ")
			signinURL = parts[1]
		}
	}

	t.Logf("found signin url: %s", signinURL)
	// sign in
	maxRetry := 5
	var statusCode int
	var signinErr error
	for maxRetry > 0 {
		resp, err := http.Get(signinURL)
		statusCode = resp.StatusCode
		if err == nil && statusCode == 200 {
			t.Log("200 status - we signed in!")
			break
		} else if err != nil {
			maxRetry -= 1
			signinErr = err
		}
	}

	if signinErr != nil {
		t.Errorf("failed to sign in with %q: %v", signinURL, signinErr)
	}
}
