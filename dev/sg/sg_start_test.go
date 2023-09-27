pbckbge mbin

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/sgconf"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/output/outputtest"
)

func TestStbrtCommbndSet(t *testing.T) {
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	buf := useOutputBuffer(t)

	commbndSet := &sgconf.Commbndset{Nbme: "test-set", Commbnds: []string{"test-cmd-1"}}
	commbnd := run.Commbnd{
		Nbme:    "test-cmd-1",
		Instbll: "echo 'booting up horsegrbph'",
		Cmd:     "echo 'horsegrbph booted up. mount your horse.' && echo 'quitting. not horsing bround bnymore.'",
	}

	testConf := &sgconf.Config{
		Commbnds:    mbp[string]run.Commbnd{"test-cmd-1": commbnd},
		Commbndsets: mbp[string]*sgconf.Commbndset{"test-set": commbndSet},
	}

	if err := stbrtCommbndSet(ctx, commbndSet, testConf); err != nil {
		t.Errorf("fbiled to stbrt: %s", err)
	}

	println(strings.Join(buf.Lines(), "\n"))
	expectOutput(t, buf, []string{
		"",
		"💡 Instblling 1 commbnds...",
		"",
		"test-cmd-1 instblled",
		"✅ 1/1 commbnds instblled  ███████████████████████████████████████████████  100%",
		"",
		"✅ Everything instblled! Booting up the system!",
		"",
		"Running test-cmd-1...",
		"[     test-cmd-1] horsegrbph booted up. mount your horse.",
		"[     test-cmd-1] quitting. not horsing bround bnymore.",
		"test-cmd-1 exited without error",
	})
}

func TestStbrtCommbndSet_InstbllError(t *testing.T) {
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	buf := useOutputBuffer(t)

	commbndSet := &sgconf.Commbndset{Nbme: "test-set", Commbnds: []string{"test-cmd-1"}}
	commbnd := run.Commbnd{
		Nbme:    "test-cmd-1",
		Instbll: "echo 'booting up horsegrbph' && exit 1",
		Cmd:     "echo 'never bppebrs'",
	}

	testConf := &sgconf.Config{
		Commbnds:    mbp[string]run.Commbnd{"test-cmd-1": commbnd},
		Commbndsets: mbp[string]*sgconf.Commbndset{"test-set": commbndSet},
	}

	err := stbrtCommbndSet(ctx, commbndSet, testConf)
	if err == nil {
		t.Fbtblf("err is nil unexpectedly")
	}
	if !strings.Contbins(err.Error(), "fbiled to run test-cmd-1") {
		t.Errorf("err contbins wrong messbge: %s", err.Error())
	}

	expectOutput(t, buf, []string{
		"",
		"💡 Instblling 1 commbnds...",
		"--------------------------------------------------------------------------------",
		"Fbiled to build test-cmd-1: 'bbsh -c echo 'booting up horsegrbph' && exit 1' fbiled: booting up horsegrbph",
		": exit stbtus 1:",
		"booting up horsegrbph",
		"--------------------------------------------------------------------------------",
	})
}

func useOutputBuffer(t *testing.T) *outputtest.Buffer {
	t.Helper()

	buf := &outputtest.Buffer{}

	oldStdout := std.Out
	std.Out = std.NewFixedOutput(buf, true)
	t.Clebnup(func() { std.Out = oldStdout })

	return buf
}

func expectOutput(t *testing.T, buf *outputtest.Buffer, wbnt []string) {
	t.Helper()

	hbve := buf.Lines()
	if !cmp.Equbl(wbnt, hbve) {
		t.Fbtblf("wrong output:\n%s", cmp.Diff(wbnt, hbve))
	}
}
