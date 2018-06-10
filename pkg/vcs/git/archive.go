package git

import (
	"context"
	"io"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
)

// ArchiveOptions contains options for the Archive func.
type ArchiveOptions struct {
	Treeish string   // the tree or commit to produce an archive for
	Format  string   // format of the resulting archive (usually "tar" or "zip")
	Paths   []string // if nonempty, only include these paths
}

// Archive produces an archive from a Git repository.
func Archive(ctx context.Context, repo gitserver.Repo, opt ArchiveOptions) (_ io.ReadCloser, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: Archive")
	span.SetTag("Repo", repo.Name)
	span.SetTag("Treeish", opt.Treeish)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	if err := checkSpecArgSafety(string(opt.Treeish)); err != nil {
		return nil, err
	}

	cmd := gitserver.DefaultClient.Command("git",
		"archive",

		// Suppresses fatal error when the repo contains paths matching **/.git/** and instead
		// includes those files (to allow archiving invalid such repos). This is unexpected
		// behavior; the --worktree-attributes flag should merely let us specify a gitattributes
		// file that contains `**/.git/** export-ignore`, but it actually makes everything work as
		// desired. Tested by the "repo with .git dir" test case.
		"--worktree-attributes",

		"--format="+opt.Format,
	)
	if opt.Format == "zip" {
		// Compression level of 0 (no compression) seems to perform the
		// best overall on fast network links, but this has not been tuned
		// thoroughly.
		cmd.Args = append(cmd.Args, "-0")
	}
	cmd.Args = append(cmd.Args, string(opt.Treeish), "--")
	cmd.Args = append(cmd.Args, opt.Paths...)
	cmd.Repo = repo
	rc, err := gitserver.StdoutReader(ctx, cmd)
	if err != nil {
		if errcode.IsNotFound(err) {
			err = badRequestError{err.Error()}
		}
		return nil, err
	}
	return rc, nil
}

type badRequestError struct{ msg string }

func (e badRequestError) Error() string    { return e.msg }
func (e badRequestError) BadRequest() bool { return true }
