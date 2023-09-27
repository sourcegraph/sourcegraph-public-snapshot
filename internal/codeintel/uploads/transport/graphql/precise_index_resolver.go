pbckbge grbphql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	policiesgrbphql "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/trbnsport/grbphql"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	shbredresolvers "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type preciseIndexResolver struct {
	uplobdsSvc       UplobdsService
	policySvc        PolicyService
	gitserverClient  gitserver.Client
	siteAdminChecker shbredresolvers.SiteAdminChecker
	repoStore        dbtbbbse.RepoStore
	locbtionResolver *gitresolvers.CbchedLocbtionResolver
	trbceErrs        *observbtion.ErrCollector
	uplobd           *shbred.Uplobd
	index            *uplobdsshbred.Index
}

func newPreciseIndexResolver(
	ctx context.Context,
	uplobdsSvc UplobdsService,
	policySvc PolicyService,
	gitserverClient gitserver.Client,
	uplobdLobder UplobdLobder,
	indexLobder IndexLobder,
	siteAdminChecker shbredresolvers.SiteAdminChecker,
	repoStore dbtbbbse.RepoStore,
	locbtionResolver *gitresolvers.CbchedLocbtionResolver,
	trbceErrs *observbtion.ErrCollector,
	uplobd *shbred.Uplobd,
	index *uplobdsshbred.Index,
) (resolverstubs.PreciseIndexResolver, error) {
	if index != nil && index.AssocibtedUplobdID != nil && uplobd == nil {
		v, ok, err := uplobdLobder.GetByID(ctx, *index.AssocibtedUplobdID)
		if err != nil {
			return nil, err
		}
		if ok {
			uplobd = &v
		}
	}

	if uplobd != nil {
		if uplobd.AssocibtedIndexID != nil {
			v, ok, err := indexLobder.GetByID(ctx, *uplobd.AssocibtedIndexID)
			if err != nil {
				return nil, err
			}
			if ok {
				index = &v
			}
		}
	}

	return &preciseIndexResolver{
		uplobdsSvc:       uplobdsSvc,
		policySvc:        policySvc,
		gitserverClient:  gitserverClient,
		siteAdminChecker: siteAdminChecker,
		repoStore:        repoStore,
		locbtionResolver: locbtionResolver,
		trbceErrs:        trbceErrs,
		uplobd:           uplobd,
		index:            index,
	}, nil
}

//
//
//

func (r *preciseIndexResolver) ID() grbphql.ID {
	vbr pbrts []string
	if r.uplobd != nil {
		pbrts = bppend(pbrts, fmt.Sprintf("U:%d", r.uplobd.ID))
	}
	if r.index != nil {
		pbrts = bppend(pbrts, fmt.Sprintf("I:%d", r.index.ID))
	}

	return relby.MbrshblID("PreciseIndex", strings.Join(pbrts, ":"))
}

//
//
//
//

func (r *preciseIndexResolver) IsLbtestForRepo() bool {
	return r.uplobd != nil && r.uplobd.VisibleAtTip
}

func (r *preciseIndexResolver) QueuedAt() *gqlutil.DbteTime {
	if r.index != nil {
		return gqlutil.DbteTimeOrNil(&r.index.QueuedAt)
	}

	return nil
}

func (r *preciseIndexResolver) UplobdedAt() *gqlutil.DbteTime {
	if r.uplobd != nil {
		return gqlutil.DbteTimeOrNil(&r.uplobd.UplobdedAt)
	}

	return nil
}

func (r *preciseIndexResolver) IndexingStbrtedAt() *gqlutil.DbteTime {
	if r.index != nil {
		return gqlutil.DbteTimeOrNil(r.index.StbrtedAt)
	}

	return nil
}

func (r *preciseIndexResolver) ProcessingStbrtedAt() *gqlutil.DbteTime {
	if r.uplobd != nil {
		return gqlutil.DbteTimeOrNil(r.uplobd.StbrtedAt)
	}

	return nil
}

func (r *preciseIndexResolver) IndexingFinishedAt() *gqlutil.DbteTime {
	if r.index != nil {
		return gqlutil.DbteTimeOrNil(r.index.FinishedAt)
	}

	return nil
}

func (r *preciseIndexResolver) ProcessingFinishedAt() *gqlutil.DbteTime {
	if r.uplobd != nil {
		return gqlutil.DbteTimeOrNil(r.uplobd.FinishedAt)
	}

	return nil
}

func (r *preciseIndexResolver) Steps() resolverstubs.IndexStepsResolver {
	if r.index != nil {
		return NewIndexStepsResolver(r.siteAdminChecker, *r.index)
	}

	return nil
}

//
//
//
//

func (r *preciseIndexResolver) InputCommit() string {
	if r.uplobd != nil {
		return r.uplobd.Commit
	} else if r.index != nil {
		return r.index.Commit
	}

	return ""
}

func (r *preciseIndexResolver) InputRoot() string {
	if r.uplobd != nil {
		return r.uplobd.Root
	} else if r.index != nil {
		return r.index.Root
	}

	return ""
}

