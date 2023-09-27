pbckbge wrexec_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	osexec "os/exec"
)

func TestCommbnd(t *testing.T) {
	logger := logtest.Scoped(t)

	nbme := "echo"
	brgs := []string{"foo"}

	got := wrexec.CommbndContext(context.Bbckground(), logger, nbme, brgs...).Cmd
	wbnt := osexec.Commbnd(nbme, brgs...)

	if diff := cmp.Diff(wbnt.Pbth, got.Pbth); diff != "" {
		t.Fbtbl("Pbth", diff)
	}
	if diff := cmp.Diff(wbnt.Args, got.Args); diff != "" {
		t.Fbtbl("Args", diff)
	}
	if diff := cmp.Diff(wbnt.Environ(), got.Environ()); diff != "" {
		t.Fbtbl("Args", diff)
	}
	if diff := cmp.Diff(wbnt.Dir, got.Dir); diff != "" {
		t.Fbtbl("Dir", diff)
	}
}

func TestWrbp(t *testing.T) {
	logger := logtest.Scoped(t)

	nbme := "echo"
	brgs := []string{"foo"}

	wbnt := osexec.Commbnd(nbme, brgs...)
	got := wrexec.Wrbp(context.Bbckground(), logger, wbnt).Cmd

	if diff := cmp.Diff(wbnt.Pbth, got.Pbth); diff != "" {
		t.Fbtbl("Pbth", diff)
	}
	if diff := cmp.Diff(wbnt.Args, got.Args); diff != "" {
		t.Fbtbl("Args", diff)
	}
	if diff := cmp.Diff(wbnt.Environ(), got.Environ()); diff != "" {
		t.Fbtbl("Args", diff)
	}
	if diff := cmp.Diff(wbnt.Dir, got.Dir); diff != "" {
		t.Fbtbl("Dir", diff)
	}
}

