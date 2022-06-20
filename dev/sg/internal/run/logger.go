package run

import (
	"context"
	"hash/fnv"
	"strings"

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
	name = compactName(name)
	color := nameToColor(name)

	sink := func(data string) {
		// NOTE: This always adds a newline, which is not always what we want. When
		// we flush partial lines, we don't want to add a newline character. What
		// we need to do: extend the `*output.Output` type to have a
		// `WritefNoNewline` (yes, bad name) method.
		//
		// Some rare commands will have names larger than 15 chars, but that's fine.
		// How to quickly check commands names: cue eval --out=json sg.config.yaml | jq '.commands | keys'
		out.Writef("%s%s[%+15s]%s %s", output.StyleBold, color, name, output.StyleReset, data)
	}

	return process.NewLogger(ctx, sink)
}

func compactName(name string) string {
	if strings.HasPrefix(name, "enterprise-") {
		return strings.Replace(name, "enterprise-", "e-", 1)
	}
	if strings.HasPrefix(name, "zoekt-") {
		return strings.Replace(name, "zoekt-", "z-", 1)
	}
	return name
}
