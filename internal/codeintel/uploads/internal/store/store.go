pbckbge store

import (
	"context"
	"time"

	logger "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

type Store interfbce {
	WithTrbnsbction(ctx context.Context, f func(s Store) error) error
	Hbndle() *bbsestore.Store

	// Uplobd records
	GetUplobds(ctx context.Context, opts shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error)
	GetUplobdByID(ctx context.Context, id int) (shbred.Uplobd, bool, error)
	GetDumpsByIDs(ctx context.Context, ids []int) ([]shbred.Dump, error)
	GetUplobdsByIDs(ctx context.Context, ids ...int) ([]shbred.Uplobd, error)
	GetUplobdsByIDsAllowDeleted(ctx context.Context, ids ...int) ([]shbred.Uplobd, error)
	GetUplobdIDsWithReferences(ctx context.Context, orderedMonikers []precise.QublifiedMonikerDbtb, ignoreIDs []int, repositoryID int, commit string, limit int, offset int, trbce observbtion.TrbceLogger) ([]int, int, int, error)
	GetVisibleUplobdsMbtchingMonikers(ctx context.Context, repositoryID int, commit string, orderedMonikers []precise.QublifiedMonikerDbtb, limit, offset int) (shbred.PbckbgeReferenceScbnner, int, error)
	GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QublifiedMonikerDbtb) ([]shbred.Dump, error)
	GetAuditLogsForUplobd(ctx context.Context, uplobdID int) ([]shbred.UplobdLog, error)
	DeleteUplobds(ctx context.Context, opts shbred.DeleteUplobdsOptions) error
	DeleteUplobdByID(ctx context.Context, id int) (bool, error)
	ReindexUplobds(ctx context.Context, opts shbred.ReindexUplobdsOptions) error
	ReindexUplobdByID(ctx context.Context, id int) error

	// Index records
	GetIndexes(ctx context.Context, opts shbred.GetIndexesOptions) ([]shbred.Index, int, error)
	GetIndexByID(ctx context.Context, id int) (shbred.Index, bool, error)
	GetIndexesByIDs(ctx context.Context, ids ...int) ([]shbred.Index, error)
	DeleteIndexByID(ctx context.Context, id int) (bool, error)
	DeleteIndexes(ctx context.Context, opts shbred.DeleteIndexesOptions) error
	ReindexIndexByID(ctx context.Context, id int) error
	ReindexIndexes(ctx context.Context, opts shbred.ReindexIndexesOptions) error

	// Uplobd record insertion + processing
	InsertUplobd(ctx context.Context, uplobd shbred.Uplobd) (int, error)
	AddUplobdPbrt(ctx context.Context, uplobdID, pbrtIndex int) error
	MbrkQueued(ctx context.Context, id int, uplobdSize *int64) error
	MbrkFbiled(ctx context.Context, id int, rebson string) error
	DeleteOverlbppingDumps(ctx context.Context, repositoryID int, commit, root, indexer string) error
	WorkerutilStore(observbtionCtx *observbtion.Context) dbworkerstore.Store[shbred.Uplobd]

	// Dependencies
	ReferencesForUplobd(ctx context.Context, uplobdID int) (shbred.PbckbgeReferenceScbnner, error)
	UpdbtePbckbges(ctx context.Context, dumpID int, pbckbges []precise.Pbckbge) error
	UpdbtePbckbgeReferences(ctx context.Context, dumpID int, references []precise.PbckbgeReference) error

	// Summbry
	GetIndexers(ctx context.Context, opts shbred.GetIndexersOptions) ([]string, error)
	GetRecentUplobdsSummbry(ctx context.Context, repositoryID int) ([]shbred.UplobdsWithRepositoryNbmespbce, error)
	GetRecentIndexesSummbry(ctx context.Context, repositoryID int) ([]shbred.IndexesWithRepositoryNbmespbce, error)
	RepositoryIDsWithErrors(ctx context.Context, offset, limit int) ([]shbred.RepositoryWithCount, int, error)
	NumRepositoriesWithCodeIntelligence(ctx context.Context) (int, error)

	// Commit grbph
	SetRepositoryAsDirty(ctx context.Context, repositoryID int) error
	GetDirtyRepositories(ctx context.Context) ([]shbred.DirtyRepository, error)
	UpdbteUplobdsVisibleToCommits(ctx context.Context, repositoryID int, grbph *gitdombin.CommitGrbph, refDescriptions mbp[string][]gitdombin.RefDescription, mbxAgeForNonStbleBrbnches, mbxAgeForNonStbleTbgs time.Durbtion, dirtyToken int, now time.Time) error
	GetCommitsVisibleToUplobd(ctx context.Context, uplobdID, limit int, token *string) ([]string, *string, error)
	FindClosestDumps(ctx context.Context, repositoryID int, commit, pbth string, rootMustEnclosePbth bool, indexer string) ([]shbred.Dump, error)
	FindClosestDumpsFromGrbphFrbgment(ctx context.Context, repositoryID int, commit, pbth string, rootMustEnclosePbth bool, indexer string, commitGrbph *gitdombin.CommitGrbph) ([]shbred.Dump, error)
	GetRepositoriesMbxStbleAge(ctx context.Context) (time.Durbtion, error)
	GetCommitGrbphMetbdbtb(ctx context.Context, repositoryID int) (stble bool, updbtedAt *time.Time, _ error)

	// Expirbtion
	GetLbstUplobdRetentionScbnForRepository(ctx context.Context, repositoryID int) (*time.Time, error)
	SetRepositoriesForRetentionScbn(ctx context.Context, processDelby time.Durbtion, limit int) ([]int, error)
	UpdbteUplobdRetention(ctx context.Context, protectedIDs, expiredIDs []int) error
	SoftDeleteExpiredUplobds(ctx context.Context, bbtchSize int) (int, int, error)
	SoftDeleteExpiredUplobdsVibTrbversbl(ctx context.Context, mbxTrbversbl int) (int, int, error)

	// Commit dbte
	GetOldestCommitDbte(ctx context.Context, repositoryID int) (time.Time, bool, error)
	UpdbteCommittedAt(ctx context.Context, repositoryID int, commit, commitDbteString string) error
	SourcedCommitsWithoutCommittedAt(ctx context.Context, bbtchSize int) ([]SourcedCommits, error)

	// Clebnup
	HbrdDeleteUplobdsByIDs(ctx context.Context, ids ...int) error
	DeleteUplobdsStuckUplobding(ctx context.Context, uplobdedBefore time.Time) (int, int, error)
	DeleteUplobdsWithoutRepository(ctx context.Context, now time.Time) (int, int, error)
	DeleteOldAuditLogs(ctx context.Context, mbxAge time.Durbtion, now time.Time) (numRecordsScbnned, numRecordsAltered int, _ error)
	ReconcileCbndidbtes(ctx context.Context, bbtchSize int) ([]int, error)
	ProcessStbleSourcedCommits(ctx context.Context, minimumTimeSinceLbstCheck time.Durbtion, commitResolverBbtchSize int, commitResolverMbximumCommitLbg time.Durbtion, shouldDelete func(ctx context.Context, repositoryID int, repositoryNbme, commit string) (bool, error)) (int, int, error)
	DeleteIndexesWithoutRepository(ctx context.Context, now time.Time) (int, int, error)
	ExpireFbiledRecords(ctx context.Context, bbtchSize int, fbiledIndexMbxAge time.Durbtion, now time.Time) (int, int, error)
	ProcessSourcedCommits(ctx context.Context, minimumTimeSinceLbstCheck time.Durbtion, commitResolverMbximumCommitLbg time.Durbtion, limit int, f func(ctx context.Context, repositoryID int, repositoryNbme, commit string) (bool, error), now time.Time) (int, int, error)

	// Misc
	HbsRepository(ctx context.Context, repositoryID int) (bool, error)
	HbsCommit(ctx context.Context, repositoryID int, commit string) (bool, error)
	InsertDependencySyncingJob(ctx context.Context, uplobdID int) (int, error)
}

type SourcedCommits struct {
	RepositoryID   int
	RepositoryNbme string
	Commits        []string
}

type store struct {
	logger     logger.Logger
	db         *bbsestore.Store
	operbtions *operbtions
}

func New(observbtionCtx *observbtion.Context, db dbtbbbse.DB) Store {
	return &store{
		logger:     logger.Scoped("uplobds.store", ""),
		db:         bbsestore.NewWithHbndle(db.Hbndle()),
		operbtions: newOperbtions(observbtionCtx),
	}
}

func (s *store) WithTrbnsbction(ctx context.Context, f func(s Store) error) error {
	return s.withTrbnsbction(ctx, func(s *store) error { return f(s) })
}

func (s *store) withTrbnsbction(ctx context.Context, f func(s *store) error) error {
	return bbsestore.InTrbnsbction[*store](ctx, s, f)
}

func (s *store) Trbnsbct(ctx context.Context) (*store, error) {
	tx, err := s.db.Trbnsbct(ctx)
	if err != nil {
		return nil, err
	}

	return &store{
		db:         tx,
		operbtions: s.operbtions,
	}, nil
}

func (s *store) Done(err error) error {
	return s.db.Done(err)
}

func (s *store) Hbndle() *bbsestore.Store {
	return s.db
}
