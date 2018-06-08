package git

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
)

// ShortLog returns the per-author commit statistics of the repo.
func (r *Repository) ShortLog(ctx context.Context, opt vcs.ShortLogOptions) ([]*vcs.PersonCount, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ShortLog")
	span.SetTag("Opt", opt)
	defer span.Finish()

	if opt.Range == "" {
		opt.Range = "HEAD"
	}
	if err := checkSpecArgSafety(opt.Range); err != nil {
		return nil, err
	}

	args := []string{"shortlog", "-sne", "--no-merges"}
	if opt.After != "" {
		args = append(args, "--after="+opt.After)
	}
	args = append(args, opt.Range, "--")
	if opt.Path != "" {
		args = append(args, opt.Path)
	}
	cmd := r.command("git", args...)
	out, err := cmd.Output(ctx)
	if err != nil {
		return nil, fmt.Errorf("exec `git shortlog -sne` failed: %v", err)
	}

	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil, nil
	}
	lines := bytes.Split(out, []byte{'\n'})
	results := make([]*vcs.PersonCount, len(lines))
	for i, line := range lines {
		match := logEntryPattern.FindSubmatch(line)
		if match == nil {
			return nil, fmt.Errorf("invalid git shortlog line: %q", line)
		}
		count, err := strconv.Atoi(string(match[1]))
		if err != nil {
			return nil, err
		}
		addr, err := mail.ParseAddress(string(match[2]))
		if err != nil || addr == nil {
			addr = &mail.Address{Name: string(match[2])}
		}
		results[i] = &vcs.PersonCount{
			Count: int32(count),
			Name:  addr.Name,
			Email: addr.Address,
		}
	}
	return results, nil
}
