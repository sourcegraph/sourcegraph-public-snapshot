package gitcli

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) ReadFile(ctx context.Context, commit api.CommitID, path string) (io.ReadCloser, error) {
	if err := gitdomain.EnsureAbsoluteCommit(commit); err != nil {
		return nil, err
	}

	blobOID, err := g.getBlobOID(ctx, commit, path)
	if err != nil {
		if err == errIsSubmodule {
			return io.NopCloser(bytes.NewReader(nil)), nil
		}
		return nil, err
	}

	cmd, cancel, err := g.gitCommand(ctx, "cat-file", "-p", string(blobOID))
	if err != nil {
		cancel()
		return nil, err
	}

	r, err := g.runGitCommand(ctx, cmd)
	if err != nil {
		cancel()
		return nil, err
	}

	return &closingFileReader{
		ReadCloser: r,
		onClose:    cancel,
	}, nil
}

var errIsSubmodule = errors.New("blob is a submodule")

func (g *gitCLIBackend) getBlobOID(ctx context.Context, commit api.CommitID, path string) (api.CommitID, error) {
	cmd, cancel, err := g.gitCommand(ctx, "ls-tree", string(commit), "--", path)
	defer cancel()
	if err != nil {
		return "", err
	}

	out, err := g.runGitCommand(ctx, cmd)
	if err != nil {
		return "", err
	}
	defer out.Close()

	stdout, err := io.ReadAll(out)
	if err != nil {
		// If exit code is 128 and `not a tree object` is part of stderr, most likely we
		// are referencing a commit that does not exist.
		// We want to return a gitdomain.RevisionNotFoundError in that case.
		var e *CommandFailedError
		if errors.As(err, &e) && e.ExitStatus == 128 {
			if bytes.Contains(e.Stderr, []byte("not a tree object")) || bytes.Contains(e.Stderr, []byte("Not a valid object name")) {
				return "", &gitdomain.RevisionNotFoundError{Repo: g.repoName, Spec: string(commit)}
			}
		}

		return "", err
	}

	stdout = bytes.TrimSpace(stdout)
	if len(stdout) == 0 {
		return "", &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}

	// format: 100644 blob 3bad331187e39c05c78a9b5e443689f78f4365a7	README.md
	fields := bytes.Fields(stdout)
	if len(fields) < 3 {
		return "", errors.Newf("unexpected output while parsing blob OID: %q", string(stdout))
	}
	if string(fields[0]) == "160000" {
		return "", errIsSubmodule
	}
	return api.CommitID(fields[2]), nil
}

type closingFileReader struct {
	io.ReadCloser
	onClose func()
}

func (r *closingFileReader) Close() error {
	err := r.ReadCloser.Close()
	r.onClose()
	return err
}

// func (c *clientImplementor) showRef(ctx context.Context, repo api.RepoName, args ...string) ([]gitdomain.Ref, error) {
// 	cmdArgs := append([]string{"show-ref"}, args...)
// 	cmd := c.gitCommand(repo, cmdArgs...)
// 	out, err := cmd.CombinedOutput(ctx)
// 	if err != nil {
// 		if gitdomain.IsRepoNotExist(err) {
// 			return nil, err
// 		}
// 		// Exit status of 1 and no output means there were no
// 		// results. This is not a fatal error.
// 		if cmd.ExitStatus() == 1 && len(out) == 0 {
// 			return nil, nil
// 		}
// 		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
// 	}

// 	out = bytes.TrimSuffix(out, []byte("\n")) // remove trailing newline
// 	lines := bytes.Split(out, []byte("\n"))
// 	sort.Sort(byteSlices(lines)) // sort for consistency
// 	refs := make([]gitdomain.Ref, len(lines))
// 	for i, line := range lines {
// 		if len(line) <= 41 {
// 			return nil, errors.New("unexpectedly short (<=41 bytes) line in `git show-ref ...` output")
// 		}
// 		id := line[:40]
// 		name := line[41:]
// 		refs[i] = gitdomain.Ref{Name: string(name), CommitID: api.CommitID(id)}
// 	}
// 	return refs, nil
// }

// type showRefOpts struct {
// 	pointsAt  string
// 	headsOnly bool
// 	tagsOnly  bool
// }

// func (c *clientImplementor) showRef(ctx context.Context, repo api.RepoName, opt showRefOpts) ([]gitdomain.Ref, error) {
// 	cmdArgs := []string{
// 		"for-each-ref",
// 		"--sort", "-refname",
// 		"--sort", "-creatordate",
// 		"--sort", "-HEAD",
// 		"--format", "%(objecttype)%00%%(refname)%00%%(refname:short)%00%%(objectname)%00%%(creatordate:unix)",
// 	}

// 	if opt.headsOnly {
// 		cmdArgs = append(cmdArgs, "refs/heads/")
// 	}

// 	if opt.tagsOnly {
// 		cmdArgs = append(cmdArgs, "refs/tags/")
// 	}

// 	if opt.pointsAt != "" {
// 		cmdArgs = append(cmdArgs, "--points-at", opt.pointsAt)
// 	}

// 	// Support both lightweight tags and tag objects. For creatordate, use an %(if) to prefer the
// 	// taggerdate for tag objects, otherwise use the commit's committerdate (instead of just always
// 	// using committerdate).
// 	args := []string{
// 		"tag",
// 		"--list",
// 		"--sort", "-creatordate",
// 		"--format", "%(if)%(*objectname)%(then)%(*objectname)%(else)%(objectname)%(end)%00%(refname:short)%00%(if)%(creatordate:unix)%(then)%(creatordate:unix)%(else)%(*creatordate:unix)%(end)",
// 	}

// 	cmd := c.gitCommand(repo, cmdArgs...)
// 	out, err := cmd.CombinedOutput(ctx)
// 	if err != nil {
// 		if gitdomain.IsRepoNotExist(err) {
// 			return nil, err
// 		}
// 		// Exit status of 1 and no output means there were no
// 		// results. This is not a fatal error.
// 		if cmd.ExitStatus() == 1 && len(out) == 0 {
// 			return nil, nil
// 		}
// 		return nil, errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args(), out))
// 	}

// 	out = bytes.TrimSuffix(out, []byte("\n")) // remove trailing newline
// 	lines := bytes.Split(out, []byte("\n"))
// 	// sort.Sort(byteSlices(lines)) // sort for consistency
// 	refs := make([]gitdomain.Ref, len(lines))
// 	for i, line := range lines {
// 		if len(line) <= 41 {
// 			return nil, errors.New("unexpectedly short (<=41 bytes) line in `git show-ref ...` output")
// 		}
// 		fields := bytes.Split(line, []byte("\x00"))
// 		if len(fields) != 4 {
// 			continue
// 		}
// 		if string(fields[0]) != "commit" {
// 			continue
// 		}
// 		id := fields[2]
// 		name := fields[1]
// 		refs[i] = gitdomain.Ref{Name: string(name), CommitID: api.CommitID(id)}
// 	}
// 	return refs, nil
// }
