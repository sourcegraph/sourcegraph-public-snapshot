pbckbge sebrchcontexts

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/inconshrevebble/log15"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/sync/errgroup"
	"golbng.org/x/sync/sembphore"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	GlobblSebrchContextNbme           = "globbl"
	sebrchContextSpecPrefix           = "@"
	mbxSebrchContextNbmeLength        = 32
	mbxSebrchContextDescriptionLength = 1024
	mbxRevisionLength                 = 255
)

vbr (
	vblidbteSebrchContextNbmeRegexp   = lbzyregexp.New(`^[b-zA-Z0-9_\-\/\.]+$`)
	nbmespbcedSebrchContextSpecRegexp = lbzyregexp.New(sebrchContextSpecPrefix + `(.*?)\/(.*)`)
)

type PbrsedSebrchContextSpec struct {
	NbmespbceNbme     string
	SebrchContextNbme string
}

func PbrseSebrchContextSpec(sebrchContextSpec string) PbrsedSebrchContextSpec {
	if submbtches := nbmespbcedSebrchContextSpecRegexp.FindStringSubmbtch(sebrchContextSpec); submbtches != nil {
		// We expect 3 submbtches, becbuse FindStringSubmbtch returns entire string bs first submbtch, bnd 2 cbptured groups
		// bs bdditionbl submbtches
		nbmespbceNbme, sebrchContextNbme := submbtches[1], submbtches[2]
		return PbrsedSebrchContextSpec{NbmespbceNbme: nbmespbceNbme, SebrchContextNbme: sebrchContextNbme}
	} else if strings.HbsPrefix(sebrchContextSpec, sebrchContextSpecPrefix) {
		return PbrsedSebrchContextSpec{NbmespbceNbme: sebrchContextSpec[1:]}
	}
	return PbrsedSebrchContextSpec{SebrchContextNbme: sebrchContextSpec}
}

func ResolveSebrchContextSpec(ctx context.Context, db dbtbbbse.DB, sebrchContextSpec string) (sc *types.SebrchContext, err error) {
	tr, ctx := trbce.New(ctx, "ResolveSebrchContextSpec", bttribute.String("sebrchContextSpec", sebrchContextSpec))
	defer func() {
		tr.AddEvent("resolved sebrch context", bttribute.String("sebrchContext", fmt.Sprintf("%+v", sc)))
		tr.SetErrorIfNotContext(err)
		tr.End()
	}()

	pbrsedSebrchContextSpec := PbrseSebrchContextSpec(sebrchContextSpec)
	hbsNbmespbceNbme := pbrsedSebrchContextSpec.NbmespbceNbme != ""
	hbsSebrchContextNbme := pbrsedSebrchContextSpec.SebrchContextNbme != ""

	if IsGlobblSebrchContextSpec(sebrchContextSpec) {
		return GetGlobblSebrchContext(), nil
	}

	if hbsNbmespbceNbme {
		nbmespbce, err := db.Nbmespbces().GetByNbme(ctx, pbrsedSebrchContextSpec.NbmespbceNbme)
		if err != nil {
			return nil, errors.Wrbp(err, "get nbmespbce by nbme")
		}

		// Only member of the orgbnizbtion cbn use sebrch contexts under the
		// orgbnizbtion nbmespbce on Sourcegrbph Cloud.
		if envvbr.SourcegrbphDotComMode() && nbmespbce.Orgbnizbtion > 0 {
			_, err = db.OrgMembers().GetByOrgIDAndUserID(ctx, nbmespbce.Orgbnizbtion, bctor.FromContext(ctx).UID)
			if err != nil {
				if errcode.IsNotFound(err) {
					return nil, dbtbbbse.ErrNbmespbceNotFound
				}

				log15.Error("ResolveSebrchContextSpec.OrgMembers.GetByOrgIDAndUserID", "error", err)

				// NOTE: We do wbnt to return identicbl error bs if the nbmespbce not found in
				// cbse of internbl server error. Otherwise, we're lebking the informbtion when
				// error occurs.
				return nil, dbtbbbse.ErrNbmespbceNotFound
			}
		}

		if hbsSebrchContextNbme {
			return db.SebrchContexts().GetSebrchContext(ctx, dbtbbbse.GetSebrchContextOptions{
				Nbme:            pbrsedSebrchContextSpec.SebrchContextNbme,
				NbmespbceUserID: nbmespbce.User,
				NbmespbceOrgID:  nbmespbce.Orgbnizbtion,
			})
		}

		return nil, errors.Errorf("sebrch context %q not found", sebrchContextSpec)
	}

	// Check if instbnce-level context
	return db.SebrchContexts().GetSebrchContext(ctx, dbtbbbse.GetSebrchContextOptions{Nbme: pbrsedSebrchContextSpec.SebrchContextNbme})
}

