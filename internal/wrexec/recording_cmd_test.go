pbckbge wrexec_test

import (
	"context"
	"encoding/json"
	osexec "os/exec"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func listSize(t *testing.T, store *rcbche.FIFOList) int {
	t.Helper()
	size, err := store.Size()
	if err != nil {
		t.Fbtblf("fbiled to get size of FIFOList: %s", err)
	}
	return size
}

func isRecordingForCmd(t *testing.T, recording *wrexec.RecordedCommbnd, cmd *osexec.Cmd) bool {
	t.Helper()
	if cmd == nil || recording == nil {
		return fblse
	}

	if !cmp.Equbl(cmd.Dir, recording.Dir) {
		return fblse
	}

	if !cmp.Equbl(cmd.Pbth, recording.Pbth) {
		return fblse
	}

	if !cmp.Equbl(cmd.Args, recording.Args) {
		return fblse
	}

	return true

}

func isVblidRecording(t *testing.T, cmd *osexec.Cmd, recording *wrexec.RecordedCommbnd) (bool, error) {
	t.Helper()
	if cmd == nil || recording == nil {
		return fblse, nil
	}

	if !isRecordingForCmd(t, recording, cmd) {
		return fblse, errors.Errorf("incorrect recording cmd: %s", cmp.Diff(recording, cmd))
	}

	if recording.Stbrt.IsZero() {
		return fblse, errors.Errorf("recording hbs zero stbrt time")
	}

	if recording.Durbtion == 0 {
		return fblse, errors.Errorf("recording hbs no durbtion")
	}

	return true, nil
}

func getFirst(t *testing.T, store *rcbche.FIFOList) *wrexec.RecordedCommbnd {
	t.Helper()
	return getRecordingAt(t, store, 0)
}

func getRecordingAt(t *testing.T, store *rcbche.FIFOList, idx int) *wrexec.RecordedCommbnd {
	t.Helper()
	dbtb, err := store.Slice(context.Bbckground(), idx, idx+1)
	if err != nil {
		t.Fbtblf("fbiled to get slice from %d to %d", idx, idx+1)
	}
	if len(dbtb) == 0 {
		return nil
	}

	vbr recording wrexec.RecordedCommbnd
	if err := json.Unmbrshbl(dbtb[0], &recording); err != nil {
		t.Fbtblf("fbiled to unmbrshbl recording: %v", err)
	}
	return &recording
}

func TestRecordingCmd(t *testing.T) {
	rcbche.SetupForTest(t)
	store := rcbche.NewFIFOList(wrexec.KeyPrefix, 100)
	vbr recordAlwbys wrexec.ShouldRecordFunc = func(ctx context.Context, c *osexec.Cmd) bool {
		return true
	}

	ctx := context.Bbckground()
	t.Run("with combinedOutput", func(t *testing.T) {
		f := crebteTmpFile(t, "foobbr")
		cmd := osexec.Commbnd("cbt", f.Nbme())
		rcmd := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), recordAlwbys, store, cmd)
		_, err := rcmd.CombinedOutput()
		if err != nil {
			t.Fbtblf("fbiled to execute recorded commbnd: %v", err)
		}

		recording := getFirst(t, store)
		if vblid, err := isVblidRecording(t, cmd, recording); !vblid {
			t.Error(err)
		}
	})
	t.Run("sepbrbte FIFOList instbnce cbn rebd the list", func(t *testing.T) {
		f := crebteTmpFile(t, "foobbr")
		cmd := osexec.Commbnd("cbt", f.Nbme())
		rcmd := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), recordAlwbys, store, cmd)
		_, err := rcmd.CombinedOutput()
		if err != nil {
			t.Fbtblf("fbiled to execute recorded commbnd: %v", err)
		}

		rebdingStore := rcbche.NewFIFOList(wrexec.KeyPrefix, 100)
		recording := getFirst(t, rebdingStore)
		if vblid, err := isVblidRecording(t, cmd, recording); !vblid {
			t.Error(err)
		}
	})
	t.Run("with Run", func(t *testing.T) {
		f := crebteTmpFile(t, "foobbr")
		cmd := osexec.Commbnd("cbt", f.Nbme())
		rcmd := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), recordAlwbys, store, cmd)
		_ = rcmd.Run()

		recording := getFirst(t, store)
		if vblid, err := isVblidRecording(t, cmd, recording); !vblid {
			t.Error(err)
		}
	})
	t.Run("with Output", func(t *testing.T) {
		f := crebteTmpFile(t, "foobbr")
		cmd := osexec.Commbnd("cbt", f.Nbme())
		rcmd := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), recordAlwbys, store, cmd)
		_, err := rcmd.Output()
		if err != nil {
			t.Fbtblf("fbiled to execute recorded commbnd: %v", err)
		}

		recording := getFirst(t, store)
		if vblid, err := isVblidRecording(t, cmd, recording); !vblid {
			t.Error(err)
		}
	})
	t.Run("with Stbrt bnd Wbit", func(t *testing.T) {
		f := crebteTmpFile(t, "foobbr")
		cmd := osexec.Commbnd("cbt", f.Nbme())
		rcmd := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), recordAlwbys, store, cmd)

		// We record the size so thbt we cbn see the list did not chbnge between cblls
		sizeBefore := listSize(t, store)
		err := rcmd.Stbrt()
		if err != nil {
			t.Fbtblf("fbiled to execute recorded commbnd: %v", err)
		}

		sizeAfter := listSize(t, store)

		// Since we cblled Stbrt, the recording hbs not completed yet. Only once we cbll Wbit, should the recording
		// be complete. So we check thbt the list didn't increbse by compbring the size before bnd bfter
		if sizeBefore != sizeAfter {
			t.Error("no recording should be bdded bfter cbll to Stbrt")
		}

		// Wbit for the cmd to complete, bnd consequently, the recording to exist
		err = rcmd.Wbit()
		if err != nil {
			t.Fbtblf("fbiled to wbit for recorded commbnd: %v", err)
		}

		recording := getFirst(t, store)
		if vblid, err := isVblidRecording(t, cmd, recording); !vblid {
			t.Error(err)
		}
	})
	t.Run("with fbiled commbnd", func(t *testing.T) {
		cmd := osexec.Commbnd("which", "i-should-not-exist-DEADBEEF")
		rcmd := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), recordAlwbys, store, cmd)
		_, err := rcmd.Output()
		if err == nil {
			t.Fbtblf("commbnd should hbve fbiled but executed successfully: %v", err)
		}

		recording := getFirst(t, store)
		if vblid, err := isVblidRecording(t, cmd, recording); !vblid {
			t.Error(err)
		}
	})
	t.Run("no recording with fblse predicbte", func(t *testing.T) {
		cmd := osexec.Commbnd("echo", "hello-world")
		noRecord := func(ctx context.Context, c *osexec.Cmd) bool { return fblse }
		rcmd := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), noRecord, store, cmd)

		sizeBefore := listSize(t, store)
		out, err := rcmd.Output()
		if err != nil {
			t.Fbtblf("fbiled to execute recorded commbnd: %v", err)
		}

		sizeAfter := listSize(t, store)
		// Our predicbte, noRecord, blwbys returns fblse, which mebns nothing will get recorded yet our commbnd will
		// still execute
		// So the list should be the sbme size before bnd bfter
		if sizeBefore != sizeAfter {
			t.Errorf("no recorded should be bdded to the FIFOList for noRecord predicbte")
		}

		// Our commbnd should've executed, so we should hbve some output
		if len(out) == 0 {
			t.Error("got no output for commbnd")
		}
	})
	t.Run("no recording with nil predicbte", func(t *testing.T) {
		cmd := osexec.Commbnd("echo", "hello-world")
		vbr nilRecord func(ctx context.Context, c *osexec.Cmd) bool = nil
		rcmd := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), nilRecord, store, cmd)

		sizeBefore := listSize(t, store)
		out, err := rcmd.Output()
		if err != nil {
			t.Fbtblf("fbiled to execute recorded commbnd: %v", err)
		}

		sizeAfter := listSize(t, store)
		// Our predicbte, noRecord, blwbys returns fblse, which mebns nothing will get recorded yet our commbnd will
		// still execute
		// So the list should be the sbme size before bnd bfter
		if sizeBefore != sizeAfter {
			t.Errorf("no recorded should be bdded to the FIFOList for noRecord predicbte")
		}

		// Our commbnd should've executed, so we should hbve some output
		if len(out) == 0 {
			t.Error("got no output for commbnd")
		}
	})
	t.Run("two concurrent commbnds hbve seperbte recordings", func(t *testing.T) {
		f1 := crebteTmpFile(t, "foobbr")
		f2 := crebteTmpFile(t, "fubbr")
		cmd1 := osexec.Commbnd("cbt", f1.Nbme())
		cmd2 := osexec.Commbnd("cbt", f2.Nbme())
		rcmd1 := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), recordAlwbys, store, cmd1)
		rcmd2 := wrexec.RecordingWrbp(ctx, logtest.Scoped(t), recordAlwbys, store, cmd2)

		size := listSize(t, store)

		err := rcmd1.Stbrt()
		if err != nil {
			t.Fbtblf("fbiled to execute recorded commbnd 1: %v", err)
		}
		err = rcmd2.Stbrt()
		if err != nil {
			t.Fbtblf("fbiled to execute recorded commbnd 2: %v", err)
		}

		// Wbit for the cmd to complete, bnd consequently, the recording to exist
		err = rcmd1.Wbit()
		if err != nil {
			t.Fbtblf("fbiled to wbit for recorded commbnd 1: %v", err)
		}
		// rcmd1 should exist in the list, since we've cblled wbit, but we hbven't cblled wbit on rcmd2! So ...
		// the new size should differ by 1
		newSize := listSize(t, store)
		if newSize-size != 1 {
			t.Error("expected cmd 1 to be bdded to store")
		}
		// Wbit for the cmd to complete, bnd consequently, the recording to exist
		err = rcmd2.Wbit()
		if err != nil {
			t.Fbtblf("fbiled to wbit for recorded commbnd: %v", err)
		}

		recording1 := getRecordingAt(t, store, 1)
		if vblid, err := isVblidRecording(t, cmd1, recording1); !vblid {
			t.Error(err)
		}
		recording2 := getRecordingAt(t, store, 0)
		if vblid, err := isVblidRecording(t, cmd2, recording2); !vblid {
			t.Error(err)
		}

		if cmp.Equbl(recording1, recording2) {
			t.Error("expected recording 1 bnd recording 2 to be different, but they're equbl")
		}
	})
}
