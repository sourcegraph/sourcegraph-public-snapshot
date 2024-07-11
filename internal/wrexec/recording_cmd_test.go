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

func listSize(t *testing.T, store *rcache.FIFOList) int {
	t.Helper()
	size, err := store.Size()
	if err != nil {
		t.Fatalf("failed to get size of FIFOList: %s", err)
	}
	return size
}

func isRecordingForCmd(t *testing.T, recording *wrexec.RecordedCommand, cmd *osexec.Cmd) bool {
	t.Helper()
	if cmd == nil || recording == nil {
		return false
	}

	if !cmp.Equal(cmd.Dir, recording.Dir) {
		return false
	}

	if !cmp.Equal(cmd.Path, recording.Path) {
		return false
	}

	if !cmp.Equal(cmd.Args, recording.Args) {
		return false
	}

	return true

}

func isValidRecording(t *testing.T, cmd *osexec.Cmd, recording *wrexec.RecordedCommand) (bool, error) {
	t.Helper()
	if cmd == nil || recording == nil {
		return false, nil
	}

	if !isRecordingForCmd(t, recording, cmd) {
		return false, errors.Errorf("incorrect recording cmd: %s", cmp.Diff(recording, cmd))
	}

	if recording.Start.IsZero() {
		return false, errors.Errorf("recording has zero start time")
	}

	if recording.Duration == 0 {
		return false, errors.Errorf("recording has no duration")
	}

	return true, nil
}

func getFirst(t *testing.T, store *rcache.FIFOList) *wrexec.RecordedCommand {
	t.Helper()
	return getRecordingAt(t, store, 0)
}

func getRecordingAt(t *testing.T, store *rcache.FIFOList, idx int) *wrexec.RecordedCommand {
	t.Helper()
	data, err := store.Slice(context.Background(), idx, idx+1)
	if err != nil {
		t.Fatalf("failed to get slice from %d to %d", idx, idx+1)
	}
	if len(data) == 0 {
		return nil
	}

	var recording wrexec.RecordedCommand
	if err := json.Unmarshal(data[0], &recording); err != nil {
		t.Fatalf("failed to unmarshal recording: %v", err)
	}
	return &recording
}

func TestRecordingCmd(t *testing.T) {
	kv := rcache.SetupForTest(t)
	store := rcache.NewFIFOList(kv, wrexec.KeyPrefix, 100)
	var recordAlways wrexec.ShouldRecordFunc = func(ctx context.Context, c *osexec.Cmd) bool {
		return true
	}

	ctx := context.Background()
	t.Run("with combinedOutput", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		cmd := osexec.Command("cat", f.Name())
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, store, cmd)
		_, err := rcmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}

		recording := getFirst(t, store)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("separate FIFOList instance can read the list", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		cmd := osexec.Command("cat", f.Name())
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, store, cmd)
		_, err := rcmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}

		readingStore := rcache.NewFIFOList(kv, wrexec.KeyPrefix, 100)
		recording := getFirst(t, readingStore)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("with Run", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		cmd := osexec.Command("cat", f.Name())
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, store, cmd)
		_ = rcmd.Run()

		recording := getFirst(t, store)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("with Output", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		cmd := osexec.Command("cat", f.Name())
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, store, cmd)
		_, err := rcmd.Output()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}

		recording := getFirst(t, store)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("with Start and Wait", func(t *testing.T) {
		f := createTmpFile(t, "foobar")
		cmd := osexec.Command("cat", f.Name())
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, store, cmd)

		// We record the size so that we can see the list did not change between calls
		sizeBefore := listSize(t, store)
		err := rcmd.Start()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}

		sizeAfter := listSize(t, store)

		// Since we called Start, the recording has not completed yet. Only once we call Wait, should the recording
		// be complete. So we check that the list didn't increase by comparing the size before and after
		if sizeBefore != sizeAfter {
			t.Error("no recording should be added after call to Start")
		}

		// Wait for the cmd to complete, and consequently, the recording to exist
		err = rcmd.Wait()
		if err != nil {
			t.Fatalf("failed to wait for recorded command: %v", err)
		}

		recording := getFirst(t, store)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("with failed command", func(t *testing.T) {
		cmd := osexec.Command("which", "i-should-not-exist-DEADBEEF")
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, store, cmd)
		_, err := rcmd.Output()
		if err == nil {
			t.Fatalf("command should have failed but executed successfully: %v", err)
		}

		recording := getFirst(t, store)
		if valid, err := isValidRecording(t, cmd, recording); !valid {
			t.Error(err)
		}
	})
	t.Run("no recording with false predicate", func(t *testing.T) {
		cmd := osexec.Command("echo", "hello-world")
		noRecord := func(ctx context.Context, c *osexec.Cmd) bool { return false }
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), noRecord, store, cmd)

		sizeBefore := listSize(t, store)
		out, err := rcmd.Output()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}

		sizeAfter := listSize(t, store)
		// Our predicate, noRecord, always returns false, which means nothing will get recorded yet our command will
		// still execute
		// So the list should be the same size before and after
		if sizeBefore != sizeAfter {
			t.Errorf("no recorded should be added to the FIFOList for noRecord predicate")
		}

		// Our command should've executed, so we should have some output
		if len(out) == 0 {
			t.Error("got no output for command")
		}
	})
	t.Run("no recording with nil predicate", func(t *testing.T) {
		cmd := osexec.Command("echo", "hello-world")
		var nilRecord func(ctx context.Context, c *osexec.Cmd) bool = nil
		rcmd := wrexec.RecordingWrap(ctx, logtest.Scoped(t), nilRecord, store, cmd)

		sizeBefore := listSize(t, store)
		out, err := rcmd.Output()
		if err != nil {
			t.Fatalf("failed to execute recorded command: %v", err)
		}

		sizeAfter := listSize(t, store)
		// Our predicate, noRecord, always returns false, which means nothing will get recorded yet our command will
		// still execute
		// So the list should be the same size before and after
		if sizeBefore != sizeAfter {
			t.Errorf("no recorded should be added to the FIFOList for noRecord predicate")
		}

		// Our command should've executed, so we should have some output
		if len(out) == 0 {
			t.Error("got no output for command")
		}
	})
	t.Run("two concurrent commands have seperate recordings", func(t *testing.T) {
		f1 := createTmpFile(t, "foobar")
		f2 := createTmpFile(t, "fubar")
		cmd1 := osexec.Command("cat", f1.Name())
		cmd2 := osexec.Command("cat", f2.Name())
		rcmd1 := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, store, cmd1)
		rcmd2 := wrexec.RecordingWrap(ctx, logtest.Scoped(t), recordAlways, store, cmd2)

		size := listSize(t, store)

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
		// rcmd1 should exist in the list, since we've called wait, but we haven't called wait on rcmd2! So ...
		// the new size should differ by 1
		newSize := listSize(t, store)
		if newSize-size != 1 {
			t.Error("expected cmd 1 to be added to store")
		}
		// Wait for the cmd to complete, and consequently, the recording to exist
		err = rcmd2.Wait()
		if err != nil {
			t.Fatalf("failed to wait for recorded command: %v", err)
		}

		recording1 := getRecordingAt(t, store, 1)
		if valid, err := isValidRecording(t, cmd1, recording1); !valid {
			t.Error(err)
		}
		recording2 := getRecordingAt(t, store, 0)
		if valid, err := isValidRecording(t, cmd2, recording2); !valid {
			t.Error(err)
		}

		if cmp.Equal(recording1, recording2) {
			t.Error("expected recording 1 and recording 2 to be different, but they're equal")
		}
	})
}
