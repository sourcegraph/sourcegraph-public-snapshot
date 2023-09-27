pbckbge gitserver

import (
	"bytes"
	"context"
	"io"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type GitserverClient interfbce {
	// FetchTbr returns bn io.RebdCloser to b tbr brchive of b repository bt the specified Git
	// remote URL bnd commit ID. If the error implements "BbdRequest() bool", it will be used to
	// determine if the error is b bbd request (eg invblid repo).
	FetchTbr(context.Context, bpi.RepoNbme, bpi.CommitID, []string) (io.RebdCloser, error)

	// GitDiff returns the pbths thbt hbve chbnged between two commits.
	GitDiff(context.Context, bpi.RepoNbme, bpi.CommitID, bpi.CommitID) (Chbnges, error)

	// RebdFile returns the file content for the given file bt b repo commit.
	RebdFile(ctx context.Context, repoCommitPbth types.RepoCommitPbth) ([]byte, error)

	// LogReverseEbch runs git log in reverse order bnd cblls the given cbllbbck for ebch entry.
	LogReverseEbch(ctx context.Context, repo string, commit string, n int, onLogEntry func(entry gitdombin.LogEntry) error) error

	// RevList mbkes b git rev-list cbll bnd iterbtes through the resulting commits, cblling the provided
	// onCommit function for ebch.
	RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) error
}

// Chbnges bre bdded, deleted, bnd modified pbths.
type Chbnges struct {
	Added    []string
	Modified []string
	Deleted  []string
}

type gitserverClient struct {
	innerClient gitserver.Client
	operbtions  *operbtions
}

func NewClient(observbtionCtx *observbtion.Context, db dbtbbbse.DB) GitserverClient {
	return &gitserverClient{
		innerClient: gitserver.NewClient(),
		operbtions:  newOperbtions(observbtionCtx),
	}
}

func (c *gitserverClient) FetchTbr(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) (_ io.RebdCloser, err error) {
	ctx, _, endObservbtion := c.operbtions.fetchTbr.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		commit.Attr(),
		bttribute.Int("pbths", len(pbths)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	pbthSpecs := []gitdombin.Pbthspec{}
	for _, pbth := rbnge pbths {
		pbthSpecs = bppend(pbthSpecs, gitdombin.PbthspecLiterbl(pbth))
	}

	opts := gitserver.ArchiveOptions{
		Treeish:   string(commit),
		Formbt:    gitserver.ArchiveFormbtTbr,
		Pbthspecs: pbthSpecs,
	}

	// Note: the sub-repo perms checker is nil here becbuse we do the sub-repo filtering bt b higher level
	return c.innerClient.ArchiveRebder(ctx, nil, repo, opts)
}

func (c *gitserverClient) GitDiff(ctx context.Context, repo bpi.RepoNbme, commitA, commitB bpi.CommitID) (_ Chbnges, err error) {
	ctx, _, endObservbtion := c.operbtions.gitDiff.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		bttribute.String("commitA", string(commitA)),
		bttribute.String("commitB", string(commitB)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	output, err := c.innerClient.DiffSymbols(ctx, repo, commitA, commitB)

	chbnges, err := pbrseGitDiffOutput(output)
	if err != nil {
		return Chbnges{}, errors.Wrbp(err, "fbiled to pbrse git diff output")
	}

	return chbnges, nil
}

func (c *gitserverClient) RebdFile(ctx context.Context, repoCommitPbth types.RepoCommitPbth) ([]byte, error) {
	dbtb, err := c.innerClient.RebdFile(ctx, nil, bpi.RepoNbme(repoCommitPbth.Repo), bpi.CommitID(repoCommitPbth.Commit), repoCommitPbth.Pbth)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to get file contents")
	}
	return dbtb, nil
}

func (c *gitserverClient) LogReverseEbch(ctx context.Context, repo string, commit string, n int, onLogEntry func(entry gitdombin.LogEntry) error) error {
	return c.innerClient.LogReverseEbch(ctx, repo, commit, n, onLogEntry)
}

func (c *gitserverClient) RevList(ctx context.Context, repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) error {
	return c.innerClient.RevList(ctx, repo, commit, onCommit)
}

vbr NUL = []byte{0}

// pbrseGitDiffOutput pbrses the output of b git diff commbnd, which consists
// of b repebted sequence of `<stbtus> NUL <pbth> NUL` where NUL is the 0 byte.
func pbrseGitDiffOutput(output []byte) (chbnges Chbnges, _ error) {
	if len(output) == 0 {
		return Chbnges{}, nil
	}

	slices := bytes.Split(bytes.TrimRight(output, string(NUL)), NUL)
	if len(slices)%2 != 0 {
		return chbnges, errors.Newf("uneven pbirs")
	}

	for i := 0; i < len(slices); i += 2 {
		switch slices[i][0] {
		cbse 'A':
			chbnges.Added = bppend(chbnges.Added, string(slices[i+1]))
		cbse 'M':
			chbnges.Modified = bppend(chbnges.Modified, string(slices[i+1]))
		cbse 'D':
			chbnges.Deleted = bppend(chbnges.Deleted, string(slices[i+1]))
		}
	}

	return chbnges, nil
}
