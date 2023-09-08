package wrexec_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	osexec "os/exec"
)

func TestCommand(t *testing.T) {
	logger := logtest.Scoped(t)

	name := "echo"
	args := []string{"foo"}

	got := wrexec.CommandContext(context.Background(), logger, name, args...).Cmd
	want := osexec.Command(name, args...)

	if diff := cmp.Diff(want.Path, got.Path); diff != "" {
		t.Fatal("Path", diff)
	}
	if diff := cmp.Diff(want.Args, got.Args); diff != "" {
		t.Fatal("Args", diff)
	}
	if diff := cmp.Diff(want.Environ(), got.Environ()); diff != "" {
		t.Fatal("Args", diff)
	}
	if diff := cmp.Diff(want.Dir, got.Dir); diff != "" {
		t.Fatal("Dir", diff)
	}
}

func TestWrap(t *testing.T) {
	logger := logtest.Scoped(t)

	name := "echo"
	args := []string{"foo"}

	want := osexec.Command(name, args...)
	got := wrexec.Wrap(context.Background(), logger, want).Cmd

	if diff := cmp.Diff(want.Path, got.Path); diff != "" {
		t.Fatal("Path", diff)
	}
	if diff := cmp.Diff(want.Args, got.Args); diff != "" {
		t.Fatal("Args", diff)
	}
	if diff := cmp.Diff(want.Environ(), got.Environ()); diff != "" {
		t.Fatal("Args", diff)
	}
	if diff := cmp.Diff(want.Dir, got.Dir); diff != "" {
		t.Fatal("Dir", diff)
	}
}

