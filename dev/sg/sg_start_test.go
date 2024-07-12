package main

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output/outputtest"
)

func TestStartCommandSet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buf := useOutputBuffer(t)

	commandSet := &sgconf.Commandset{Name: "test-set", Commands: []string{"test-cmd-1"}}
	command := &run.Command{
		Config: run.SGConfigCommandOptions{
			Name: "test-cmd-1",
		},
		Install: "echo 'booting up horsegraph'",
		Cmd:     "echo 'horsegraph booted up. mount your horse.' && echo 'quitting. not horsing around anymore.'",
	}

	testConf := &sgconf.Config{
		Commands:    map[string]*run.Command{"test-cmd-1": command},
		Commandsets: map[string]*sgconf.Commandset{"test-set": commandSet},
	}

	args := StartArgs{
		CommandSet: "test-set",
	}
	cmds, _ := args.toCommands(testConf)

	if err := cmds.start(ctx); err != nil {
		t.Errorf("failed to start: %s", err)
	}

	expectOutput(t, buf, []string{
		"",
		"💡 Installing 1 commands...",
		"",
		"test-cmd-1 installed",
		"✅ 1/1 commands installed  ███████████████████████████████████████████████  100%",
		"",
		"✅ Everything installed! Booting up the system!",
		"",
		"Starting 1 cmds",
		"Running test-cmd-1...",
		"[     test-cmd-1] horsegraph booted up. mount your horse.",
		"[     test-cmd-1] quitting. not horsing around anymore.",
		"test-cmd-1 exited without error",
	})
}

func TestStartCommandSet_InstallError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buf := useOutputBuffer(t)

	commandSet := &sgconf.Commandset{Name: "test-set", Commands: []string{"test-cmd-1"}}
	command := &run.Command{
		Config: run.SGConfigCommandOptions{
			Name: "test-cmd-1",
		},
		Install: "echo 'booting up horsegraph' && exit 1",
		Cmd:     "echo 'never appears'",
	}

	testConf := &sgconf.Config{
		Commands:    map[string]*run.Command{"test-cmd-1": command},
		Commandsets: map[string]*sgconf.Commandset{"test-set": commandSet},
	}

	args := StartArgs{
		CommandSet: "test-set",
	}
	cmds, err := args.toCommands(testConf)
	if err != nil {
		t.Errorf("unexected error constructing commands: %s", err)
	}

	err = cmds.start(ctx)
	if err == nil {
		t.Fatalf("err is nil unexpectedly")
	}
	if !strings.Contains(err.Error(), "failed to run test-cmd-1") {
		t.Errorf("err contains wrong message: %s", err.Error())
	}

	expectOutput(t, buf, []string{
		"",
		"💡 Installing 1 commands...",
		"--------------------------------------------------------------------------------",
		`Failed to build test-cmd-1: 'bash -c echo 'booting up horsegraph' && exit 1' failed with "exit status 1":`,
		"booting up horsegraph",
		"--------------------------------------------------------------------------------",
	})
}

func useOutputBuffer(t *testing.T) *outputtest.Buffer {
	t.Helper()

	buf := &outputtest.Buffer{}

	oldStdout := std.Out
	std.Out = std.NewFixedOutput(buf, true)
	t.Cleanup(func() { std.Out = oldStdout })

	return buf
}

func expectOutput(t *testing.T, buf *outputtest.Buffer, want []string) {
	t.Helper()

	// TODO: DINF-104, find out why we need this sleep. Basically, when running in tests,
	// for some reason, even though cmds.start() returned, it hasn't finished writing
	// the outputs in rare cases (6 out out 100 runs on my machine).
	time.Sleep(300 * time.Millisecond)
	have := buf.Lines()

	// TODO: See DINF-104, without this, we can see the cmd output printed out after
	// the exit message. For now, and to prevent flakes, we'll keep the sort.
	sort.Strings(want)
	sort.Strings(have)
	if !cmp.Equal(want, have) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, have))
	}
}
