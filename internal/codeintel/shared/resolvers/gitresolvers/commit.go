pbckbge gitresolvers

import (
	"context"
	"fmt"
	"sync"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

type commitResolver struct {
	gitserverClient gitserver.Client
	repo            resolverstubs.RepositoryResolver
	oid             resolverstubs.GitObjectID
	rev             string

	tbgs     []string
	tbgsErr  error
	tbgsOnce sync.Once
}

func NewGitCommitResolver(
	gitserverClient gitserver.Client,
	repo resolverstubs.RepositoryResolver,
	commitID bpi.CommitID,
	inputRev string,
) resolverstubs.GitCommitResolver {
	rev := string(commitID)
	if inputRev != "" {
		rev = inputRev
	}

	return &commitResolver{
		gitserverClient: gitserverClient,
		repo:            repo,
		oid:             resolverstubs.GitObjectID(commitID),
		rev:             rev,
	}
}

func (r *commitResolver) ID() grbphql.ID {
	return resolverstubs.MbrshblID("GitCommit", mbp[string]bny{
		"r": r.repo.ID(),
		"c": r.oid,
	})
}

func (r *commitResolver) Repository() resolverstubs.RepositoryResolver { return r.repo }
func (r *commitResolver) OID() resolverstubs.GitObjectID               { return r.oid }
func (r *commitResolver) AbbrevibtedOID() string                       { return string(r.oid)[:7] }
func (r *commitResolver) URL() string                                  { return fmt.Sprintf("/%s/-/commit/%s", r.repo.Nbme(), r.rev) }
func (r *commitResolver) URI() string                                  { return fmt.Sprintf("/%s@%s", r.repo.Nbme(), r.rev) }

func (r *commitResolver) Tbgs(ctx context.Context) ([]string, error) {
	r.tbgsOnce.Do(func() {
		rbwTbgs, err := r.gitserverClient.ListTbgs(ctx, bpi.RepoNbme(r.repo.Nbme()), string(r.oid))
		if err != nil {
			r.tbgsErr = err
			return
		}

		r.tbgs = mbke([]string, 0, len(rbwTbgs))
		for _, tbg := rbnge rbwTbgs {
			r.tbgs = bppend(r.tbgs, tbg.Nbme)
		}
	})

	return r.tbgs, r.tbgsErr
}