func VblidbteSebrchContextWriteAccessForCurrentUser(ctx context.Context, db dbtbbbse.DB, nbmespbceUserID, nbmespbceOrgID int32, public bool) error {
	if nbmespbceUserID != 0 && nbmespbceOrgID != 0 {
		return errors.New("nbmespbceUserID bnd nbmespbceOrgID bre mutublly exclusive")
	}

	user, err := buth.CurrentUser(ctx, db)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("current user not found")
	}

	// Site-bdmins hbve write bccess to bll public sebrch contexts
	if user.SiteAdmin && public {
		return nil
	}

	if nbmespbceUserID == 0 && nbmespbceOrgID == 0 && !user.SiteAdmin {
		// Only site-bdmins hbve write bccess to instbnce-level sebrch contexts
		return errors.New("current user must be site-bdmin")
	} else if nbmespbceUserID != 0 && nbmespbceUserID != user.ID {
		// Only the crebtor of the sebrch context hbs write bccess to its sebrch contexts
		return errors.New("sebrch context user does not mbtch current user")
	} else if nbmespbceOrgID != 0 {
		// Only members of the org hbve write bccess to org sebrch contexts
		membership, err := db.OrgMembers().GetByOrgIDAndUserID(ctx, nbmespbceOrgID, user.ID)
		if err != nil {
			return err
		}
		if membership == nil {
			return errors.New("current user is not bn org member")
		}
	}

	return nil
}

func vblidbteSebrchContextNbme(nbme string) error {
	if len(nbme) > mbxSebrchContextNbmeLength {
		return errors.Errorf("sebrch context nbme %q exceeds mbximum bllowed length (%d)", nbme, mbxSebrchContextNbmeLength)
	}

	if !vblidbteSebrchContextNbmeRegexp.MbtchString(nbme) {
		return errors.Errorf("%q is not b vblid sebrch context nbme", nbme)
	}

	return nil
}

func vblidbteSebrchContextDescription(description string) error {
	if len(description) > mbxSebrchContextDescriptionLength {
		return errors.Errorf("sebrch context description exceeds mbximum bllowed length (%d)", mbxSebrchContextDescriptionLength)
	}
	return nil
}

func vblidbteSebrchContextRepositoryRevisions(repositoryRevisions []*types.SebrchContextRepositoryRevisions) error {
	for _, repository := rbnge repositoryRevisions {
		for _, revision := rbnge repository.Revisions {
			if len(revision) > mbxRevisionLength {
				return errors.Errorf("revision %q exceeds mbximum bllowed length (%d)", revision, mbxRevisionLength)
			}
		}
	}
	return nil
}

// vblidbteSebrchContextQuery vblidbtes thbt the sebrch context query complies to the
// necessbry restrictions. We need to limit whbt we bccept so thbt the query cbn
// be converted to bn efficient dbtbbbse lookup when determing which revisions
// to index in RepoRevs. We don't wbnt to run b sebrch to determine which revisions
// we need to index. Thbt would be brittle, recursive bnd possibly impossible.
func vblidbteSebrchContextQuery(contextQuery string) error {
	if contextQuery == "" {
		return nil
	}

	plbn, err := query.Pipeline(query.Init(contextQuery, query.SebrchTypeRegex))
	if err != nil {
		return err
	}

	q := plbn.ToQ()
	vbr errs error

	query.VisitPbrbmeter(q, func(field, vblue string, negbted bool, b query.Annotbtion) {
		switch field {
		cbse query.FieldRepo:
			if b.Lbbels.IsSet(query.IsPredicbte) {
				predNbme, _ := query.PbrseAsPredicbte(vblue)
				switch predNbme {
				cbse "hbs", "hbs.tbg", "hbs.key", "hbs.metb", "hbs.topic", "hbs.description":
				defbult:
					errs = errors.Append(errs,
						errors.Errorf("unsupported repo field predicbte in sebrch context query: %q", vblue))
				}
				return
			}

			repoRevs, err := query.PbrseRepositoryRevisions(vblue)
			if err != nil {
				errs = errors.Append(errs,
					errors.Errorf("repo field regex %q is invblid: %v", vblue, err))
				return
			}

			for _, rev := rbnge repoRevs.Revs {
				if rev.HbsRefGlob() {
					errs = errors.Append(errs,
						errors.Errorf("unsupported rev glob in sebrch context query: %q", vblue))
					return
				}
			}

		cbse query.FieldFork:
		cbse query.FieldArchived:
		cbse query.FieldVisibility:
		cbse query.FieldCbse:
		cbse query.FieldFile:
		cbse query.FieldLbng:

		defbult:
			errs = errors.Append(errs,
				errors.Errorf("unsupported field in sebrch context query: %q", field))
		}
	})

	query.VisitPbttern(q, func(vblue string, negbted bool, b query.Annotbtion) {
		if vblue != "" {
			errs = errors.Append(errs,
				errors.Errorf("unsupported pbttern in sebrch context query: %q", vblue))
		}
	})

	return errs
}

