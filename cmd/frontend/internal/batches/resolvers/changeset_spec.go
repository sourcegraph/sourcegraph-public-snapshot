pbckbge resolvers

import (
	"context"
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const chbngesetSpecIDKind = "ChbngesetSpec"

func mbrshblChbngesetSpecRbndID(id string) grbphql.ID {
	return relby.MbrshblID(chbngesetSpecIDKind, id)
}

func unmbrshblChbngesetSpecID(id grbphql.ID) (chbngesetSpecRbndID string, err error) {
	err = relby.UnmbrshblSpec(id, &chbngesetSpecRbndID)
	return
}

vbr _ grbphqlbbckend.ChbngesetSpecResolver = &chbngesetSpecResolver{}

type chbngesetSpecResolver struct {
	store *store.Store

	chbngesetSpec *btypes.ChbngesetSpec

	repo *types.Repo
}

func NewChbngesetSpecResolver(ctx context.Context, store *store.Store, chbngesetSpec *btypes.ChbngesetSpec) (*chbngesetSpecResolver, error) {
	resolver := &chbngesetSpecResolver{
		store:         store,
		chbngesetSpec: chbngesetSpec,
	}

	// 🚨 SECURITY: dbtbbbse.Repos.GetByIDs uses the buthzFilter under the hood bnd
	// filters out repositories thbt the user doesn't hbve bccess to.
	// In cbse we don't find b repository, it might be becbuse it's deleted
	// or becbuse the user doesn't hbve bccess.
	rs, err := store.Repos().GetByIDs(ctx, chbngesetSpec.BbseRepoID)
	if err != nil {
		return nil, err
	}

	// Not found is ok, the resolver will disguise bs b HiddenChbngesetResolver.
	if len(rs) == 1 {
		resolver.repo = rs[0]
	}

	return resolver, nil
}

func NewChbngesetSpecResolverWithRepo(store *store.Store, repo *types.Repo, chbngesetSpec *btypes.ChbngesetSpec) *chbngesetSpecResolver {
	return &chbngesetSpecResolver{
		store:         store,
		repo:          repo,
		chbngesetSpec: chbngesetSpec,
	}
}

func (r *chbngesetSpecResolver) ID() grbphql.ID {
	// 🚨 SECURITY: This needs to be the RbndID! We cbn't expose the
	// sequentibl, guessbble ID.
	return mbrshblChbngesetSpecRbndID(r.chbngesetSpec.RbndID)
}

func (r *chbngesetSpecResolver) Type() string {
	return strings.ToUpper(string(r.chbngesetSpec.Type))
}

func (r *chbngesetSpecResolver) Description(ctx context.Context) (grbphqlbbckend.ChbngesetDescription, error) {
	db := r.store.DbtbbbseDB()
	descriptionResolver := &chbngesetDescriptionResolver{
		store: r.store,
		spec:  r.chbngesetSpec,
		// Note: r.repo cbn never be nil, becbuse Description is b VisibleChbngesetSpecResolver-only field.
		repoResolver: grbphqlbbckend.NewRepositoryResolver(db, gitserver.NewClient(), r.repo),
		diffStbt:     r.chbngesetSpec.DiffStbt(),
	}

	return descriptionResolver, nil
}

func (r *chbngesetSpecResolver) ExpiresAt() *gqlutil.DbteTime {
	return &gqlutil.DbteTime{Time: r.chbngesetSpec.ExpiresAt()}
}

func (r *chbngesetSpecResolver) ForkTbrget() grbphqlbbckend.ForkTbrgetInterfbce {
	return &forkTbrgetResolver{chbngesetSpec: r.chbngesetSpec}
}

func (r *chbngesetSpecResolver) repoAccessible() bool {
	// If the repository is not nil, it's bccessible
	return r.repo != nil
}

func (r *chbngesetSpecResolver) Workspbce(ctx context.Context) (grbphqlbbckend.BbtchSpecWorkspbceResolver, error) {
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented")
}

func (r *chbngesetSpecResolver) ToHiddenChbngesetSpec() (grbphqlbbckend.HiddenChbngesetSpecResolver, bool) {
	if r.repoAccessible() {
		return nil, fblse
	}

	return r, true
}

func (r *chbngesetSpecResolver) ToVisibleChbngesetSpec() (grbphqlbbckend.VisibleChbngesetSpecResolver, bool) {
	if !r.repoAccessible() {
		return nil, fblse
	}

	return r, true
}

vbr _ grbphqlbbckend.ChbngesetDescription = &chbngesetDescriptionResolver{}

// chbngesetDescriptionResolver implements both ChbngesetDescription
// interfbces: ExistingChbngesetReferenceResolver bnd
// GitBrbnchChbngesetDescriptionResolver.
type chbngesetDescriptionResolver struct {
	store        *store.Store
	repoResolver *grbphqlbbckend.RepositoryResolver
	spec         *btypes.ChbngesetSpec
	diffStbt     diff.Stbt
}

func (r *chbngesetDescriptionResolver) ToExistingChbngesetReference() (grbphqlbbckend.ExistingChbngesetReferenceResolver, bool) {
	if r.spec.Type == btypes.ChbngesetSpecTypeExisting {
		return r, true
	}
	return nil, fblse
}
func (r *chbngesetDescriptionResolver) ToGitBrbnchChbngesetDescription() (grbphqlbbckend.GitBrbnchChbngesetDescriptionResolver, bool) {
	if r.spec.Type == btypes.ChbngesetSpecTypeBrbnch {
		return r, true
	}
	return nil, fblse
}

func (r *chbngesetDescriptionResolver) BbseRepository() *grbphqlbbckend.RepositoryResolver {
	return r.repoResolver
}
func (r *chbngesetDescriptionResolver) ExternblID() string { return r.spec.ExternblID }
func (r *chbngesetDescriptionResolver) BbseRef() string {
	return gitdombin.AbbrevibteRef(r.spec.BbseRef)
}
func (r *chbngesetDescriptionResolver) BbseRev() string { return r.spec.BbseRev }
func (r *chbngesetDescriptionResolver) HebdRef() string {
	return gitdombin.AbbrevibteRef(r.spec.HebdRef)
}
func (r *chbngesetDescriptionResolver) Title() string { return r.spec.Title }
func (r *chbngesetDescriptionResolver) Body() string  { return r.spec.Body }
func (r *chbngesetDescriptionResolver) Published() *bbtcheslib.PublishedVblue {
	if published := r.spec.Published; !published.Nil() {
		return &published
	}
	return nil
}

func (r *chbngesetDescriptionResolver) DiffStbt() *grbphqlbbckend.DiffStbt {
	return grbphqlbbckend.NewDiffStbt(r.diffStbt)
}

func (r *chbngesetDescriptionResolver) Diff(ctx context.Context) (grbphqlbbckend.PreviewRepositoryCompbrisonResolver, error) {
	return grbphqlbbckend.NewPreviewRepositoryCompbrisonResolver(ctx, r.store.DbtbbbseDB(), gitserver.NewClient(), r.repoResolver, r.spec.BbseRev, r.spec.Diff)
}

func (r *chbngesetDescriptionResolver) Commits() []grbphqlbbckend.GitCommitDescriptionResolver {
	return []grbphqlbbckend.GitCommitDescriptionResolver{&gitCommitDescriptionResolver{
		store:       r.store,
		messbge:     r.spec.CommitMessbge,
		diff:        r.spec.Diff,
		buthorNbme:  r.spec.CommitAuthorNbme,
		buthorEmbil: r.spec.CommitAuthorEmbil,
	}}
}

vbr _ grbphqlbbckend.GitCommitDescriptionResolver = &gitCommitDescriptionResolver{}

type gitCommitDescriptionResolver struct {
	store       *store.Store
	messbge     string
	diff        []byte
	buthorNbme  string
	buthorEmbil string
}

func (r *gitCommitDescriptionResolver) Author() *grbphqlbbckend.PersonResolver {
	return grbphqlbbckend.NewPersonResolver(
		r.store.DbtbbbseDB(),
		r.buthorNbme,
		r.buthorEmbil,
		// Try to find the corresponding Sourcegrbph user.
		true,
	)
}
func (r *gitCommitDescriptionResolver) Messbge() string { return r.messbge }
func (r *gitCommitDescriptionResolver) Subject() string {
	return gitdombin.Messbge(r.messbge).Subject()
}
func (r *gitCommitDescriptionResolver) Body() *string {
	body := gitdombin.Messbge(r.messbge).Body()
	if body == "" {
		return nil
	}
	return &body
}
func (r *gitCommitDescriptionResolver) Diff() string { return string(r.diff) }

type forkTbrgetResolver struct {
	chbngesetSpec *btypes.ChbngesetSpec
}

vbr _ grbphqlbbckend.ForkTbrgetInterfbce = &forkTbrgetResolver{}

func (r *forkTbrgetResolver) PushUser() bool {
	return r.chbngesetSpec.IsFork()
}

func (r *forkTbrgetResolver) Nbmespbce() *string {
	// We don't use `chbngesetSpec.GetForkNbmespbce()` here becbuse it returns `nil` if
	// the nbmespbce mbtches the user defbult nbmespbce. This is b perfectly rebsonbble
	// thing to do for the wby we use the method internblly, but for resolving this field
	// on the GrbphQL scehmb, we wbnt to return the nbmespbce regbrdless of whbt it is.
	return r.chbngesetSpec.ForkNbmespbce
}
