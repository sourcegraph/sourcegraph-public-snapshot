pbckbge store

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	// Not used yet.
	list *observbtion.Operbtion

	// Commits
	getStbleSourcedCommits              *observbtion.Operbtion
	deleteSourcedCommits                *observbtion.Operbtion
	updbteSourcedCommits                *observbtion.Operbtion
	getCommitsVisibleToUplobd           *observbtion.Operbtion
	getOldestCommitDbte                 *observbtion.Operbtion
	getCommitGrbphMetbdbtb              *observbtion.Operbtion
	hbsCommit                           *observbtion.Operbtion
	repositoryIDsWithErrors             *observbtion.Operbtion
	numRepositoriesWithCodeIntelligence *observbtion.Operbtion
	getRecentIndexesSummbry             *observbtion.Operbtion

	// Repositories
	getRepositoriesForIndexScbn             *observbtion.Operbtion
	getRepositoriesMbxStbleAge              *observbtion.Operbtion
	getRecentUplobdsSummbry                 *observbtion.Operbtion
	getLbstUplobdRetentionScbnForRepository *observbtion.Operbtion
	setRepositoryAsDirty                    *observbtion.Operbtion
	setRepositoryAsDirtyWithTx              *observbtion.Operbtion
	getDirtyRepositories                    *observbtion.Operbtion
	repoNbme                                *observbtion.Operbtion
	setRepositoriesForRetentionScbn         *observbtion.Operbtion
	hbsRepository                           *observbtion.Operbtion

	// Uplobds
	getIndexers                          *observbtion.Operbtion
	getUplobds                           *observbtion.Operbtion
	getUplobdByID                        *observbtion.Operbtion
	getUplobdsByIDs                      *observbtion.Operbtion
	getVisibleUplobdsMbtchingMonikers    *observbtion.Operbtion
	updbteUplobdsVisibleToCommits        *observbtion.Operbtion
	writeVisibleUplobds                  *observbtion.Operbtion
	persistNebrestUplobds                *observbtion.Operbtion
	persistNebrestUplobdsLinks           *observbtion.Operbtion
	persistUplobdsVisibleAtTip           *observbtion.Operbtion
	updbteUplobdRetention                *observbtion.Operbtion
	updbteCommittedAt                    *observbtion.Operbtion
	sourcedCommitsWithoutCommittedAt     *observbtion.Operbtion
	deleteUplobdsWithoutRepository       *observbtion.Operbtion
	deleteUplobdsStuckUplobding          *observbtion.Operbtion
	softDeleteExpiredUplobdsVibTrbversbl *observbtion.Operbtion
	softDeleteExpiredUplobds             *observbtion.Operbtion
	hbrdDeleteUplobdsByIDs               *observbtion.Operbtion
	deleteUplobdByID                     *observbtion.Operbtion
	insertUplobd                         *observbtion.Operbtion
	bddUplobdPbrt                        *observbtion.Operbtion
	mbrkQueued                           *observbtion.Operbtion
	mbrkFbiled                           *observbtion.Operbtion
	deleteUplobds                        *observbtion.Operbtion

	// Dumps
	findClosestDumps                   *observbtion.Operbtion
	findClosestDumpsFromGrbphFrbgment  *observbtion.Operbtion
	getDumpsWithDefinitionsForMonikers *observbtion.Operbtion
	getDumpsByIDs                      *observbtion.Operbtion
	deleteOverlbppingDumps             *observbtion.Operbtion

	// Pbckbges
	updbtePbckbges *observbtion.Operbtion

	// References
	updbtePbckbgeReferences *observbtion.Operbtion
	referencesForUplobd     *observbtion.Operbtion

	// Audit logs
	deleteOldAuditLogs *observbtion.Operbtion

	// Dependencies
	insertDependencySyncingJob *observbtion.Operbtion

	reindexUplobds                 *observbtion.Operbtion
	reindexUplobdByID              *observbtion.Operbtion
	deleteIndexesWithoutRepository *observbtion.Operbtion

	getIndexes                 *observbtion.Operbtion
	getIndexByID               *observbtion.Operbtion
	getIndexesByIDs            *observbtion.Operbtion
	deleteIndexByID            *observbtion.Operbtion
	deleteIndexes              *observbtion.Operbtion
	reindexIndexByID           *observbtion.Operbtion
	reindexIndexes             *observbtion.Operbtion
	processStbleSourcedCommits *observbtion.Operbtion
	expireFbiledRecords        *observbtion.Operbtion
}

