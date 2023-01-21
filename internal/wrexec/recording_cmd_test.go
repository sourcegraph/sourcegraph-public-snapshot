package wrexec_test

import (
	"context"
	"encoding/json"
	osexec "os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func isValidRecording(t *testing.T, cmd *osexec.Cmd, recording *wrexec.RecordedCommand) (bool, error) {
	t.Helper()

	if !cmp.Equal(cmd.Dir, recording.Dir) {
		return false, errors.Errorf("recording and command Dir differ, got %s wanted %s", recording.Dir, cmd.Dir)
	}

	if !cmp.Equal(cmd.Path, recording.Path) {
		return false, errors.Errorf("recording and command Path differ, got %s wanted %s", recording.Path, cmd.Path)
	}

	if diff := cmp.Diff(cmd.Args, recording.Args); diff != "" {
		return false, errors.Errorf("recording and command args differ: %s", diff)
	}

	if recording.Start.IsZero() {
		return false, errors.Errorf("recording has zero start time")
	}

	if recording.Duration == 0 {
		return false, errors.Errorf("recording has no duration")
	}

	return true, nil
}

func getRecording(t *testing.T, store *rcache.Cache, rcmd *wrexec.RecordingCmd) *wrexec.RecordedCommand {
	t.Helper()
	data, ok := store.Get(rcmd.Key())
	if !ok {
		t.Errorf("expected key %q to exist in redis but it was not found", rcmd.Key())
	}

	var recording wrexec.RecordedCommand
	if err := json.Unmarshal(data, &recording); err != nil {
		t.Fatalf("failed to unmarshal recording: %v", err)
	}
	return &recording
}

func TestRecordingCmd(t *testing.T) {
	rcache.SetupForTest(t)
	store := rcache.New(wrexec.KeyPrefix)
	var recordAlways wrexec.ShouldRecordFunc = func(ctx context.Context, c *osexec.Cmd) bool {
		return true
	}

	ctx := context.Background()
	t.Run("with combinedOutput", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		cmd := osexec.Command("md5sum", "-b", f.Name())
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, cmd)
		_, err := rcmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}

		recording := getRecording(t, store, rcmd)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("with Run", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		cmd := osexec.Command("md5sum", "-b", f.Name())
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, cmd)
		rcmd.Run()

		recording := getRecording(t, store, rcmd)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("with Output", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		cmd := osexec.Command("md5sum", "-b", f.Name())
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, cmd)
		_, err := rcmd.Output()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}

		recording := getRecording(t, store, rcmd)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("with Start and Wait", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		cmd := osexec.Command("md5sum", "-b", f.Name())
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, cmd)
		err := rcmd.Start()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}

		// Since we called Start, the recording has not completed yet. Only once we call Wait, should the recording
		// be complete. So we check that the recording does not exist in redis
		_, ok := store.Get(rcmd.Key())
		if ok {
			t.Errorf("expected key %q to NOT exist in redis", rcmd.Key())
		}

		// Wait for the cmd to complete, and consequently, the recording to exist
		err = rcmd.Wait()
		if err != nil {
			t.Fatalf("failed to wait for recorded command: %v", err)
		}

		recording := getRecording(t, store, rcmd)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("with failed command", func(t *testing.T) {
		cmd := osexec.Command("which", "i-should-not-exist-DEADBEEF")
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, cmd)
		_, err := rcmd.Output()
		if err == nil {
			t.Fatalf("command should have failed but executed successfully: %v", err)
		}

		recording := getRecording(t, store, rcmd)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("no recording with false predicate", func(t *testing.T) {
		cmd := osexec.Command("echo", "hello-world")
		noRecord := func(ctx context.Context, c *osexec.Cmd) bool { return false }
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), noRecord, cmd)
		out, err := rcmd.Output()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}
		// Our predicate, noRecord, always returns false, which means nothing will get recorded yet our command will
		// still execute
		// So we shouldn't get a key now
		_, ok := store.Get(rcmd.Key())
		if ok {
			t.Errorf("got %q key, but expected no key for a recording", rcmd.Key())
		}

		// Our command should've executed, so we should have some output
		if len(out) == 0 {
			t.Error("got no output for command")
		}
	})
	t.Run("two concurrent commands have seperate recordings", func(t *testing.T) {
		f1 := createTmpFile(t, "foobar")
		f2 := createTmpFile(t, "fubar")
		cmd1 := osexec.Command("md5sum", "-b", f1.Name())
		cmd2 := osexec.Command("md5sum", "-b", f2.Name())
		rcmd1 := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, cmd1)
		rcmd2 := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, cmd2)
		err := rcmd1.Start()
		if err != nil {
			t.Fatalf("failed to execute recorded command 1: %v", err)
		}
		err = rcmd2.Start()
		if err != nil {
			t.Fatalf("failed to execute recorded command 2: %v", err)
		}

		// Wait for the cmd to complete, and consequently, the recording to exist
		err = rcmd1.Wait()
		if err != nil {
			t.Fatalf("failed to wait for recorded command 1: %v", err)
		}
		// rcmd1 should exist, since we've called wait, but we haven't called wait on rcmd2! So ...
		// rcmd2 key should not exist
		_, ok := store.Get(rcmd2.Key())
		if ok {
			t.Errorf("got %q key, but expected no key for a recording 2", rcmd2.Key())
		}
		// Wait for the cmd to complete, and consequently, the recording to exist
		err = rcmd2.Wait()
		if err != nil {
			t.Fatalf("failed to wait for recorded command: %v", err)
		}

		recording1 := getRecording(t, store, rcmd1)
		if valid, err := isValidRecording(t, cmd1, recording1); !valid {
			t.Error(err)
		}
		recording2 := getRecording(t, store, rcmd2)
		if valid, err := isValidRecording(t, cmd2, recording2); !valid {
			t.Error(err)
		}

		if cmp.Equal(recording1, recording2) {
			t.Error("expected recording 1 and recording 2 to be different, but they're equal")
		}
	})

}
