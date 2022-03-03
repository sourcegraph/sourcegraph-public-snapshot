package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/output/outputtest"
)

func TestStartCommandSet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buf := useOutputBuffer(t)

	commandSet := &Commandset{Name: "test-set", Commands: []string{"test-cmd-1"}}
	command := run.Command{
		Name:    "test-cmd-1",
		Install: "echo 'booting up horsegraph'",
		Cmd:     "echo 'horsegraph booted up. mount your horse.' && echo 'quitting. not horsing around anymore.'",
	}

	testConf := &Config{
		Commands:    map[string]run.Command{"test-cmd-1": command},
		Commandsets: map[string]*Commandset{"test-set": commandSet},
	}

	if err := startCommandSet(ctx, commandSet, testConf, false); err != nil {
		t.Errorf("failed to start: %s", err)
	}

	expectOutput(t, buf, []string{
		"",
		"💡 Installing 1 commands...",
		"",
		"test-cmd-1 installed",
		"✅ 1/1 commands installed  █████████████████████████████████████████████████████",
		"",
		"✅ Everything installed! Booting up the system!",
		"",
		"Running test-cmd-1...",
		"[test-cmd-1] horsegraph booted up. mount your horse.",
		"[test-cmd-1] quitting. not horsing around anymore.",
		"test-cmd-1 exited without error",
	})
}

func TestStartCommandSet_InstallError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buf := useOutputBuffer(t)

	commandSet := &Commandset{Name: "test-set", Commands: []string{"test-cmd-1"}}
	command := run.Command{
		Name:    "test-cmd-1",
		Install: "echo 'booting up horsegraph' && exit 1",
		Cmd:     "echo 'never appears'",
	}

	testConf := &Config{
		Commands:    map[string]run.Command{"test-cmd-1": command},
		Commandsets: map[string]*Commandset{"test-set": commandSet},
	}

	if err := startCommandSet(ctx, commandSet, testConf, false); err == nil {
		t.Errorf("err is nil unexpectedly")
	}

	expectOutput(t, buf, []string{
		"",
		"💡 Installing 1 commands...",
		"",
		"test-cmd-1 installed",
		"✅ 1/1 commands installed  █████████████████████████████████████████████████████",
		"",
		"✅ Everything installed! Booting up the system!",
		"",
		"--------------------------------------------------------------------------------",
		"Failed to build test-cmd-1: 'bash -c echo 'booting up horsegraph' && exit 1' failed: booting up horsegraph: exit status 1:",
		"booting up horsegraph",
		"--------------------------------------------------------------------------------",
	})
}

func useOutputBuffer(t *testing.T) *outputtest.Buffer {
	t.Helper()

	buf := &outputtest.Buffer{}
	out := output.NewOutput(buf, output.OutputOpts{
		ForceTTY:    true,
		ForceColor:  true,
		ForceHeight: 25,
		ForceWidth:  80,
		Verbose:     true,
	})

	oldStdout := stdout.Out
	stdout.Out = out
	t.Cleanup(func() { stdout.Out = oldStdout })

	return buf
}

func expectOutput(t *testing.T, buf *outputtest.Buffer, want []string) {
	t.Helper()

	have := buf.Lines()
	if !cmp.Equal(want, have) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, have))
	}
}