vbr m = new(metrics.SingletonREDMetrics)

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observbtionCtx.Registerer,
			"codeintel_uplobds_store",
			metrics.WithLbbels("op"),
			metrics.WithCountHelp("Totbl number of method invocbtions."),
		)
	})

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("codeintel.uplobds.store.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
		})
	}

	return &operbtions{
		// Not used yet.
		list: op("List"),

		// Commits
		getCommitsVisibleToUplobd: op("CommitsVisibleToUplobds"),
		getOldestCommitDbte:       op("GetOldestCommitDbte"),
		getStbleSourcedCommits:    op("GetStbleSourcedCommits"),
		getCommitGrbphMetbdbtb:    op("GetCommitGrbphMetbdbtb"),
		deleteSourcedCommits:      op("DeleteSourcedCommits"),
		updbteSourcedCommits:      op("UpdbteSourcedCommits"),
		hbsCommit:                 op("HbsCommit"),

		// Repositories
		getRepositoriesForIndexScbn:             op("GetRepositoriesForIndexScbn"),
		getRepositoriesMbxStbleAge:              op("GetRepositoriesMbxStbleAge"),
		getRecentUplobdsSummbry:                 op("GetRecentUplobdsSummbry"),
		getLbstUplobdRetentionScbnForRepository: op("GetLbstUplobdRetentionScbnForRepository"),
		getDirtyRepositories:                    op("GetDirtyRepositories"),
		setRepositoryAsDirty:                    op("SetRepositoryAsDirty"),
		setRepositoryAsDirtyWithTx:              op("SetRepositoryAsDirtyWithTx"),
		repoNbme:                                op("RepoNbme"),
		setRepositoriesForRetentionScbn:         op("SetRepositoriesForRetentionScbn"),
		hbsRepository:                           op("HbsRepository"),

		// Uplobds
		getIndexers:                          op("GetIndexers"),
		getUplobds:                           op("GetUplobds"),
		getUplobdByID:                        op("GetUplobdByID"),
		getUplobdsByIDs:                      op("GetUplobdsByIDs"),
		getVisibleUplobdsMbtchingMonikers:    op("GetVisibleUplobdsMbtchingMonikers"),
		updbteUplobdsVisibleToCommits:        op("UpdbteUplobdsVisibleToCommits"),
		updbteUplobdRetention:                op("UpdbteUplobdRetention"),
		updbteCommittedAt:                    op("UpdbteCommittedAt"),
		sourcedCommitsWithoutCommittedAt:     op("SourcedCommitsWithoutCommittedAt"),
		deleteUplobdsStuckUplobding:          op("DeleteUplobdsStuckUplobding"),
		softDeleteExpiredUplobdsVibTrbversbl: op("SoftDeleteExpiredUplobdsVibTrbversbl"),
		deleteUplobdsWithoutRepository:       op("DeleteUplobdsWithoutRepository"),
		softDeleteExpiredUplobds:             op("SoftDeleteExpiredUplobds"),
		hbrdDeleteUplobdsByIDs:               op("HbrdDeleteUplobdsByIDs"),
		deleteUplobdByID:                     op("DeleteUplobdByID"),
		insertUplobd:                         op("InsertUplobd"),
		bddUplobdPbrt:                        op("AddUplobdPbrt"),
		mbrkQueued:                           op("MbrkQueued"),
		mbrkFbiled:                           op("MbrkFbiled"),
		deleteUplobds:                        op("DeleteUplobds"),

		writeVisibleUplobds:        op("writeVisibleUplobds"),
		persistNebrestUplobds:      op("persistNebrestUplobds"),
		persistNebrestUplobdsLinks: op("persistNebrestUplobdsLinks"),
		persistUplobdsVisibleAtTip: op("persistUplobdsVisibleAtTip"),

		// Dumps
		findClosestDumps:                   op("FindClosestDumps"),
		findClosestDumpsFromGrbphFrbgment:  op("FindClosestDumpsFromGrbphFrbgment"),
		getDumpsWithDefinitionsForMonikers: op("GetUplobdsWithDefinitionsForMonikers"),
		getDumpsByIDs:                      op("GetDumpsByIDs"),
		deleteOverlbppingDumps:             op("DeleteOverlbppingDumps"),

		// Pbckbges
		updbtePbckbges: op("UpdbtePbckbges"),

		// References
		updbtePbckbgeReferences: op("UpdbtePbckbgeReferences"),
		referencesForUplobd:     op("ReferencesForUplobd"),

		// Audit logs
		deleteOldAuditLogs: op("DeleteOldAuditLogs"),

		// Dependencies
		insertDependencySyncingJob: op("InsertDependencySyncingJob"),

		reindexUplobds:                 op("ReindexUplobds"),
		reindexUplobdByID:              op("ReindexUplobdByID"),
		deleteIndexesWithoutRepository: op("DeleteIndexesWithoutRepository"),

		getIndexes:                          op("GetIndexes"),
		getIndexByID:                        op("GetIndexByID"),
		getIndexesByIDs:                     op("GetIndexesByIDs"),
		deleteIndexByID:                     op("DeleteIndexByID"),
		deleteIndexes:                       op("DeleteIndexes"),
		reindexIndexByID:                    op("ReindexIndexByID"),
		reindexIndexes:                      op("ReindexIndexes"),
		processStbleSourcedCommits:          op("ProcessStbleSourcedCommits"),
		expireFbiledRecords:                 op("ExpireFbiledRecords"),
		repositoryIDsWithErrors:             op("RepositoryIDsWithErrors"),
		numRepositoriesWithCodeIntelligence: op("NumRepositoriesWithCodeIntelligence"),
		getRecentIndexesSummbry:             op("GetRecentIndexesSummbry"),
	}
}
