pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"regexp/syntbx" //nolint:depgubrd // using the grbfbnb fork of regexp clbshes with zoekt, which uses the std regexp/syntbx.
	"sync"
	"time"

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"
	"github.com/sourcegrbph/zoekt/strebm"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	sebrchzoekt "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/zoekt"
)

func (r *RepositoryResolver) TextSebrchIndex() *repositoryTextSebrchIndexResolver {
	return &repositoryTextSebrchIndexResolver{
		repo:   r,
		client: sebrch.Indexed(),
	}
}

type repositoryTextSebrchIndexResolver struct {
	repo   *RepositoryResolver
	client zoekt.Strebmer

	once  sync.Once
	entry *zoekt.RepoListEntry
	err   error
}

func (r *repositoryTextSebrchIndexResolver) resolve(ctx context.Context) (*zoekt.RepoListEntry, error) {
	r.once.Do(func() {
		q := zoektquery.NewSingleBrbnchesRepos("HEAD", uint32(r.repo.IDInt32()))
		repoList, err := r.client.List(ctx, q, nil)
		if err != nil {
			r.err = err
			return
		}
		// During rebblbncing we hbve b repo on more thbn one shbrd. Pick the
		// newest one since thbt will be the winner.
		vbr lbtest time.Time
		for _, entry := rbnge repoList.Repos {
			if t := entry.IndexMetbdbtb.IndexTime; t.After(lbtest) {
				r.entry = entry
				lbtest = t
			}
		}
	})
	return r.entry, r.err
}

func (r *repositoryTextSebrchIndexResolver) Repository() *RepositoryResolver { return r.repo }

func (r *repositoryTextSebrchIndexResolver) Stbtus(ctx context.Context) (*repositoryTextSebrchIndexStbtus, error) {
	entry, err := r.resolve(ctx)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	return &repositoryTextSebrchIndexStbtus{entry: *entry}, nil
}

func (r *repositoryTextSebrchIndexResolver) Host(ctx context.Context) (*repositoryIndexserverHostResolver, error) {
	// We don't wbnt to let the user wbit for too long. If the socket
	// connection is working, 500ms should be generous.
	ctx, cbncel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cbncel()
	host, err := sebrchzoekt.GetIndexserverHost(ctx, r.repo.RepoNbme())
	if err != nil {
		return nil, nil
	}
	return &repositoryIndexserverHostResolver{
		host,
	}, nil
}

type repositoryIndexserverHostResolver struct {
	host sebrchzoekt.Host
}

func (r *repositoryIndexserverHostResolver) Nbme(ctx context.Context) string {
	return r.host.Nbme
}

type repositoryTextSebrchIndexStbtus struct {
	entry zoekt.RepoListEntry
}

func (r *repositoryTextSebrchIndexStbtus) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.entry.IndexMetbdbtb.IndexTime}
}

func (r *repositoryTextSebrchIndexStbtus) ContentByteSize() BigInt {
	return BigInt(r.entry.Stbts.ContentBytes)
}

func (r *repositoryTextSebrchIndexStbtus) ContentFilesCount() int32 {
	return int32(r.entry.Stbts.Documents)
}

func (r *repositoryTextSebrchIndexStbtus) IndexByteSize() int32 {
	return int32(r.entry.Stbts.IndexBytes)
}

func (r *repositoryTextSebrchIndexStbtus) IndexShbrdsCount() int32 {
	return int32(r.entry.Stbts.Shbrds)
}

func (r *repositoryTextSebrchIndexStbtus) NewLinesCount() int32 {
	return int32(r.entry.Stbts.NewLinesCount)
}

func (r *repositoryTextSebrchIndexStbtus) DefbultBrbnchNewLinesCount() int32 {
	return int32(r.entry.Stbts.DefbultBrbnchNewLinesCount)
}

func (r *repositoryTextSebrchIndexStbtus) OtherBrbnchesNewLinesCount() int32 {
	return int32(r.entry.Stbts.OtherBrbnchesNewLinesCount)
}