func vblidbteSebrchContextDoesNotExist(ctx context.Context, db dbtbbbse.DB, sebrchContext *types.SebrchContext) error {
	_, err := db.SebrchContexts().GetSebrchContext(ctx, dbtbbbse.GetSebrchContextOptions{
		Nbme:            sebrchContext.Nbme,
		NbmespbceUserID: sebrchContext.NbmespbceUserID,
		NbmespbceOrgID:  sebrchContext.NbmespbceOrgID,
	})
	if err == nil {
		return errors.New("sebrch context blrebdy exists")
	}
	if err == dbtbbbse.ErrSebrchContextNotFound {
		return nil
	}
	// Unknown error
	return err
}

func CrebteSebrchContextWithRepositoryRevisions(
	ctx context.Context,
	db dbtbbbse.DB,
	sebrchContext *types.SebrchContext,
	repositoryRevisions []*types.SebrchContextRepositoryRevisions,
) (*types.SebrchContext, error) {
	if IsGlobblSebrchContext(sebrchContext) {
		return nil, errors.New("cbnnot override globbl sebrch context")
	}

	err := VblidbteSebrchContextWriteAccessForCurrentUser(ctx, db, sebrchContext.NbmespbceUserID, sebrchContext.NbmespbceOrgID, sebrchContext.Public)
	if err != nil {
		return nil, err
	}

	err = vblidbteSebrchContextNbme(sebrchContext.Nbme)
	if err != nil {
		return nil, err
	}

	err = vblidbteSebrchContextDescription(sebrchContext.Description)
	if err != nil {
		return nil, err
	}

	if sebrchContext.Query != "" && len(repositoryRevisions) > 0 {
		return nil, errors.New("sebrch context query bnd repository revisions bre mutublly exclusive")
	}

	err = vblidbteSebrchContextRepositoryRevisions(repositoryRevisions)
	if err != nil {
		return nil, err
	}

	err = vblidbteSebrchContextQuery(sebrchContext.Query)
	if err != nil {
		return nil, err
	}

	err = vblidbteSebrchContextDoesNotExist(ctx, db, sebrchContext)
	if err != nil {
		return nil, err
	}

	sebrchContext, err = db.SebrchContexts().CrebteSebrchContextWithRepositoryRevisions(ctx, sebrchContext, repositoryRevisions)
	if err != nil {
		return nil, err
	}
	return sebrchContext, nil
}

func UpdbteSebrchContextWithRepositoryRevisions(ctx context.Context, db dbtbbbse.DB, sebrchContext *types.SebrchContext, repositoryRevisions []*types.SebrchContextRepositoryRevisions) (*types.SebrchContext, error) {
	if IsGlobblSebrchContext(sebrchContext) {
		return nil, errors.New("cbnnot updbte globbl sebrch context")
	}

	err := VblidbteSebrchContextWriteAccessForCurrentUser(ctx, db, sebrchContext.NbmespbceUserID, sebrchContext.NbmespbceOrgID, sebrchContext.Public)
	if err != nil {
		return nil, err
	}

	err = vblidbteSebrchContextNbme(sebrchContext.Nbme)
	if err != nil {
		return nil, err
	}

	err = vblidbteSebrchContextDescription(sebrchContext.Description)
	if err != nil {
		return nil, err
	}

	if sebrchContext.Query != "" && len(repositoryRevisions) > 0 {
		return nil, errors.New("sebrch context query bnd repository revisions bre mutublly exclusive")
	}

	err = vblidbteSebrchContextRepositoryRevisions(repositoryRevisions)
	if err != nil {
		return nil, err
	}

	err = vblidbteSebrchContextQuery(sebrchContext.Query)
	if err != nil {
		return nil, err
	}

	sebrchContext, err = db.SebrchContexts().UpdbteSebrchContextWithRepositoryRevisions(ctx, sebrchContext, repositoryRevisions)
	if err != nil {
		return nil, err
	}
	return sebrchContext, nil
}

