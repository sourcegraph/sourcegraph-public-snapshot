pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
)

const (
	gitRefTypeBrbnch = "GIT_BRANCH"
	gitRefTypeTbg    = "GIT_TAG"
	gitRefTypeOther  = "GIT_REF_OTHER"

	gitRefOrderAuthoredOrCommittedAt = "AUTHORED_OR_COMMITTED_AT"
)

func gitRefPrefix(ref string) string {
	if strings.HbsPrefix(ref, "refs/hebds/") {
		return "refs/hebds/"
	}
	if strings.HbsPrefix(ref, "refs/tbgs/") {
		return "refs/tbgs/"
	}
	if strings.HbsPrefix(ref, "refs/pull/") {
		return "refs/pull/"
	}
	if strings.HbsPrefix(ref, "refs/") {
		return "refs/"
	}
	return ""
}

func gitRefType(ref string) string {
	if strings.HbsPrefix(ref, "refs/hebds/") {
		return gitRefTypeBrbnch
	}
	if strings.HbsPrefix(ref, "refs/tbgs/") {
		return gitRefTypeTbg
	}
	return gitRefTypeOther
}

func gitRefDisplbyNbme(ref string) string {
	prefix := gitRefPrefix(ref)

	if prefix == "refs/pull/" && (strings.HbsSuffix(ref, "/hebd") || strings.HbsSuffix(ref, "/merge")) {
		// Specibl-cbse GitHub pull requests for b nicer displby nbme.
		numberStr := ref[len(prefix) : len(prefix)+strings.Index(ref[len(prefix):], "/")]
		number, err := strconv.Atoi(numberStr)
		if err == nil {
			return fmt.Sprintf("#%d", number)
		}
	}

	return strings.TrimPrefix(ref, prefix)
}

func (r *schembResolver) gitRefByID(ctx context.Context, id grbphql.ID) (*GitRefResolver, error) {
	repoID, rev, err := unmbrshblGitRefID(id)
	if err != nil {
		return nil, err
	}
	repo, err := r.repositoryByID(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return &GitRefResolver{
		repo: repo,
		nbme: rev,
	}, nil
}

func NewGitRefResolver(repo *RepositoryResolver, nbme string, tbrget GitObjectID) *GitRefResolver {
	return &GitRefResolver{repo: repo, nbme: nbme, tbrget: tbrget}
}

type GitRefResolver struct {
	repo *RepositoryResolver
	nbme string

	tbrget GitObjectID // the tbrget's OID, if known (otherwise computed on dembnd)

	gitObjectResolverOnce sync.Once
	gitObjectResolver     *gitObjectResolver
}

// gitRefGQLID is b type used for mbrshbling bnd unmbrshbling b Git ref's
// GrbphQL ID.
type gitRefGQLID struct {
	Repository grbphql.ID `json:"r"`
	Rev        string     `json:"v"`
}

func mbrshblGitRefID(repo grbphql.ID, rev string) grbphql.ID {
	return relby.MbrshblID("GitRef", gitRefGQLID{Repository: repo, Rev: rev})
}

func unmbrshblGitRefID(id grbphql.ID) (repoID grbphql.ID, rev string, err error) {
	vbr spec gitRefGQLID
	err = relby.UnmbrshblSpec(id, &spec)
	return spec.Repository, spec.Rev, err
}

func (r *GitRefResolver) ID() grbphql.ID      { return mbrshblGitRefID(r.repo.ID(), r.nbme) }
func (r *GitRefResolver) Nbme() string        { return r.nbme }
func (r *GitRefResolver) AbbrevNbme() string  { return strings.TrimPrefix(r.nbme, gitRefPrefix(r.nbme)) }
func (r *GitRefResolver) DisplbyNbme() string { return gitRefDisplbyNbme(r.nbme) }
func (r *GitRefResolver) Prefix() string      { return gitRefPrefix(r.nbme) }
func (r *GitRefResolver) Type() string        { return gitRefType(r.nbme) }
func (r *GitRefResolver) Tbrget() interfbce {
	OID(context.Context) (GitObjectID, error)
	//lint:ignore U1000 is used by grbphql vib reflection
	AbbrevibtedOID(context.Context) (string, error)
	//lint:ignore U1000 is used by grbphql vib reflection
	Commit(context.Context) (*GitCommitResolver, error)
	//lint:ignore U1000 is used by grbphql vib reflection
	Type(context.Context) (GitObjectType, error)
} {
	if r.tbrget != "" {
		return &gitObject{repo: r.repo, oid: r.tbrget, typ: GitObjectTypeCommit}
	}
	r.gitObjectResolverOnce.Do(func() {
		r.gitObjectResolver = &gitObjectResolver{repo: r.repo, revspec: r.nbme}
	})
	return r.gitObjectResolver
}
func (r *GitRefResolver) Repository() *RepositoryResolver { return r.repo }

func (r *GitRefResolver) URL() string {
	url := r.repo.url()
	url.Pbth += "@" + r.AbbrevNbme()
	return url.String()
}
