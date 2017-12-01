package gitcmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

func readUntilTimeout(ctx context.Context, cmd *gitserver.Cmd) (data []byte, complete bool, err error) {
	sr, err := gitserver.StdoutReader(ctx, cmd)
	if urlErr, ok := err.(*url.Error); ok && urlErr.Err == context.DeadlineExceeded {
		// Continue; the gitserver call exceeded our deadline before the command
		// produced any output.
	} else if err != nil {
		return nil, false, err
	}

	if sr != nil {
		defer sr.Close()
		var err error
		data, err = ioutil.ReadAll(sr)
		if err == nil {
			complete = true
		} else if err != nil && err != context.DeadlineExceeded {
			data = bytes.TrimSpace(data)
			if isBadObjectErr(string(data), "") || isInvalidRevisionRangeError(string(data), "") {
				return nil, true, vcs.ErrRevisionNotFound
			}
			if len(data) > 100 {
				data = append(data[:100], []byte("... (truncated)")...)
			}
			return nil, true, fmt.Errorf("exec git %v command failed: %s. Output was:\n\n%s", cmd.Args, err, data)
		}
	}

	return data, complete, nil
}
