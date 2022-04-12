package run

import (
	"context"
	"hash/fnv"

	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/lib/process"
)

func nameToColor(s string) output.Style {
	h := fnv.New32()
	h.Write([]byte(s))
	// We don't use 256 colors because some of those are too dark/bright and hard to read
	return output.Fg256Color(int(h.Sum32()) % 220)
}

// newCmdLogger returns a new process.Logger with a unique color based on the name of the cmd.
func newCmdLogger(ctx context.Context, name string, out *output.Output) *process.Logger {
	color := nameToColor(name)

	sink := func(data string) {
		// NOTE: This always adds a newline, which is not always what we want. When
		// we flush partial lines, we don't want to add a newline character. What
		// we need to do: extend the `*output.Output` type to have a
		// `WritefNoNewline` (yes, bad name) method.
		out.Writef("%s%s[%s]%s %s", output.StyleBold, color, name, output.StyleReset, data)
	}

	return process.NewLogger(ctx, sink)
}