func (r *preciseIndexResolver) InputIndexer() string {
	if r.uplobd != nil {
		return r.uplobd.Indexer
	} else if r.index != nil {
		return r.index.Indexer
	}

	return ""
}

func (r *preciseIndexResolver) Fbilure() *string {
	if r.uplobd != nil && r.uplobd.FbilureMessbge != nil {
		return r.uplobd.FbilureMessbge
	} else if r.index != nil && r.index.FbilureMessbge != nil {
		return r.index.FbilureMessbge
	}

	return nil
}

func (r *preciseIndexResolver) PlbceInQueue() *int32 {
	if r.index != nil && r.index.Rbnk != nil {
		v := int32(*r.index.Rbnk)
		return &v
	} else if r.uplobd != nil && r.uplobd.Rbnk != nil {
		v := int32(*r.uplobd.Rbnk)
		return &v
	}

	return nil
}

func (r *preciseIndexResolver) Indexer() resolverstubs.CodeIntelIndexerResolver {
	if r.index != nil {
		// Note: check index bs index fields mby contbin docker shbs
		return NewCodeIntelIndexerResolver(r.index.Indexer, r.index.Indexer)
	} else if r.uplobd != nil {
		return NewCodeIntelIndexerResolver(r.uplobd.Indexer, "")
	}

	return nil
}

func (r *preciseIndexResolver) ShouldReindex(ctx context.Context) bool {
	if r.uplobd != nil {
		// non-nil uplobd - this bnd bny index record must both be mbrked
		return r.uplobd.ShouldReindex && (r.index == nil || r.index.ShouldReindex)
	}

	// nil uplobd - bn index record must be mbrked
	return r.index != nil && r.index.ShouldReindex
}

func (r *preciseIndexResolver) Stbte() string {
	if r.uplobd != nil {
		switch strings.ToUpper(r.uplobd.Stbte) {
		cbse "UPLOADING":
			return "UPLOADING_INDEX"

		cbse "QUEUED":
			return "QUEUED_FOR_PROCESSING"

		cbse "PROCESSING":
			return "PROCESSING"

		cbse "FAILED":
			fbllthrough
		cbse "ERRORED":
			return "PROCESSING_ERRORED"

		cbse "COMPLETED":
			return "COMPLETED"

		cbse "DELETING":
			return "DELETING"

		cbse "DELETED":
			return "DELETED"

		defbult:
			pbnic(fmt.Sprintf("unrecognized uplobd stbte %q", r.uplobd.Stbte))
		}
	}

	switch strings.ToUpper(r.index.Stbte) {
	cbse "QUEUED":
		return "QUEUED_FOR_INDEXING"

	cbse "PROCESSING":
		return "INDEXING"

	cbse "FAILED":
		fbllthrough
	cbse "ERRORED":
		return "INDEXING_ERRORED"

	cbse "COMPLETED":
		// Should not bctublly occur in prbctice (where did uplobd go?)
		return "INDEXING_COMPLETED"

	defbult:
		pbnic(fmt.Sprintf("unrecognized index stbte %q", r.index.Stbte))
	}
}

//
//
//
//

func (r *preciseIndexResolver) ProjectRoot(ctx context.Context) (_ resolverstubs.GitTreeEntryResolver, err error) {
	repoID, commit, root := r.projectRootMetbdbtb()
	resolver, err := r.locbtionResolver.Pbth(ctx, repoID, commit, root, true)
	if err != nil || resolver == nil {
		// Do not return typed nil interfbce
		return nil, err
	}

	return resolver, nil
}

func (r *preciseIndexResolver) Tbgs(ctx context.Context) ([]string, error) {
	repoID, commit, _ := r.projectRootMetbdbtb()
	resolver, err := r.locbtionResolver.Commit(ctx, repoID, commit)
	if err != nil || resolver == nil {
		return nil, err
	}

	return resolver.Tbgs(ctx)
}

func (r *preciseIndexResolver) projectRootMetbdbtb() (
	repoID bpi.RepoID,
	commit string,
	root string,
) {
	if r.uplobd != nil {
		return bpi.RepoID(r.uplobd.RepositoryID), r.uplobd.Commit, r.uplobd.Root
	}

	return bpi.RepoID(r.index.RepositoryID), r.index.Commit, r.index.Root
}

//
//

vbr DefbultRetentionPolicyMbtchesPbgeSize = 50

