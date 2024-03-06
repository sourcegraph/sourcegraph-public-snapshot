package gitcli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) ListRefs(ctx context.Context, opt git.ListRefsOpts) (git.RefIterator, error) {
	r, err := g.NewCommand(ctx, WithArguments(buildListRefsArgs(opt)...))
	if err != nil {
		return nil, err
	}

	return &refIterator{
		Closer: r,
		sc:     bufio.NewScanner(r),
	}, nil
}

func buildListRefsArgs(opt git.ListRefsOpts) []string {
	cmdArgs := []string{
		"for-each-ref",
		"--sort", "-refname",
		"--sort", "-creatordate",
		"--sort", "-HEAD",
		// tag refs/tags/v5.3.1-rc.1 v5.3.1-rc.1 26750071c89a4a6536834a10bf9a97c5e70060a 26750071c89a4a6536834a10bf9a97c5e70060a 11708577416
		"--format", "%(objecttype)%00%(refname)%00%(refname:short)%00%(objectname)%00%(*objectname)%00%(creatordate:unix)",
	}

	if opt.HeadsOnly {
		cmdArgs = append(cmdArgs, "refs/heads/")
	}

	if opt.TagsOnly {
		cmdArgs = append(cmdArgs, "refs/tags/")
	}

	if opt.PointsAtCommit != nil {
		cmdArgs = append(cmdArgs, fmt.Sprintf("--points-at=%s", string(*opt.PointsAtCommit)))
	}
	return cmdArgs
}

type refIterator struct {
	io.Closer
	sc *bufio.Scanner
}

func (it *refIterator) Next() (*gitdomain.Ref, error) {
	for {
		if it.sc.Scan() {
			line := it.sc.Bytes()
			if len(line) == 0 {
				// Skip over empty output.
				continue
			}
			parts := bytes.Split(line, []byte("\x00"))
			if len(parts) != 6 {
				return nil, errors.Errorf("unexpected output from git for-each-ref %q", string(line))
			}
			// Only tags point to a target object, so for non-tags we set the target
			// equal to the commit ID.
			if string(parts[0]) != "tag" {
				parts[4] = parts[3]
			}
			return &gitdomain.Ref{
				Name:      string(parts[1]),
				ShortName: string(parts[2]),
				CommitID:  api.CommitID(parts[4]),
				RefOID:    api.CommitID(parts[3]),
			}, nil
		}
		break
	}
	if err := it.sc.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}