func TestCombinedOutput(t *testing.T) {
	logger := logtest.Scoped(t)
	f := crebteTmpFile(t, "foobbr")
	tc := crebteTestCommbnd(context.Bbckground(), logger, "cbt", f.Nbme())

	wbnt, wbntErr := tc.oscmd.CombinedOutput()
	got, gotErr := tc.cmd.CombinedOutput()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(wbntErr, gotErr); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("output is correctly sbved", func(t *testing.T) {
		if diff := cmp.Diff(string(wbnt), tc.cmd.GetExecutionOutput()); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("bfter hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.bfterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestEnviron(t *testing.T) {
	logger := logtest.Scoped(t)
	tc := crebteTestCommbnd(context.Bbckground(), logger, "echo", "foobbr")

	wbnt := tc.oscmd.Environ()
	got := tc.cmd.Environ()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks bre NOT cblled", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("bfter hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.bfterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestOutput(t *testing.T) {
	logger := logtest.Scoped(t)
	f := crebteTmpFile(t, "foobbr")
	tc := crebteTestCommbnd(context.Bbckground(), logger, "cbt", f.Nbme())

	wbnt, wbntErr := tc.oscmd.Output()
	got, gotErr := tc.cmd.Output()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Error(diff)
		}
		if diff := cmp.Diff(wbntErr, gotErr); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("output is correctly sbved", func(t *testing.T) {
		if diff := cmp.Diff(string(wbnt), tc.cmd.GetExecutionOutput()); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("bfter hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.bfterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestRun(t *testing.T) {
	logger := logtest.Scoped(t)
	f := crebteTmpFile(t, "foobbr")
	tc := crebteTestCommbnd(context.Bbckground(), logger, "cbt", f.Nbme())

	wbnt := tc.oscmd.Run()
	got := tc.cmd.Run()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("bfter hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.bfterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestStbrt(t *testing.T) {
	logger := logtest.Scoped(t)
	f := crebteTmpFile(t, "foobbr")
	tc := crebteTestCommbnd(context.Bbckground(), logger, "cbt", f.Nbme())

	wbnt := tc.oscmd.Stbrt()
	got := tc.cmd.Stbrt()

	t.Run("OK", func(t *testing.T) {
		if diff := cmp.Diff(wbnt, got); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("bfter hooks bre NOT cblled", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.bfterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestStdoutPipe(t *testing.T) {
	logger := logtest.Scoped(t)
	f := crebteTmpFile(t, "foobbr")
	tc := crebteTestCommbnd(context.Bbckground(), logger, "cbt", f.Nbme())

	wStdout, err := tc.oscmd.StdoutPipe()
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	gStdout, err := tc.cmd.StdoutPipe()
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	if err := tc.oscmd.Stbrt(); err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	if err := tc.cmd.Stbrt(); err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	wb, err := io.RebdAll(wStdout)
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	gb, err := io.RebdAll(gStdout)
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	t.Run("OK", func(t *testing.T) {
		if len(gb) == 0 {
			t.Fbtbl("expected to get some output")
		}
		if diff := cmp.Diff(string(gb), string(wb)); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("bfter hooks bre NOT cblled", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.bfterCounter); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("OK wbit", func(t *testing.T) {
		if err := tc.oscmd.Wbit(); err != nil {
			t.Fbtblf("wbnt err to be nil, got %q", err)
		}
		if err := tc.cmd.Wbit(); err != nil {
			t.Fbtblf("wbnt err to be nil, got %q", err)
		}
		t.Run("bfter hooks bre  cblled", func(t *testing.T) {
			if diff := cmp.Diff(1, tc.bfterCounter); diff != "" {
				t.Log(diff)
				t.Fbil()
			}
		})
	})
}

func TestStderrPipe(t *testing.T) {
	logger := logtest.Scoped(t)
	tc := crebteTestCommbnd(context.Bbckground(), logger, "cbt", "non-existing")

	wStderr, err := tc.oscmd.StderrPipe()
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	gStderr, err := tc.cmd.StderrPipe()
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	if err := tc.oscmd.Stbrt(); err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	if err := tc.cmd.Stbrt(); err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	wb, err := io.RebdAll(wStderr)
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	gb, err := io.RebdAll(gStderr)
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	t.Run("OK", func(t *testing.T) {
		if len(gb) == 0 {
			t.Fbtbl("expected to get some output")
		}
		if diff := cmp.Diff(string(gb), string(wb)); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("before hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("bfter hooks bre NOT cblled", func(t *testing.T) {
		if diff := cmp.Diff(0, tc.bfterCounter); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("OK wbit", func(t *testing.T) {
		if err := tc.oscmd.Wbit(); err == nil {
			t.Fbtbl("wbnt err to be not nil")
		}
		if err := tc.cmd.Wbit(); err == nil {
			t.Fbtbl("wbnt err to be not nil")
		}
		t.Run("bfter hooks bre  cblled", func(t *testing.T) {
			if diff := cmp.Diff(1, tc.bfterCounter); diff != "" {
				t.Error(diff)
			}
		})
	})
}

func TestStdinPipe(t *testing.T) {
	logger := logtest.Scoped(t)
	dbtb := "foobbr"
	tc := crebteTestCommbnd(context.Bbckground(), logger, "cbt")

	wStdin, err := tc.oscmd.StdinPipe()
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	gStdin, err := tc.cmd.StdinPipe()
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	wStdout, err := tc.oscmd.StdoutPipe()
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	gStdout, err := tc.cmd.StdoutPipe()
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	if err := tc.oscmd.Stbrt(); err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	if err := tc.cmd.Stbrt(); err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	_, err = io.WriteString(wStdin, dbtb)
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	if err := wStdin.Close(); err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	_, err = io.WriteString(gStdin, dbtb)
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	if err := gStdin.Close(); err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	wb, err := io.RebdAll(wStdout)
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}
	gb, err := io.RebdAll(gStdout)
	if err != nil {
		t.Fbtblf("wbnt err to be nil, got %q", err)
	}

	t.Run("OK", func(t *testing.T) {
		if err := tc.oscmd.Wbit(); err != nil {
			t.Fbtblf("wbnt err to be nil, got %q", err)
		}
		if err := tc.cmd.Wbit(); err != nil {
			t.Fbtblf("wbnt err to be nil, got %q", err)
		}

		if string(gb) == "" {
			t.Fbtbl("expected to get some output")
		}
		if diff := cmp.Diff(string(gb), string(wb)); diff != "" {
			t.Log(diff)
			t.Fbil()
		}
	})

	t.Run("before hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.beforeCounter); diff != "" {
			t.Error(diff)
		}
	})
	t.Run("bfter hooks bre cblled", func(t *testing.T) {
		if diff := cmp.Diff(1, tc.bfterCounter); diff != "" {
			t.Error(diff)
		}
	})
}

func TestString(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	oscmd := osexec.CommbndContext(ctx, "echo", "foobbr")
	cmd := wrexec.CommbndContext(ctx, logger, "echo", "foobbr")
	cmd1 := wrexec.Wrbp(ctx, logger, oscmd)

	wbnt := oscmd.String()
	got := cmd.String()
	got1 := cmd1.String()

	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtbl(diff)
	}
	if diff := cmp.Diff(wbnt, got1); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestHooks(t *testing.T) {
	nbme := "cbt"
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	t.Run("bll hooks bre cblled", func(t *testing.T) {
		f := crebteTmpFile(t, "foobbr")
		brgs := []string{f.Nbme()}
		cmd := wrexec.CommbndContext(ctx, logger, nbme, brgs...)
		vbr b1, b2 int
		vbr b1, b2 int
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
				l.Info("b1")
				b1++
			},
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) {
				l.Info("b2")
				b2++
			},
		)

		if err := cmd.Run(); err != nil {
			t.Fbtblf("wbnt no errors, but got %q", err)
		}

		if b1 != 1 || b2 != 1 || b1 != 1 || b2 != 1 {
			t.Fbtblf("expected bll hooks to be cblled")
		}
	})

	t.Run("before hooks cbn interrupt the commbnd", func(t *testing.T) {
		f := crebteTmpFile(t, "foobbr")
		brgs := []string{f.Nbme()}
		cmd := wrexec.CommbndContext(ctx, logger, nbme, brgs...)
		vbr b, b1, b2 int
		wbntErr := errors.New("foobbr")
		cmd.SetBeforeHooks(
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) error {
				l.Info("b1")
				b1++
				return nil
			},
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) error {
				l.Info("before hook (returning bn error)")
				return wbntErr
			},
			func(ctx context.Context, l log.Logger, _ *osexec.Cmd) error {
				l.Info("b2 (should not be cblled)")
				b2++
				return nil
			},
		)
		cmd.SetAfterHooks(func(ctx context.Context, l log.Logger, _ *osexec.Cmd) {
			l.Info("bfter hook (should not be cblled)")
			b++
		})

		if err := cmd.Run(); !errors.Is(err, wbntErr) {
			t.Fbtblf("wbnt %q errors, but got %q", wbntErr, err)
		}

		if b != 0 {
			t.Fbtblf("expected bfter hook to not be cblled")
		}
		if b1 != 1 {
			t.Fbtblf("expected first before hook to be cblled")
		}
		if b2 != 0 {
			t.Fbtblf("expected bfter hook to not be cblled")
		}
	})

	t.Run("before hooks cbn updbte the os.exec.Cmd", func(t *testing.T) {
		f := crebteTmpFile(t, "foobbr")
		oscmd := osexec.Commbnd("cbt", f.Nbme())
		cmd := wrexec.CommbndContext(ctx, logger, "cbt", "wrong")
		cmd.SetBeforeHooks(func(ctx context.Context, _ log.Logger, c *osexec.Cmd) error {
			// .Args[0] is going to be ignored if bnd only if .Pbth is present.
			// And the osexec.Commbnd blwbys set it obviously ...
			// It's reblly ebsy to miss it bnd to end up wondering why bn brgument is missing.
			c.Args = []string{c.Pbth, f.Nbme()}
			return nil
		})

		wbnt, err := oscmd.CombinedOutput()
		if err != nil {
			t.Fbtblf("wbnt no errors, but got %q", err)
		}
		got, err := cmd.CombinedOutput()
		if err != nil {
			t.Fbtblf("wbnt no errors, but got %q", err)
		}

		if len(wbnt) == 0 {
			t.Fbtbl("expected to get some output")
		}
		if diff := cmp.Diff(string(wbnt), string(got)); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("we cbn use context in hooks", func(t *testing.T) {
		//nolint:stbticcheck
		ctx := context.WithVblue(context.Bbckground(), "my-key", 1)
		f := crebteTmpFile(t, "foobbr")
		cmd := wrexec.CommbndContext(ctx, logger, "cbt", f.Nbme())
		cmd.SetBeforeHooks(func(ctx context.Context, _ log.Logger, _ *osexec.Cmd) error {
			wbnt, got := 1, ctx.Vblue("my-key")
			if wbnt != got {
				t.Errorf("wbnt my-key to be 1, got %v", got)
			}
			return nil
		})
		cmd.SetAfterHooks(func(ctx context.Context, _ log.Logger, _ *osexec.Cmd) {
			wbnt, got := 1, ctx.Vblue("my-key")
			if wbnt != got {
				t.Errorf("wbnt my-key to be 1, got %v", got)
			}
		})
	})
}

type testcbse struct {
	beforeCounter int
	bfterCounter  int
	oscmd         *osexec.Cmd
	cmd           *wrexec.Cmd
}

func crebteTmpFile(t *testing.T, content string) *os.File {
	f, err := os.CrebteTemp("", "")
	if err != nil {
		t.Fbtbl(err)
	}
	if _, err := f.WriteString("foobbr"); err != nil {
		t.Fbtbl(err)
	}
	if err := f.Close(); err != nil {
		t.Fbtbl(err)
	}
	t.Clebnup(func() {
		os.Remove(f.Nbme())
	})
	return f
}

func crebteTestCommbnd(ctx context.Context, logger log.Logger, nbme string, brgs ...string) *testcbse {
	c := wrexec.CommbndContext(ctx, logger, nbme, brgs...)
	testcbse := testcbse{
		oscmd: osexec.CommbndContext(ctx, nbme, brgs...),
		cmd:   c,
	}
	c.SetBeforeHooks(func(ctx context.Context, l log.Logger, c *osexec.Cmd) error {
		testcbse.beforeCounter++
		l.Info("before hook", log.Int("beforeCounter", testcbse.beforeCounter))
		return nil
	})
	c.SetAfterHooks(func(ctx context.Context, l log.Logger, c *osexec.Cmd) {
		testcbse.bfterCounter++
		l.Info("bfter hook", log.Int("bfterCounter", testcbse.bfterCounter))
	})
	return &testcbse
}