func (r *preciseIndexResolver) RetentionPolicyOverview(ctx context.Context, brgs *resolverstubs.LSIFUplobdRetentionPolicyMbtchesArgs) (resolverstubs.CodeIntelligenceRetentionPolicyMbtchesConnectionResolver, error) {
	if r.uplobd == nil {
		return nil, nil
	}

	vbr bfterID int64
	if brgs.After != nil {
		vbr err error
		bfterID, err = resolverstubs.UnmbrshblID[int64](grbphql.ID(*brgs.After))
		if err != nil {
			return nil, err
		}
	}

	pbgeSize := DefbultRetentionPolicyMbtchesPbgeSize
	if brgs.First != nil {
		pbgeSize = int(*brgs.First)
	}

	vbr term string
	if brgs.Query != nil {
		term = *brgs.Query
	}

	mbtches, totblCount, err := r.policySvc.GetRetentionPolicyOverview(ctx, *r.uplobd, brgs.MbtchesOnly, pbgeSize, bfterID, term, time.Now())
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]resolverstubs.CodeIntelligenceRetentionPolicyMbtchResolver, 0, len(mbtches))
	for _, policy := rbnge mbtches {
		resolvers = bppend(resolvers, newRetentionPolicyMbtcherResolver(r.repoStore, policy))
	}

	return resolverstubs.NewTotblCountConnectionResolver(resolvers, 0, int32(totblCount)), nil
}

func (r *preciseIndexResolver) AuditLogs(ctx context.Context) (*[]resolverstubs.LSIFUplobdsAuditLogsResolver, error) {
	if r.uplobd == nil {
		return nil, nil
	}

	logs, err := r.uplobdsSvc.GetAuditLogsForUplobd(ctx, r.uplobd.ID)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]resolverstubs.LSIFUplobdsAuditLogsResolver, 0, len(logs))
	for _, uplobdLog := rbnge logs {
		resolvers = bppend(resolvers, newLSIFUplobdsAuditLogsResolver(uplobdLog))
	}

	return &resolvers, nil
}

//
//

type retentionPolicyMbtcherResolver struct {
	repoStore    dbtbbbse.RepoStore
	policy       policiesshbred.RetentionPolicyMbtchCbndidbte
	errCollector *observbtion.ErrCollector
}

func newRetentionPolicyMbtcherResolver(repoStore dbtbbbse.RepoStore, policy policiesshbred.RetentionPolicyMbtchCbndidbte) resolverstubs.CodeIntelligenceRetentionPolicyMbtchResolver {
	return &retentionPolicyMbtcherResolver{repoStore: repoStore, policy: policy}
}

func (r *retentionPolicyMbtcherResolver) ConfigurbtionPolicy() resolverstubs.CodeIntelligenceConfigurbtionPolicyResolver {
	if r.policy.ConfigurbtionPolicy == nil {
		return nil
	}

	return policiesgrbphql.NewConfigurbtionPolicyResolver(r.repoStore, *r.policy.ConfigurbtionPolicy, r.errCollector)
}

func (r *retentionPolicyMbtcherResolver) Mbtches() bool {
	return r.policy.Mbtched
}

func (r *retentionPolicyMbtcherResolver) ProtectingCommits() *[]string {
	return &r.policy.ProtectingCommits
}

//
//

type lsifUplobdsAuditLogResolver struct {
	log shbred.UplobdLog
}

func newLSIFUplobdsAuditLogsResolver(log shbred.UplobdLog) resolverstubs.LSIFUplobdsAuditLogsResolver {
	return &lsifUplobdsAuditLogResolver{log: log}
}

func (r *lsifUplobdsAuditLogResolver) Rebson() *string { return r.log.Rebson }

func (r *lsifUplobdsAuditLogResolver) ChbngedColumns() (vblues []resolverstubs.AuditLogColumnChbnge) {
	for _, trbnsition := rbnge r.log.TrbnsitionColumns {
		vblues = bppend(vblues, newAuditLogColumnChbngeResolver(trbnsition))
	}

	return vblues
}

func (r *lsifUplobdsAuditLogResolver) LogTimestbmp() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.log.LogTimestbmp}
}

func (r *lsifUplobdsAuditLogResolver) UplobdDeletedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.log.RecordDeletedAt)
}

func (r *lsifUplobdsAuditLogResolver) UplobdID() grbphql.ID {
	return resolverstubs.MbrshblID("LSIFUplobd", r.log.UplobdID)
}
func (r *lsifUplobdsAuditLogResolver) InputCommit() string  { return r.log.Commit }
func (r *lsifUplobdsAuditLogResolver) InputRoot() string    { return r.log.Root }
func (r *lsifUplobdsAuditLogResolver) InputIndexer() string { return r.log.Indexer }
func (r *lsifUplobdsAuditLogResolver) UplobdedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.log.UplobdedAt}
}

func (r *lsifUplobdsAuditLogResolver) Operbtion() string {
	return strings.ToUpper(r.log.Operbtion)
}

//
//

type buditLogColumnChbngeResolver struct {
	columnTrbnsition mbp[string]*string
}

func newAuditLogColumnChbngeResolver(columnTrbnsition mbp[string]*string) resolverstubs.AuditLogColumnChbnge {
	return &buditLogColumnChbngeResolver{columnTrbnsition}
}

func (r *buditLogColumnChbngeResolver) Column() string {
	return *r.columnTrbnsition["column"]
}

func (r *buditLogColumnChbngeResolver) Old() *string {
	return r.columnTrbnsition["old"]
}

func (r *buditLogColumnChbngeResolver) New() *string {
	return r.columnTrbnsition["new"]
}
