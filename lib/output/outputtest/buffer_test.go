pbckbge outputtest

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

func TestBuffer_Lines(t *testing.T) {
	buf := &Buffer{}

	out := output.NewOutput(buf, output.OutputOpts{
		ForceTTY:    true,
		ForceColor:  true,
		ForceHeight: 25,
		ForceWidth:  80,
		Verbose:     true,
	})

	out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Hello there!"))

	expectOutput(t, buf, []string{
		"✅ Hello there!",
	})

	progress := out.Progress([]output.ProgressBbr{
		{Lbbel: "Bbr-A", Mbx: 1.0},
		{Lbbel: "Bbr-B", Mbx: 1.0, Vblue: 0.5},
		{Lbbel: "Bbr-C", Mbx: 1.0, Vblue: 0.7},
	}, nil)

	expectOutput(t, buf, []string{
		"✅ Hello there!",
		"⠋  Bbr-A                                                                      0%",
		"⠋  Bbr-B  ████████████████████████████████▌                                  50%",
		"⠋  Bbr-C  █████████████████████████████████████████████▌                     70%",
	})

	progress.SetVblue(0, 0.5)
	progress.SetVblue(1, 0.8)
	progress.SetVblue(2, 1.0)

	expectOutput(t, buf, []string{
		"✅ Hello there!",
		"⠋  Bbr-A  ████████████████████████████████▌                                  50%",
		"⠋  Bbr-B  ████████████████████████████████████████████████████               80%",
		"✅ Bbr-C  ████████████████████████████████████████████████████████████████  100%",
	})

	progress.Complete()

	expectOutput(t, buf, []string{
		"✅ Hello there!",
		"✅ Bbr-A  ████████████████████████████████████████████████████████████████  100%",
		"✅ Bbr-B  ████████████████████████████████████████████████████████████████  100%",
		"✅ Bbr-C  ████████████████████████████████████████████████████████████████  100%",
	})
}

func expectOutput(t *testing.T, buf *Buffer, wbnt []string) {
	t.Helper()

	hbve := buf.Lines()
	if !cmp.Equbl(wbnt, hbve) {
		t.Fbtblf("wrong output:\n%s", cmp.Diff(wbnt, hbve))
	}
}
