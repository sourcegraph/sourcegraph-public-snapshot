package proto

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/buf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate"
)

func Generate(ctx context.Context, bufGenFilePaths []string, verboseOutput bool) *generate.Report {
	var (
		start = time.Now()
		sb    strings.Builder
	)

	output := std.NewOutput(&sb, verboseOutput)
	err := buf.InstallDependencies(ctx, output)
	if err != nil {
		err = errors.Wrap(err, "installing buf dependencies")
		return &generate.Report{Output: sb.String(), Err: err}
	}

	for _, p := range bufGenFilePaths {
		sb.WriteString(fmt.Sprintf("> Generate %s\n", p))
		bufArgs := []string{"generate"}
		c, err := buf.Cmd(ctx, bufArgs...)
		if err != nil {
			err = errors.Wrap(err, "creating buf command")
			return &generate.Report{Err: err}
		}

		// Run buf generate in the directory of the buf.gen.yaml file
		d := filepath.Dir(p)
		c.Dir(d)

		err = c.Run().Stream(&sb)
		if err != nil {
			commandString := fmt.Sprintf("buf %s", strings.Join(bufArgs, " "))
			err = errors.Wrapf(err, "running %q", commandString)
			return &generate.Report{Output: sb.String(), Err: err}
		}
	}

	return &generate.Report{
		Output:   sb.String(),
		Duration: time.Since(start),
	}
}
