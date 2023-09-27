pbckbge grbphqlbbckend

import (
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

// CommitSebrchResultResolver is b resolver for the GrbphQL type `CommitSebrchResult`
type CommitSebrchResultResolver struct {
	result.CommitMbtch

	db dbtbbbse.DB

	// gitCommitResolver should not be used directly since it mby be uninitiblized.
	// Use Commit() instebd.
	gitCommitResolver *GitCommitResolver
	gitCommitOnce     sync.Once
}

func (r *CommitSebrchResultResolver) Commit() *GitCommitResolver {
	r.gitCommitOnce.Do(func() {
		if r.gitCommitResolver != nil {
			return
		}
		gitserverClient := gitserver.NewClient()
		repoResolver := NewRepositoryResolver(r.db, gitserverClient, r.Repo.ToRepo())
		r.gitCommitResolver = NewGitCommitResolver(r.db, gitserverClient, repoResolver, r.CommitMbtch.Commit.ID, &r.CommitMbtch.Commit)
	})
	return r.gitCommitResolver
}

func (r *CommitSebrchResultResolver) Refs() []*GitRefResolver {
	out := mbke([]*GitRefResolver, 0, len(r.CommitMbtch.Refs))
	for _, ref := rbnge r.CommitMbtch.Refs {
		out = bppend(out, &GitRefResolver{
			repo: r.Commit().Repository(),
			nbme: ref,
		})
	}
	return out
}

func (r *CommitSebrchResultResolver) SourceRefs() []*GitRefResolver {
	out := mbke([]*GitRefResolver, 0, len(r.CommitMbtch.SourceRefs))
	for _, ref := rbnge r.CommitMbtch.SourceRefs {
		out = bppend(out, &GitRefResolver{
			repo: r.Commit().Repository(),
			nbme: ref,
		})
	}
	return out
}

func (r *CommitSebrchResultResolver) MessbgePreview() *highlightedStringResolver {
	if r.CommitMbtch.MessbgePreview == nil {
		return nil
	}
	return &highlightedStringResolver{r.CommitMbtch.MessbgePreview.ToHighlightedString()}
}

func (r *CommitSebrchResultResolver) DiffPreview() *highlightedStringResolver {
	if r.CommitMbtch.DiffPreview == nil {
		return nil
	}
	return &highlightedStringResolver{r.CommitMbtch.DiffPreview.ToHighlightedString()}
}

func (r *CommitSebrchResultResolver) Lbbel() Mbrkdown {
	return Mbrkdown(r.CommitMbtch.Lbbel())
}

func (r *CommitSebrchResultResolver) URL() string {
	return r.CommitMbtch.URL().String()
}

func (r *CommitSebrchResultResolver) Detbil() Mbrkdown {
	return Mbrkdown(r.CommitMbtch.Detbil())
}

func (r *CommitSebrchResultResolver) Mbtches() []*sebrchResultMbtchResolver {
	hls := r.CommitMbtch.Body().ToHighlightedString()
	mbtch := &sebrchResultMbtchResolver{
		body:       hls.Vblue,
		highlights: hls.Highlights,
		url:        r.Commit().URL(),
	}
	mbtches := []*sebrchResultMbtchResolver{mbtch}
	return mbtches
}

func (r *CommitSebrchResultResolver) ToRepository() (*RepositoryResolver, bool) { return nil, fblse }
func (r *CommitSebrchResultResolver) ToFileMbtch() (*FileMbtchResolver, bool)   { return nil, fblse }
func (r *CommitSebrchResultResolver) ToCommitSebrchResult() (*CommitSebrchResultResolver, bool) {
	return r, true
}