func TestCombinedOutput(t *testing.T) {
	logger := logtest.Scoped(t)
	f := createTmpFile(t, "foobar")
	tc := createTestCommand(context.Background(), logger, "cat", f.Name())

	want, wantErr := tc.oscmd.CombinedOutput()
	got, gotErr := tc.cmd.CombinedOutput()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(wantErr, gotErr); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("output is correctly saved", func(t *testing.T) {
		if diff := cmp.Diff(string(want), tc.cmd.GetExecutionOutput()); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("after hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.afterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestEnviron(t *testing.T) {
	logger := logtest.Scoped(t)
	tc := createTestCommand(context.Background(), logger, "echo", "foobar")

	want := tc.oscmd.Environ()
	got := tc.cmd.Environ()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks are NOT called", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("after hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.afterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestOutput(t *testing.T) {
	logger := logtest.Scoped(t)
	f := createTmpFile(t, "foobar")
	tc := createTestCommand(context.Background(), logger, "cat", f.Name())

	want, wantErr := tc.oscmd.Output()
	got, gotErr := tc.cmd.Output()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(wantErr, gotErr); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("output is correctly saved", func(t *testing.T) {
		if diff := cmp.Diff(string(want), tc.cmd.GetExecutionOutput()); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("after hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.afterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestRun(t *testing.T) {
	logger := logtest.Scoped(t)
	f := createTmpFile(t, "foobar")
	tc := createTestCommand(context.Background(), logger, "cat", f.Name())

	want := tc.oscmd.Run()
	got := tc.cmd.Run()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("after hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.afterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestStart(t *testing.T) {
	logger := logtest.Scoped(t)
	f := createTmpFile(t, "foobar")
	tc := createTestCommand(context.Background(), logger, "cat", f.Name())

	want := tc.oscmd.Start()
	got := tc.cmd.Start()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("after hooks are NOT called", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.afterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestStdoutPipe(t *testing.T) {
	logger := logtest.Scoped(t)
	f := createTmpFile(t, "foobar")
	tc := createTestCommand(context.Background(), logger, "cat", f.Name())

	wStdout, err := tc.oscmd.StdoutPipe()
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	gStdout, err := tc.cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	if err := tc.oscmd.Start(); err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	if err := tc.cmd.Start(); err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	wb, err := io.ReadAll(wStdout)
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	gb, err := io.ReadAll(gStdout)
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	t.Run("OK", func(t *testing.T) {
		if len(gb) == 0 {
			t.Fatal("expected to get some output")
		}
		if diff := cmp.Diff(string(gb), string(wb)); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("after hooks are NOT called", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.afterCounter); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("OK wait", func(t *testing.T) {
		if err := tc.oscmd.Wait(); err != nil {
			t.Fatalf("want err to be nil, got %q", err)
		}
		if err := tc.cmd.Wait(); err != nil {
			t.Fatalf("want err to be nil, got %q", err)
		}
		t.Run("after hooks are  called", func(t *testing.T) {
			if diff := cmp.Diff(1, tc.afterCounter); diff != "" {
				t.Log(diff)
				t.Fail()
			}
		})
	})
}

func TestStderrPipe(t *testing.T) {
	logger := logtest.Scoped(t)
	tc := createTestCommand(context.Background(), logger, "cat", "non-existing")

	wStderr, err := tc.oscmd.StderrPipe()
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	gStderr, err := tc.cmd.StderrPipe()
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	if err := tc.oscmd.Start(); err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	if err := tc.cmd.Start(); err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	wb, err := io.ReadAll(wStderr)
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	gb, err := io.ReadAll(gStderr)
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	t.Run("OK", func(t *testing.T) {
		if len(gb) == 0 {
			t.Fatal("expected to get some output")
		}
		if diff := cmp.Diff(string(gb), string(wb)); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("after hooks are NOT called", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.afterCounter); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("OK wait", func(t *testing.T) {
		if err := tc.oscmd.Wait(); err == nil {
			t.Fatal("want err to be not nil")
		}
		if err := tc.cmd.Wait(); err == nil {
			t.Fatal("want err to be not nil")
		}
		t.Run("after hooks are  called", func(t *testing.T) {
			if diff := cmp.Diff(1, tc.afterCounter); diff != "" {
				t.Error(diff)
			}
		})
	})
}

func TestStdinPipe(t *testing.T) {
	logger := logtest.Scoped(t)
	data := "foobar"
	tc := createTestCommand(context.Background(), logger, "cat")

	wStdin, err := tc.oscmd.StdinPipe()
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	gStdin, err := tc.cmd.StdinPipe()
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	wStdout, err := tc.oscmd.StdoutPipe()
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	gStdout, err := tc.cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	if err := tc.oscmd.Start(); err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	if err := tc.cmd.Start(); err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	_, err = io.WriteString(wStdin, data)
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	if err := wStdin.Close(); err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	_, err = io.WriteString(gStdin, data)
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	if err := gStdin.Close(); err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	wb, err := io.ReadAll(wStdout)
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}
	gb, err := io.ReadAll(gStdout)
	if err != nil {
		t.Fatalf("want err to be nil, got %q", err)
	}

	t.Run("OK", func(t *testing.T) {
		if err := tc.oscmd.Wait(); err != nil {
			t.Fatalf("want err to be nil, got %q", err)
		}
		if err := tc.cmd.Wait(); err != nil {
			t.Fatalf("want err to be nil, got %q", err)
		}

		if string(gb) == "" {
			t.Fatal("expected to get some output")
		}
		if diff := cmp.Diff(string(gb), string(wb)); diff != "" {
			t.Log(diff)
			t.Fail()
		}
	})

	t.Run("before hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("after hooks are called", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.afterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestString(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	oscmd := osexec.CommandContext(ctx, "echo", "foobar")
	cmd := wrexec.CommandContext(ctx, logger, "echo", "foobar")
	cmd1 := wrexec.Wrap(ctx, logger, oscmd)

	want := oscmd.String()
	got := cmd.String()
	got1 := cmd1.String()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal(diff)
	}
	if diff := cmp.Diff(want, got1); diff != "" {
		t.Fatal(diff)
	}
}

func TestHooks(t *testing.T) {
	name := "cat"
	logger := logtest.Scoped(t)
	ctx := context.Background()

	t.Run("all hooks are called", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		args := []string{f.Name()}
		cmd := wrexec.CommandContext(ctx, logger, name, args...)
		var a1, a2 int
		var b1, b2 int
		cmd.SetBeforeHooks(
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) error {
				l.Info("b1")
				b1++
				return nil
			},
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) error {
				l.Info("b2")
				b2++
				return nil
			},
		)
		cmd.SetAfterHooks(
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) {
				l.Info("a1")
				a1++
			},
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) {
				l.Info("a2")
				a2++
			},
		)

		if err := cmd.Run(); err != nil {
			t.Fatalf("want no errors, but got %q", err)
		}

		if a1 != 1 || a2 != 1 || b1 != 1 || b2 != 1 {
			t.Fatalf("expected all hooks to be called")
		}
	})

	t.Run("before hooks can interrupt the command", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		args := []string{f.Name()}
		cmd := wrexec.CommandContext(ctx, logger, name, args...)
		var a, b1, b2 int
		wantErr := errors.New("foobar")
		cmd.SetBeforeHooks(
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) error {
				l.Info("b1")
				b1++
				return nil
			},
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) error {
				l.Info("before hook (returning an error)")
				return wantErr
			},
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) error {
				l.Info("b2 (should not be called)")
				b2++
				return nil
			},
		)
		cmd.SetAfterHooks(func(ctx context.Context, l log.Logger, _ *osexec.Cmd) {
			l.Info("after hook (should not be called)")
			a++
		})

		if err := cmd.Run(); !errors.Is(err, wantErr) {
			t.Fatalf("want %q errors, but got %q", wantErr, err)
		}

		if a != 0 {
			t.Fatalf("expected after hook to not be called")
		}
		if b1 != 1 {
			t.Fatalf("expected first before hook to be called")
		}
		if b2 != 0 {
			t.Fatalf("expected after hook to not be called")
		}
	})

	t.Run("before hooks can update the os.exec.Cmd", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		oscmd := osexec.Command("cat", f.Name())
		cmd := wrexec.CommandContext(ctx, logger, "cat", "wrong")
		cmd.SetBeforeHooks(func(ctx context.Context, _ log.Logger, c *osexec.Cmd) error {
			// .Args[0] is going to be ignored if and only if .Path is present.
			// And the osexec.Command always set it obviously ...
			// It's really easy to miss it and to end up wondering why an argument is missing.
			c.Args = []string{c.Path, f.Name()}
			return nil
		})

		want, err := oscmd.CombinedOutput()
		if err != nil {
			t.Fatalf("want no errors, but got %q", err)
		}
		got, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("want no errors, but got %q", err)
		}

		if len(want) == 0 {
			t.Fatal("expected to get some output")
		}
		if diff := cmp.Diff(string(want), string(got)); diff != "" {
			t.Fatal(diff)
		}
	})

	t.Run("we can use context in hooks", func(t *testing.T) {
		//nolint:staticcheck
		ctx := context.WithValue(context.Background(), "my-key", 1)
		f := createTmpFile(t, "foobar")
		cmd := wrexec.CommandContext(ctx, logger, "cat", f.Name())
		cmd.SetBeforeHooks(func(ctx context.Context, _ log.Logger, _ *osexec.Cmd) error {
			want, got := 1, ctx.Value("my-key")
			if want != got {
				t.Errorf("want my-key to be 1, got %v", got)
			}
			return nil
		})
		cmd.SetAfterHooks(func(ctx context.Context, _ log.Logger, _ *osexec.Cmd) {
			want, got := 1, ctx.Value("my-key")
			if want != got {
				t.Errorf("want my-key to be 1, got %v", got)
			}
		})
	})
}

type testcase struct {
	beforeCounter int
	afterCounter  int
	oscmd         *osexec.Cmd
	cmd           *wrexec.Cmd
}

func createTmpFile(t *testing.T, content string) *os.File {
	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("foobar"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Remove(f.Name())
	})
	return f
}

func createTestCommand(ctx context.Context, logger log.Logger, name string, args ...string) *testcase {
	c := wrexec.CommandContext(ctx, logger, name, args...)
	testcase := testcase{
		oscmd: osexec.CommandContext(ctx, name, args...),
		cmd:   c,
	}
	c.SetBeforeHooks(func(ctx context.Context, l log.Logger, c *osexec.Cmd) error {
		testcase.beforeCounter++
		l.Info("before hook", log.Int("beforeCounter", testcase.beforeCounter))
		return nil
	})
	c.SetAfterHooks(func(ctx context.Context, l log.Logger, c *osexec.Cmd) {
		testcase.afterCounter++
		l.Info("after hook", log.Int("afterCounter", testcase.afterCounter))
	})
	return &testcase
}