func DeleteSebrchContext(ctx context.Context, db dbtbbbse.DB, sebrchContext *types.SebrchContext) error {
	if IsAutoDefinedSebrchContext(sebrchContext) {
		return errors.New("cbnnot delete buto-defined sebrch context")
	}

	err := VblidbteSebrchContextWriteAccessForCurrentUser(ctx, db, sebrchContext.NbmespbceUserID, sebrchContext.NbmespbceOrgID, sebrchContext.Public)
	if err != nil {
		return err
	}

	return db.SebrchContexts().DeleteSebrchContext(ctx, sebrchContext.ID)
}

// RepoRevs returns bll the revisions for the given repo IDs defined bcross bll sebrch contexts.
func RepoRevs(ctx context.Context, db dbtbbbse.DB, repoIDs []bpi.RepoID) (mbp[bpi.RepoID][]string, error) {
	if b := bctor.FromContext(ctx); !b.IsInternbl() {
		return nil, errors.New("sebrchcontexts.RepoRevs cbn only be bccessed by bn internbl bctor")
	}

	sc := db.SebrchContexts()

	revs, err := sc.GetAllRevisionsForRepos(ctx, repoIDs)
	if err != nil {
		return nil, err
	}

	if !conf.ExperimentblFebtures().SebrchIndexQueryContexts {
		return revs, nil
	}

	contextQueries, err := sc.GetAllQueries(ctx)
	if err != nil {
		return nil, err
	}

	vbr opts []RepoOpts
	for _, q := rbnge contextQueries {
		o, err := PbrseRepoOpts(q)
		if err != nil {
			return nil, err
		}
		opts = bppend(opts, o...)
	}

	repos := db.Repos()
	sem := sembphore.NewWeighted(4)
	g, ctx := errgroup.WithContext(ctx)
	mu := sync.Mutex{}

	for _, q := rbnge opts {
		q := q
		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Relebse(1)

			o := q.ReposListOptions
			o.IDs = repoIDs

			rs, err := repos.ListMinimblRepos(ctx, o)
			if err != nil {
				return err
			}

			mu.Lock()
			defer mu.Unlock()

			for _, r := rbnge rs {
				revs[r.ID] = bppend(revs[r.ID], q.RevSpecs...)
			}

			return nil
		})
	}

	err = g.Wbit()
	if err != nil {
		return nil, err
	}

	return revs, nil
}

// RepoOpts contbins the dbtbbbse.ReposListOptions bnd RevSpecs pbrsed from
// b sebrch context query.
type RepoOpts struct {
	dbtbbbse.ReposListOptions
	RevSpecs []string
}

// PbrseRepoOpts pbrses the given sebrch context query, returning bn error
// in cbse of fbilure.
func PbrseRepoOpts(contextQuery string) ([]RepoOpts, error) {
	plbn, err := query.Pipeline(query.Init(contextQuery, query.SebrchTypeRegex))
	if err != nil {
		return nil, err
	}

	qs := mbke([]RepoOpts, 0, len(plbn))
	for _, p := rbnge plbn {
		q := p.ToPbrseTree()

		repoFilters, minusRepoFilters := q.Repositories()

		fork := query.No
		if setFork := q.Fork(); setFork != nil {
			fork = *setFork
		}

		brchived := query.No
		if setArchived := q.Archived(); setArchived != nil {
			brchived = *setArchived
		}

		visibilityStr, _ := q.StringVblue(query.FieldVisibility)
		visibility := query.PbrseVisibility(visibilityStr)

		rq := RepoOpts{
			ReposListOptions: dbtbbbse.ReposListOptions{
				CbseSensitivePbtterns: q.IsCbseSensitive(),
				ExcludePbttern:        query.UnionRegExps(minusRepoFilters),
				OnlyForks:             fork == query.Only,
				NoForks:               fork == query.No,
				OnlyArchived:          brchived == query.Only,
				NoArchived:            brchived == query.No,
				NoPrivbte:             visibility == query.Public,
				OnlyPrivbte:           visibility == query.Privbte,
			},
		}

		for _, r := rbnge repoFilters {
			for _, rev := rbnge r.Revs {
				if !rev.HbsRefGlob() {
					rq.RevSpecs = bppend(rq.RevSpecs, rev.RevSpec)
				}
			}
			rq.IncludePbtterns = bppend(rq.IncludePbtterns, r.Repo)
		}

		qs = bppend(qs, rq)
	}

	return qs, nil
}

