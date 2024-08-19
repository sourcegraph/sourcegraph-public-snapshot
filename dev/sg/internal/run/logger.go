package run

import (
	"context"
	"hash/fnv"
	"io"
	"strconv"

	"go.bobheadxi.dev/streamline/pipe"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

func nameToColor(s string) output.Style {
	h := fnv.New32()
	h.Write([]byte(s))
	// We don't use 256 colors because some of those are too dark/bright and hard to read
	return output.Fg256Color(int(h.Sum32()) % 220)
}

var (
	// NOTE: This always adds a newline, which is not always what we want. When
	// we flush partial lines, we don't want to add a newline character. What
	// we need to do: extend the `*output.Output` type to have a
	// `WritefNoNewline` (yes, bad name) method.
	//
	// Some rare commands will have names larger than 'maxNameLength' chars, but
	// that's fine, we'll truncate the names. How to quickly check commands names:
	//
	//   cue eval --out=json sg.config.yaml | jq '.commands | keys'
	//
	maxNameLength = 15
	lineFormat    = "%s%s[%+" + strconv.Itoa(maxNameLength) + "s]%s %s"
)

// newOutputPipe returns a new output with a unique color based on the name of the cmd
// that blocks until the given start signal and writes logs to the given output.Output.
func newOutputPipe(ctx context.Context, name string, out *output.Output, start <-chan struct{}) io.Writer {
	name = compactName(name)
	color := nameToColor(name)

	w, stream := pipe.NewStream()
	go func() {
		<-start
		err := stream.Stream(func(line string) {
			out.Writef(lineFormat, output.StyleBold, color, name, output.StyleReset, line)
		})
		_ = w.CloseWithError(err)
	}()
	context.AfterFunc(ctx, func() {
		_ = w.CloseWithError(ctx.Err())
	})

	return w
}

func compactName(name string) string {
	length := len(name)
	if length > maxNameLength {
		// Use the first part of the name and the very last character to hint at what's
		// up, useful for long command names with index suffices (e.g. service-1, service-2)
		name = name[:maxNameLength-4] + "..." + string(name[length-1])
	}
	return name
}