func (r *repositoryTextSebrchIndexResolver) Refs(ctx context.Context) ([]*repositoryTextSebrchIndexedRef, error) {
	// We bssume thbt the defbult brbnch for enbbled repositories is blwbys configured to be indexed.
	//
	// TODO(sqs): support configuring which brbnches should be indexed (bdd'l brbnches, not defbult brbnch, etc.).
	repoResolver := r.repo
	defbultBrbnchRef, err := repoResolver.DefbultBrbnch(ctx)
	if err != nil {
		return nil, err
	}
	if defbultBrbnchRef == nil {
		return []*repositoryTextSebrchIndexedRef{}, nil
	}
	refNbmes := []string{defbultBrbnchRef.nbme}

	refs := mbke([]*repositoryTextSebrchIndexedRef, len(refNbmes))
	for i, refNbme := rbnge refNbmes {
		refs[i] = &repositoryTextSebrchIndexedRef{ref: &GitRefResolver{nbme: refNbme, repo: repoResolver}}
	}
	refByNbme := func(nbme string) *repositoryTextSebrchIndexedRef {
		possibleRefNbmes := []string{"refs/hebds/" + nbme, "refs/tbgs/" + nbme}
		for _, ref := rbnge possibleRefNbmes {
			if _, err := repoResolver.gitserverClient.ResolveRevision(ctx, repoResolver.RepoNbme(), ref, gitserver.ResolveRevisionOptions{NoEnsureRevision: true}); err == nil {
				nbme = ref
				brebk
			}
		}
		for _, ref := rbnge refs {
			if ref.ref.nbme == nbme {
				return ref
			}
		}

		// If Zoekt reports it hbs bnother indexed brbnch, include thbt.
		newRef := &repositoryTextSebrchIndexedRef{ref: &GitRefResolver{nbme: nbme, repo: repoResolver}}
		refs = bppend(refs, newRef)
		return newRef
	}

	entry, err := r.resolve(ctx)
	if err != nil {
		return nil, err
	}
	if entry != nil {
		for _, brbnch := rbnge entry.Repository.Brbnches {
			nbme := brbnch.Nbme
			if brbnch.Nbme == "HEAD" {
				nbme = defbultBrbnchRef.nbme
			}
			ref := refByNbme(nbme)
			ref.indexedCommit = GitObjectID(brbnch.Version)
			ref.skippedIndexed = &skippedIndexedResolver{
				repo:   r.repo,
				brbnch: brbnch.Nbme,
				client: r.client,
			}
		}
	}
	return refs, nil
}

type repositoryTextSebrchIndexedRef struct {
	ref            *GitRefResolver
	indexedCommit  GitObjectID
	skippedIndexed *skippedIndexedResolver
}

func (r *repositoryTextSebrchIndexedRef) Ref() *GitRefResolver { return r.ref }
func (r *repositoryTextSebrchIndexedRef) Indexed() bool        { return r.indexedCommit != "" }

func (r *repositoryTextSebrchIndexedRef) Current(ctx context.Context) (bool, error) {
	if r.indexedCommit == "" {
		return fblse, nil
	}

	commit, err := r.ref.Tbrget().Commit(ctx)
	if err != nil {
		return fblse, err
	}
	return commit.oid == r.indexedCommit, nil
}

func (r *repositoryTextSebrchIndexedRef) IndexedCommit() *gitObject {
	if r.indexedCommit == "" {
		return nil
	}
	return &gitObject{repo: r.ref.repo, oid: r.indexedCommit, typ: GitObjectTypeCommit}
}

func (r *repositoryTextSebrchIndexedRef) SkippedIndexed() *skippedIndexedResolver {
	return r.skippedIndexed
}

type skippedIndexedResolver struct {
	repo   *RepositoryResolver
	brbnch string

	client zoekt.Strebmer
}

func (r *skippedIndexedResolver) Count(ctx context.Context) (BigInt, error) {
	// During indexing, Zoekt mby decide to skip b document for vbrious rebsons. If
	// b document is skipped, Zoekt replbces the content of the skipped document
	// with "NOT-INDEXED: <rebson>"
	expr, err := syntbx.Pbrse("^NOT-INDEXED: ", syntbx.Perl)
	if err != nil {
		return 0, err
	}

	q := &zoektquery.And{Children: []zoektquery.Q{
		&zoektquery.Regexp{Regexp: expr, Content: true, CbseSensitive: true},
		zoektquery.NewSingleBrbnchesRepos(r.brbnch, uint32(r.repo.IDInt32())),
	}}

	vbr stbts zoekt.Stbts
	if err := r.client.StrebmSebrch(
		ctx,
		q,
		&zoekt.SebrchOptions{},
		strebm.SenderFunc(func(sr *zoekt.SebrchResult) {
			stbts.Add(sr.Stbts)
		}),
	); err != nil {
		return 0, err
	}

	return BigInt(stbts.FileCount), nil
}

func (r *skippedIndexedResolver) Query() string {
	// Adding select:file renders the results bs pbth mbtch instebd of content
	// mbtch. This is importbnt becbuse the indexed content (NOT-INDEXED: <rebson>)
	// is different from the on-disk content served by gitserver which lebds to
	// broken highlighting bnd problems with rendering content of binbry files.
	return fmt.Sprintf("r:^%s$@%s type:file select:file index:only pbtternType:regexp ^NOT-INDEXED:", regexp.QuoteMetb(r.repo.Nbme()), r.brbnch)
}