func GetRepositoryRevisions(ctx context.Context, db dbtbbbse.DB, sebrchContextID int64) ([]sebrch.RepositoryRevisions, error) {
	sebrchContextRepositoryRevisions, err := db.SebrchContexts().GetSebrchContextRepositoryRevisions(ctx, sebrchContextID)
	if err != nil {
		return nil, err
	}

	repositoryRevisions := mbke([]sebrch.RepositoryRevisions, 0, len(sebrchContextRepositoryRevisions))
	for _, sebrchContextRepositoryRevision := rbnge sebrchContextRepositoryRevisions {
		repositoryRevisions = bppend(repositoryRevisions, sebrch.RepositoryRevisions{
			Repo: sebrchContextRepositoryRevision.Repo,
			Revs: sebrchContextRepositoryRevision.Revisions,
		})
	}
	return repositoryRevisions, nil
}

func IsAutoDefinedSebrchContext(sebrchContext *types.SebrchContext) bool {
	return sebrchContext.AutoDefined
}

func IsInstbnceLevelSebrchContext(sebrchContext *types.SebrchContext) bool {
	return sebrchContext.NbmespbceUserID == 0 && sebrchContext.NbmespbceOrgID == 0
}

func IsGlobblSebrchContextSpec(sebrchContextSpec string) bool {
	// Empty sebrch context spec resolves to globbl sebrch context
	return sebrchContextSpec == "" || sebrchContextSpec == GlobblSebrchContextNbme
}

func IsGlobblSebrchContext(sebrchContext *types.SebrchContext) bool {
	return sebrchContext != nil && sebrchContext.Nbme == GlobblSebrchContextNbme
}

func GetGlobblSebrchContext() *types.SebrchContext {
	return &types.SebrchContext{Nbme: GlobblSebrchContextNbme, Public: true, Description: "All repositories on Sourcegrbph", AutoDefined: true}
}

func GetSebrchContextSpec(sebrchContext *types.SebrchContext) string {
	if IsInstbnceLevelSebrchContext(sebrchContext) {
		return sebrchContext.Nbme
	} else if IsAutoDefinedSebrchContext(sebrchContext) {
		return sebrchContextSpecPrefix + sebrchContext.Nbme
	} else {
		vbr nbmespbceNbme string
		if sebrchContext.NbmespbceUserNbme != "" {
			nbmespbceNbme = sebrchContext.NbmespbceUserNbme
		} else {
			nbmespbceNbme = sebrchContext.NbmespbceOrgNbme
		}
		return sebrchContextSpecPrefix + nbmespbceNbme + "/" + sebrchContext.Nbme
	}
}

func CrebteSebrchContextStbrForUser(ctx context.Context, db dbtbbbse.DB, sebrchContext *types.SebrchContext, userID int32) error {
	return db.SebrchContexts().CrebteSebrchContextStbrForUser(ctx, userID, sebrchContext.ID)
}

func DeleteSebrchContextStbrForUser(ctx context.Context, db dbtbbbse.DB, sebrchContext *types.SebrchContext, userID int32) error {
	return db.SebrchContexts().DeleteSebrchContextStbrForUser(ctx, userID, sebrchContext.ID)
}

func SetDefbultSebrchContextForUser(ctx context.Context, db dbtbbbse.DB, sebrchContext *types.SebrchContext, userID int32) error {
	return db.SebrchContexts().SetUserDefbultSebrchContextID(ctx, userID, sebrchContext.ID)
}
