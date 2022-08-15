package outputtest

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/output"
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

	progress := out.Progress([]output.ProgressBar{
		{Label: "Bar-A", Max: 1.0},
		{Label: "Bar-B", Max: 1.0, Value: 0.5},
		{Label: "Bar-C", Max: 1.0, Value: 0.7},
	}, nil)

	expectOutput(t, buf, []string{
		"✅ Hello there!",
		"⠋  Bar-A                                                                      0%",
		"⠋  Bar-B  ████████████████████████████████▌                                  50%",
		"⠋  Bar-C  █████████████████████████████████████████████▌                     70%",
	})

	progress.SetValue(0, 0.5)
	progress.SetValue(1, 0.8)
	progress.SetValue(2, 1.0)

	expectOutput(t, buf, []string{
		"✅ Hello there!",
		"⠋  Bar-A  ████████████████████████████████▌                                  50%",
		"⠋  Bar-B  ████████████████████████████████████████████████████               80%",
		"✅ Bar-C  ████████████████████████████████████████████████████████████████  100%",
	})

	progress.Complete()

	expectOutput(t, buf, []string{
		"✅ Hello there!",
		"✅ Bar-A  ████████████████████████████████████████████████████████████████  100%",
		"✅ Bar-B  ████████████████████████████████████████████████████████████████  100%",
		"✅ Bar-C  ████████████████████████████████████████████████████████████████  100%",
	})
}

func expectOutput(t *testing.T, buf *Buffer, want []string) {
	t.Helper()

	have := buf.Lines()
	if !cmp.Equal(want, have) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, have))
	}
}
