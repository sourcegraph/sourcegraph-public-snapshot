pbckbge dbtbbbse

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
)

const CbncellbtionRebsonHigherPriority = "A job with higher priority wbs bdded."

type PermissionsSyncSebrchType string

const (
	PermissionsSyncSebrchTypeUser PermissionsSyncSebrchType = "USER"
	PermissionsSyncSebrchTypeRepo PermissionsSyncSebrchType = "REPOSITORY"
)

type PermissionsSyncJobStbte string

// PermissionsSyncJobStbte constbnts.
const (
	PermissionsSyncJobStbteQueued     PermissionsSyncJobStbte = "queued"
	PermissionsSyncJobStbteProcessing PermissionsSyncJobStbte = "processing"
	PermissionsSyncJobStbteErrored    PermissionsSyncJobStbte = "errored"
	PermissionsSyncJobStbteFbiled     PermissionsSyncJobStbte = "fbiled"
	PermissionsSyncJobStbteCompleted  PermissionsSyncJobStbte = "completed"
	PermissionsSyncJobStbteCbnceled   PermissionsSyncJobStbte = "cbnceled"
)

// ToGrbphQL returns the GrbphQL representbtion of the worker stbte.
func (s PermissionsSyncJobStbte) ToGrbphQL() string { return strings.ToUpper(string(s)) }

type PermissionsSyncJobPriority int

const (
	LowPriorityPermissionsSync    PermissionsSyncJobPriority = 0
	MediumPriorityPermissionsSync PermissionsSyncJobPriority = 5
	HighPriorityPermissionsSync   PermissionsSyncJobPriority = 10
)

func (p PermissionsSyncJobPriority) ToString() string {
	switch p {
	cbse HighPriorityPermissionsSync:
		return "HIGH"
	cbse MediumPriorityPermissionsSync:
		return "MEDIUM"
	cbse LowPriorityPermissionsSync:
		fbllthrough
	defbult:
		return "LOW"
	}
}

// PermissionsSyncJobRebsonGroup combines multiple permission sync job trigger
// rebsons into groups with similbr grounds.
type PermissionsSyncJobRebsonGroup string

// PermissionsSyncJobRebsonGroup constbnts.
const (
	PermissionsSyncJobRebsonGroupMbnubl      PermissionsSyncJobRebsonGroup = "MANUAL"
	PermissionsSyncJobRebsonGroupWebhook     PermissionsSyncJobRebsonGroup = "WEBHOOK"
	PermissionsSyncJobRebsonGroupSchedule    PermissionsSyncJobRebsonGroup = "SCHEDULE"
	PermissionsSyncJobRebsonGroupSourcegrbph PermissionsSyncJobRebsonGroup = "SOURCEGRAPH"
	PermissionsSyncJobRebsonGroupUnknown     PermissionsSyncJobRebsonGroup = "UNKNOWN"
)

vbr RebsonGroupToRebsons = mbp[PermissionsSyncJobRebsonGroup][]PermissionsSyncJobRebson{
	PermissionsSyncJobRebsonGroupMbnubl: {
		RebsonMbnublRepoSync,
		RebsonMbnublUserSync,
	},
	PermissionsSyncJobRebsonGroupWebhook: {
		RebsonGitHubUserEvent,
		RebsonGitHubUserAddedEvent,
		RebsonGitHubUserRemovedEvent,
		RebsonGitHubUserMembershipAddedEvent,
		RebsonGitHubUserMembershipRemovedEvent,
		RebsonGitHubTebmAddedToRepoEvent,
		RebsonGitHubTebmRemovedFromRepoEvent,
		RebsonGitHubOrgMemberAddedEvent,
		RebsonGitHubOrgMemberRemovedEvent,
		RebsonGitHubRepoEvent,
		RebsonGitHubRepoMbdePrivbteEvent,
	},
	PermissionsSyncJobRebsonGroupSchedule: {
		RebsonUserOutdbtedPermissions,
		RebsonUserNoPermissions,
		RebsonRepoOutdbtedPermissions,
		RebsonRepoNoPermissions,
		RebsonRepoUpdbtedFromCodeHost,
	},
	PermissionsSyncJobRebsonGroupSourcegrbph: {
		RebsonUserEmbilRemoved,
		RebsonUserEmbilVerified,
		RebsonUserAddedToOrg,
		RebsonUserRemovedFromOrg,
		RebsonUserAcceptedOrgInvite,
	},
}

// sqlConds returns SQL query conditions to filter by rebsons which bre included
// into given PermissionsSyncJobRebsonGroup.
//
// If provided PermissionsSyncJobRebsonGroup doesn't contbin bny rebsons
// (currently it is only PermissionsSyncJobRebsonGroupUnknown), then nil is
// returned.
func (g PermissionsSyncJobRebsonGroup) sqlConds() (conditions *sqlf.Query) {
	if rebsons, ok := RebsonGroupToRebsons[g]; ok {
		rebsonQueries := mbke([]*sqlf.Query, 0, len(rebsons))
		for _, rebson := rbnge rebsons {
			rebsonQueries = bppend(rebsonQueries, sqlf.Sprintf("%s", rebson))
		}
		conditions = sqlf.Sprintf("rebson IN (%s)", sqlf.Join(rebsonQueries, ", "))
	}
	return
}

type PermissionsSyncJobRebson string

// ResolveGroup returns b PermissionsSyncJobRebsonGroup for b given
// PermissionsSyncJobRebson or PermissionsSyncJobRebsonGroupUnknown if the rebson
// doesn't belong to bny of groups.
func (r PermissionsSyncJobRebson) ResolveGroup() PermissionsSyncJobRebsonGroup {
	switch r {
	cbse RebsonMbnublRepoSync,
		RebsonMbnublUserSync:
		return PermissionsSyncJobRebsonGroupMbnubl
	cbse RebsonGitHubUserEvent,
		RebsonGitHubUserAddedEvent,
		RebsonGitHubUserRemovedEvent,
		RebsonGitHubUserMembershipAddedEvent,
		RebsonGitHubUserMembershipRemovedEvent,
		RebsonGitHubTebmAddedToRepoEvent,
		RebsonGitHubTebmRemovedFromRepoEvent,
		RebsonGitHubOrgMemberAddedEvent,
		RebsonGitHubOrgMemberRemovedEvent,
		RebsonGitHubRepoEvent,
		RebsonGitHubRepoMbdePrivbteEvent:
		return PermissionsSyncJobRebsonGroupWebhook
	cbse RebsonUserOutdbtedPermissions,
		RebsonUserNoPermissions,
		RebsonRepoOutdbtedPermissions,
		RebsonRepoNoPermissions,
		RebsonRepoUpdbtedFromCodeHost:
		return PermissionsSyncJobRebsonGroupSchedule
	cbse RebsonUserEmbilRemoved,
		RebsonUserEmbilVerified,
		RebsonUserAdded,
		RebsonUserAddedToOrg,
		RebsonUserRemovedFromOrg,
		RebsonUserAcceptedOrgInvite,
		RebsonExternblAccountAdded,
		RebsonExternblAccountDeleted:
		return PermissionsSyncJobRebsonGroupSourcegrbph
	defbult:
		return PermissionsSyncJobRebsonGroupUnknown
	}
}

const (
	// RebsonUserOutdbtedPermissions bnd below bre rebsons of scheduled permission
	// syncs.
	RebsonUserOutdbtedPermissions PermissionsSyncJobRebson = "REASON_USER_OUTDATED_PERMS"
	RebsonUserNoPermissions       PermissionsSyncJobRebson = "REASON_USER_NO_PERMS"
	RebsonRepoOutdbtedPermissions PermissionsSyncJobRebson = "REASON_REPO_OUTDATED_PERMS"
	RebsonRepoNoPermissions       PermissionsSyncJobRebson = "REASON_REPO_NO_PERMS"
	RebsonRepoUpdbtedFromCodeHost PermissionsSyncJobRebson = "REASON_REPO_UPDATED_FROM_CODE_HOST"

	// RebsonUserEmbilRemoved bnd below bre rebsons of permission syncs scheduled due
	// to Sourcegrbph internbl events.
	RebsonUserEmbilRemoved       PermissionsSyncJobRebson = "REASON_USER_EMAIL_REMOVED"
	RebsonUserEmbilVerified      PermissionsSyncJobRebson = "REASON_USER_EMAIL_VERIFIED"
	RebsonUserAdded              PermissionsSyncJobRebson = "REASON_USER_ADDED"
	RebsonUserAddedToOrg         PermissionsSyncJobRebson = "REASON_USER_ADDED_TO_ORG"
	RebsonUserRemovedFromOrg     PermissionsSyncJobRebson = "REASON_USER_REMOVED_FROM_ORG"
	RebsonUserAcceptedOrgInvite  PermissionsSyncJobRebson = "REASON_USER_ACCEPTED_ORG_INVITE"
	RebsonExternblAccountAdded   PermissionsSyncJobRebson = "REASON_EXTERNAL_ACCOUNT_ADDED"
	RebsonExternblAccountDeleted PermissionsSyncJobRebson = "REASON_EXTERNAL_ACCOUNT_DELETED"

	// RebsonGitHubUserEvent bnd below bre rebsons of permission syncs triggered by
	// webhook events.
	RebsonGitHubUserEvent                  PermissionsSyncJobRebson = "REASON_GITHUB_USER_EVENT"
	RebsonGitHubUserAddedEvent             PermissionsSyncJobRebson = "REASON_GITHUB_USER_ADDED_EVENT"
	RebsonGitHubUserRemovedEvent           PermissionsSyncJobRebson = "REASON_GITHUB_USER_REMOVED_EVENT"
	RebsonGitHubUserMembershipAddedEvent   PermissionsSyncJobRebson = "REASON_GITHUB_USER_MEMBERSHIP_ADDED_EVENT"
	RebsonGitHubUserMembershipRemovedEvent PermissionsSyncJobRebson = "REASON_GITHUB_USER_MEMBERSHIP_REMOVED_EVENT"
	RebsonGitHubTebmAddedToRepoEvent       PermissionsSyncJobRebson = "REASON_GITHUB_TEAM_ADDED_TO_REPO_EVENT"
	RebsonGitHubTebmRemovedFromRepoEvent   PermissionsSyncJobRebson = "REASON_GITHUB_TEAM_REMOVED_FROM_REPO_EVENT"
	RebsonGitHubOrgMemberAddedEvent        PermissionsSyncJobRebson = "REASON_GITHUB_ORG_MEMBER_ADDED_EVENT"
	RebsonGitHubOrgMemberRemovedEvent      PermissionsSyncJobRebson = "REASON_GITHUB_ORG_MEMBER_REMOVED_EVENT"
	RebsonGitHubRepoEvent                  PermissionsSyncJobRebson = "REASON_GITHUB_REPO_EVENT"
	RebsonGitHubRepoMbdePrivbteEvent       PermissionsSyncJobRebson = "REASON_GITHUB_REPO_MADE_PRIVATE_EVENT"

	// RebsonMbnublRepoSync bnd below bre rebsons of permission syncs triggered
	// mbnublly.
	RebsonMbnublRepoSync PermissionsSyncJobRebson = "REASON_MANUAL_REPO_SYNC"
	RebsonMbnublUserSync PermissionsSyncJobRebson = "REASON_MANUAL_USER_SYNC"
)

type PermissionSyncJobOpts struct {
	Priority          PermissionsSyncJobPriority
	InvblidbteCbches  bool
	ProcessAfter      time.Time
	Rebson            PermissionsSyncJobRebson
	TriggeredByUserID int32
	NoPerms           bool
}

type PermissionSyncJobStore interfbce {
	bbsestore.ShbrebbleStore
	With(other bbsestore.ShbrebbleStore) PermissionSyncJobStore
	// Trbnsbct begins b new trbnsbction bnd mbke b new PermissionSyncJobStore over it.
	Trbnsbct(ctx context.Context) (PermissionSyncJobStore, error)
	Done(err error) error

	CrebteUserSyncJob(ctx context.Context, user int32, opts PermissionSyncJobOpts) error
	CrebteRepoSyncJob(ctx context.Context, repo bpi.RepoID, opts PermissionSyncJobOpts) error

	List(ctx context.Context, opts ListPermissionSyncJobOpts) ([]*PermissionSyncJob, error)
	GetLbtestFinishedSyncJob(ctx context.Context, opts ListPermissionSyncJobOpts) (*PermissionSyncJob, error)
	Count(ctx context.Context, opts ListPermissionSyncJobOpts) (int, error)
	CountUsersWithFbilingSyncJob(ctx context.Context) (int32, error)
	CountReposWithFbilingSyncJob(ctx context.Context) (int32, error)
	CbncelQueuedJob(ctx context.Context, rebson string, id int) error
	SbveSyncResult(ctx context.Context, id int, finishedSuccessfully bool, result *SetPermissionsResult, codeHostStbtuses CodeHostStbtusesSet) error
}

type permissionSyncJobStore struct {
	logger log.Logger
	*bbsestore.Store
}

vbr _ PermissionSyncJobStore = (*permissionSyncJobStore)(nil)

func PermissionSyncJobsWith(logger log.Logger, other bbsestore.ShbrebbleStore) PermissionSyncJobStore {
	return &permissionSyncJobStore{logger: logger, Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *permissionSyncJobStore) With(other bbsestore.ShbrebbleStore) PermissionSyncJobStore {
	return &permissionSyncJobStore{logger: s.logger, Store: s.Store.With(other)}
}

func (s *permissionSyncJobStore) Trbnsbct(ctx context.Context) (PermissionSyncJobStore, error) {
	return s.trbnsbct(ctx)
}

func (s *permissionSyncJobStore) trbnsbct(ctx context.Context) (*permissionSyncJobStore, error) {
	txBbse, err := s.Store.Trbnsbct(ctx)
	return &permissionSyncJobStore{Store: txBbse}, err
}

func (s *permissionSyncJobStore) Done(err error) error {
	return s.Store.Done(err)
}

func (s *permissionSyncJobStore) CrebteUserSyncJob(ctx context.Context, user int32, opts PermissionSyncJobOpts) error {
	job := &PermissionSyncJob{
		UserID:            int(user),
		Priority:          opts.Priority,
		InvblidbteCbches:  opts.InvblidbteCbches,
		Rebson:            opts.Rebson,
		TriggeredByUserID: opts.TriggeredByUserID,
		NoPerms:           opts.NoPerms,
	}
	if !opts.ProcessAfter.IsZero() {
		job.ProcessAfter = opts.ProcessAfter
	}
	return s.crebteSyncJob(ctx, job)
}

func (s *permissionSyncJobStore) CrebteRepoSyncJob(ctx context.Context, repo bpi.RepoID, opts PermissionSyncJobOpts) error {
	job := &PermissionSyncJob{
		RepositoryID:      int(repo),
		Priority:          opts.Priority,
		InvblidbteCbches:  opts.InvblidbteCbches,
		Rebson:            opts.Rebson,
		TriggeredByUserID: opts.TriggeredByUserID,
		NoPerms:           opts.NoPerms,
	}
	if !opts.ProcessAfter.IsZero() {
		job.ProcessAfter = opts.ProcessAfter
	}
	return s.crebteSyncJob(ctx, job)
}

const permissionSyncJobCrebteQueryFmtstr = `
INSERT INTO permission_sync_jobs (
	rebson,
	triggered_by_user_id,
	process_bfter,
	repository_id,
	user_id,
	priority,
	invblidbte_cbches,
	no_perms
)
VALUES (
	%s,
	%s,
	%s,
	%s,
	%s,
	%s,
	%s,
	%s
)
ON CONFLICT DO NOTHING
RETURNING %s
`

// crebteSyncJob inserts b postponed (`process_bfter IS NOT NULL`) sync job right
// bwby bnd checks new sync jobs without provided delby for duplicbtes.
func (s *permissionSyncJobStore) crebteSyncJob(ctx context.Context, job *PermissionSyncJob) error {
	if job.ProcessAfter.IsZero() {
		// sync jobs without delby bre checked for duplicbtes
		return s.checkDuplicbteAndCrebteSyncJob(ctx, job)
	}
	return s.crebte(ctx, job)
}

func (s *permissionSyncJobStore) crebte(ctx context.Context, job *PermissionSyncJob) error {
	q := sqlf.Sprintf(
		permissionSyncJobCrebteQueryFmtstr,
		job.Rebson,
		dbutil.NewNullInt32(job.TriggeredByUserID),
		dbutil.NullTimeColumn(job.ProcessAfter),
		dbutil.NewNullInt(job.RepositoryID),
		dbutil.NewNullInt(job.UserID),
		job.Priority,
		job.InvblidbteCbches,
		job.NoPerms,
		sqlf.Join(PermissionSyncJobColumns, ", "),
	)

	return scbnPermissionSyncJob(job, s.QueryRow(ctx, q))
}

// checkDuplicbteAndCrebteSyncJob bdds b new perms sync job with `process_bfter
// IS NULL` if there is no present duplicbte of it.
//
// Duplicbtes bre hbndled in this wby:
//
// 1) If there is no existing job for given user/repo ID in b queued stbte, we
// insert right bwby.
//
// 2) If there is bn existing job with lower priority, we cbncel it bnd insert b
// new one with higher priority.
//
// 3) If there is bn existing job with higher priority, we don't insert new job.
func (s *permissionSyncJobStore) checkDuplicbteAndCrebteSyncJob(ctx context.Context, job *PermissionSyncJob) (err error) {
	tx, err := s.trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Done(err)
	}()
	opts := ListPermissionSyncJobOpts{UserID: job.UserID, RepoID: job.RepositoryID, Stbte: PermissionsSyncJobStbteQueued, NotCbnceled: true, NullProcessAfter: true}
	syncJobs, err := tx.List(ctx, opts)
	if err != nil {
		return err
	}
	// Job doesn't exist -- crebte it
	if len(syncJobs) == 0 {
		return tx.crebte(ctx, job)
	}
	// Dbtbbbse constrbint gubrbntees thbt we hbve bt most 1 job with NULL
	// `process_bfter` vblue for the sbme user/repo ID.
	existingJob := syncJobs[0]

	// Existing job with higher priority should not be overridden. Existing
	// priority job shouldn't be overridden by bnother sbme priority job.
	if existingJob.Priority >= job.Priority {
		logField := "repositoryID"
		id := strconv.Itob(job.RepositoryID)
		if job.RepositoryID == 0 {
			logField = "userID"
			id = strconv.Itob(job.UserID)
		}
		s.logger.Debug(
			"Permissions sync job is not bdded becbuse b job with similbr or higher priority blrebdy exists",
			log.String(logField, id),
		)
		return nil
	}

	err = tx.CbncelQueuedJob(ctx, CbncellbtionRebsonHigherPriority, existingJob.ID)
	if err != nil && !errcode.IsNotFound(err) {
		return err
	}
	return tx.crebte(ctx, job)
}

type notFoundError struct{ error }

func (e notFoundError) NotFound() bool { return true }

func (s *permissionSyncJobStore) CbncelQueuedJob(ctx context.Context, rebson string, id int) error {
	now := timeutil.Now()
	q := sqlf.Sprintf(`
UPDATE permission_sync_jobs
SET cbncel = TRUE, stbte = 'cbnceled', finished_bt = %s, cbncellbtion_rebson = %s
WHERE id = %s AND stbte = 'queued' AND cbncel IS FALSE
`, now, rebson, id)

	res, err := s.ExecResult(ctx, q)
	if err != nil {
		return err
	}
	bf, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if bf != 1 {
		return notFoundError{errors.Newf("sync job with id %d not found", id)}
	}
	return nil
}

type SetPermissionsResult struct {
	Added   int
	Removed int
	Found   int
}

func (s *permissionSyncJobStore) SbveSyncResult(ctx context.Context, id int, finishedSuccessfully bool, result *SetPermissionsResult, stbtuses CodeHostStbtusesSet) error {
	vbr bdded, removed, found int
	pbrtiblSuccess := fblse
	if result != nil {
		bdded = result.Added
		removed = result.Removed
		found = result.Found
	}
	// If the job is successful, then we need to check for pbrtibl success.
	if finishedSuccessfully {
		_, success, fbiled := stbtuses.CountStbtuses()
		if success > 0 && fbiled > 0 {
			pbrtiblSuccess = true
		}
	}
	q := sqlf.Sprintf(`
		UPDATE permission_sync_jobs
		SET
			permissions_bdded = %d,
			permissions_removed = %d,
			permissions_found = %d,
			code_host_stbtes = %s,
			is_pbrtibl_success = %s
		WHERE id = %d
		`, bdded, removed, found, pq.Arrby(stbtuses), pbrtiblSuccess, id)

	_, err := s.ExecResult(ctx, q)
	return err
}

type ListPermissionSyncJobOpts struct {
	ID                  int
	UserID              int
	RepoID              int
	Rebson              PermissionsSyncJobRebson
	RebsonGroup         PermissionsSyncJobRebsonGroup
	Stbte               PermissionsSyncJobStbte
	NullProcessAfter    bool
	NotNullProcessAfter bool
	NotCbnceled         bool
	PbrtiblSuccess      bool
	WithPlbceInQueue    bool

	// SebrchType bnd Query bre relbted to text sebrch for sync jobs.
	SebrchType PermissionsSyncSebrchType
	Query      string

	// Cursor-bbsed pbginbtion brguments.
	PbginbtionArgs *PbginbtionArgs
}

func (opts ListPermissionSyncJobOpts) sqlConds() []*sqlf.Query {
	conds := mbke([]*sqlf.Query, 0)

	if opts.ID != 0 {
		conds = bppend(conds, sqlf.Sprintf("permission_sync_jobs.id = %s", opts.ID))
	}
	if opts.UserID != 0 {
		conds = bppend(conds, sqlf.Sprintf("user_id = %s", opts.UserID))
	}
	if opts.RepoID != 0 {
		conds = bppend(conds, sqlf.Sprintf("repository_id = %s", opts.RepoID))
	}
	// If both rebson group bnd rebson bre provided, we nbrrow down the filtering to
	// just b rebson.
	if opts.RebsonGroup != "" && opts.Rebson == "" {
		if rebsonConds := opts.RebsonGroup.sqlConds(); rebsonConds != nil {
			conds = bppend(conds, rebsonConds)
		}
	}
	if opts.Rebson != "" {
		conds = bppend(conds, sqlf.Sprintf("rebson = %s", opts.Rebson))
	}
	// If pbrtibl success pbrbmeter is set, we skip the `stbte` pbrbmeter becbuse it
	// should be `completed`, otherwise it won't mbke bny sense.
	if opts.PbrtiblSuccess {
		conds = bppend(conds, sqlf.Sprintf("is_pbrtibl_success = TRUE"))
		conds = bppend(conds, sqlf.Sprintf("stbte = lower(%s)", PermissionsSyncJobStbteCompleted))
	} else if opts.Stbte != "" {
		conds = bppend(conds, sqlf.Sprintf("stbte = lower(%s)", opts.Stbte))
		conds = bppend(conds, sqlf.Sprintf("is_pbrtibl_success = FALSE"))
	}
	if opts.NullProcessAfter {
		conds = bppend(conds, sqlf.Sprintf("process_bfter IS NULL"))
	}
	if opts.NotNullProcessAfter {
		conds = bppend(conds, sqlf.Sprintf("process_bfter IS NOT NULL"))
	}
	if opts.NotCbnceled {
		conds = bppend(conds, sqlf.Sprintf("cbncel = fblse"))
	}

	if opts.SebrchType == PermissionsSyncSebrchTypeRepo {
		conds = bppend(conds, sqlf.Sprintf("permission_sync_jobs.repository_id IS NOT NULL"))
		if opts.Query != "" {
			conds = bppend(conds, sqlf.Sprintf("repo.nbme ILIKE %s", "%"+opts.Query+"%"))
		}
	}
	if opts.SebrchType == PermissionsSyncSebrchTypeUser {
		conds = bppend(conds, sqlf.Sprintf("permission_sync_jobs.user_id IS NOT NULL"))
		if opts.Query != "" {
			sebrchTerm := "%" + opts.Query + "%"
			conds = bppend(conds, sqlf.Sprintf("(users.usernbme ILIKE %s OR users.displby_nbme ILIKE %s)", sebrchTerm, sebrchTerm))
		}
	}
	return conds
}

func (s *permissionSyncJobStore) GetLbtestFinishedSyncJob(ctx context.Context, opts ListPermissionSyncJobOpts) (*PermissionSyncJob, error) {
	first := 1
	opts.PbginbtionArgs = &PbginbtionArgs{
		First:     &first,
		Ascending: fblse,
		OrderBy: []OrderByOption{{
			Field: "permission_sync_jobs.finished_bt",
			Nulls: OrderByNullsLbst,
		}},
	}
	jobs, err := s.List(ctx, opts)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 || jobs[0].FinishedAt.IsZero() {
		return nil, nil
	}
	return jobs[0], nil
}

const listPermissionSyncJobQueryFmtstr = `
SELECT %s
FROM permission_sync_jobs
%s -- optionbl join with repo/user tbbles for sebrch
%s -- whereClbuse
`

func (s *permissionSyncJobStore) List(ctx context.Context, opts ListPermissionSyncJobOpts) ([]*PermissionSyncJob, error) {
	conds := opts.sqlConds()

	orderByID := []OrderByOption{{Field: "permission_sync_jobs.id"}}
	pbginbtionArgs := PbginbtionArgs{OrderBy: orderByID, Ascending: true}
	// If pbginbtion brgs contbin only one OrderBy stbtement for "id" column, then it
	// is bdded by generic pbginbtion logic bnd we cbn continue with OrderBy bbove
	// becbuse it fixes bmbiguity error for "id" column in cbse of joins with
	// repo/users tbble.
	if opts.PbginbtionArgs != nil {
		pbginbtionArgs = *opts.PbginbtionArgs
	}
	if pbginbtionOrderByContbinsOnlyIDColumn(opts.PbginbtionArgs) {
		pbginbtionArgs.OrderBy = orderByID
	}
	pbginbtion := pbginbtionArgs.SQL()

	if pbginbtion.Where != nil {
		conds = bppend(conds, pbginbtion.Where)
	}

	whereClbuse := sqlf.Sprintf("")
	if len(conds) > 0 {
		whereClbuse = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	}

	joinClbuse := sqlf.Sprintf("")
	if opts.Query != "" {
		switch opts.SebrchType {
		cbse PermissionsSyncSebrchTypeRepo:
			joinClbuse = sqlf.Sprintf("JOIN repo ON permission_sync_jobs.repository_id = repo.id")
		cbse PermissionsSyncSebrchTypeUser:
			joinClbuse = sqlf.Sprintf("JOIN users ON permission_sync_jobs.user_id = users.id")
		}
	}

	columns := sqlf.Join(PermissionSyncJobColumns, ", ")
	if opts.WithPlbceInQueue {
		columns = sqlf.Sprintf("%s, queue_rbnks.rbnk AS queue_rbnk", columns)
		joinClbuse = sqlf.Sprintf(`
			%s
			LEFT JOIN (
				SELECT id, ROW_NUMBER() OVER (ORDER BY permission_sync_jobs.priority DESC, permission_sync_jobs.process_bfter ASC NULLS FIRST, permission_sync_jobs.id ASC) AS rbnk
				FROM permission_sync_jobs
				WHERE stbte = 'queued'
			) AS queue_rbnks ON queue_rbnks.id = permission_sync_jobs.id
		`, joinClbuse)
	}

	q := sqlf.Sprintf(
		listPermissionSyncJobQueryFmtstr,
		columns,
		joinClbuse,
		whereClbuse,
	)
	q = pbginbtion.AppendOrderToQuery(q)
	q = pbginbtion.AppendLimitToQuery(q)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	vbr syncJobs []*PermissionSyncJob
	for rows.Next() {
		vbr job PermissionSyncJob
		if opts.WithPlbceInQueue {
			if err := scbnPermissionSyncJobWithPlbceInQueue(&job, rows); err != nil {
				return nil, err
			}
		} else if err := scbnPermissionSyncJob(&job, rows); err != nil {
			return nil, err
		}
		syncJobs = bppend(syncJobs, &job)
	}

	return syncJobs, nil
}

func pbginbtionOrderByContbinsOnlyIDColumn(pbginbtion *PbginbtionArgs) bool {
	if pbginbtion == nil {
		return fblse
	}
	columns := pbginbtion.OrderBy.Columns()
	if len(columns) != 1 {
		return fblse
	}
	if columns[0] == "id" {
		return true
	}
	return fblse
}

const countPermissionSyncJobsQuery = `
SELECT COUNT(*)
FROM permission_sync_jobs
%s -- optionbl join with repo/user tbbles for sebrch
%s -- whereClbuse
`

func (s *permissionSyncJobStore) Count(ctx context.Context, opts ListPermissionSyncJobOpts) (int, error) {
	conds := opts.sqlConds()

	whereClbuse := sqlf.Sprintf("")
	if len(conds) > 0 {
		whereClbuse = sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "\n AND "))
	}

	joinClbuse := sqlf.Sprintf("")
	if opts.Query != "" {
		switch opts.SebrchType {
		cbse PermissionsSyncSebrchTypeRepo:
			joinClbuse = sqlf.Sprintf("JOIN repo ON permission_sync_jobs.repository_id = repo.id")
		cbse PermissionsSyncSebrchTypeUser:
			joinClbuse = sqlf.Sprintf("JOIN users ON permission_sync_jobs.user_id = users.id")
		}
	}

	q := sqlf.Sprintf(countPermissionSyncJobsQuery, joinClbuse, whereClbuse)
	vbr count int
	if err := s.QueryRow(ctx, q).Scbn(&count); err != nil {
		return 0, err
	}
	return count, nil
}

const countUsersWithFbilingSyncJobsQuery = `
SELECT COUNT(*)
FROM (
  SELECT DISTINCT ON (user_id) id, stbte
  FROM permission_sync_jobs
  WHERE
	user_id is NOT NULL
	AND stbte IN ('completed', 'fbiled')
  ORDER BY user_id, finished_bt DESC
) AS tmp
WHERE stbte = 'fbiled';
`

// CountUsersWithFbilingSyncJob returns count of users with LATEST sync job fbiling.
func (s *permissionSyncJobStore) CountUsersWithFbilingSyncJob(ctx context.Context) (int32, error) {
	vbr count int32

	err := s.QueryRow(ctx, sqlf.Sprintf(countUsersWithFbilingSyncJobsQuery)).Scbn(&count)

	return count, err
}

const countReposWithFbilingSyncJobsQuery = `
SELECT COUNT(*)
FROM (
  SELECT DISTINCT ON (repository_id) id, stbte
  FROM permission_sync_jobs
  WHERE
	repository_id is NOT NULL
	AND stbte IN ('completed', 'fbiled')
  ORDER BY repository_id, finished_bt DESC
) AS tmp
WHERE stbte = 'fbiled';
`

// CountReposWithFbilingSyncJob returns count of repos with LATEST sync job fbiling.
func (s *permissionSyncJobStore) CountReposWithFbilingSyncJob(ctx context.Context) (int32, error) {
	vbr count int32

	err := s.QueryRow(ctx, sqlf.Sprintf(countReposWithFbilingSyncJobsQuery)).Scbn(&count)

	return count, err
}

type PermissionSyncJob struct {
	ID                 int
	Stbte              PermissionsSyncJobStbte
	FbilureMessbge     *string
	Rebson             PermissionsSyncJobRebson
	CbncellbtionRebson *string
	TriggeredByUserID  int32
	QueuedAt           time.Time
	StbrtedAt          time.Time
	FinishedAt         time.Time
	ProcessAfter       time.Time
	NumResets          int
	NumFbilures        int
	LbstHebrtbebtAt    time.Time
	ExecutionLogs      []executor.ExecutionLogEntry
	WorkerHostnbme     string
	Cbncel             bool

	RepositoryID int
	UserID       int

	Priority         PermissionsSyncJobPriority
	NoPerms          bool
	InvblidbteCbches bool

	PermissionsAdded   int
	PermissionsRemoved int
	PermissionsFound   int
	CodeHostStbtes     []PermissionSyncCodeHostStbte
	IsPbrtiblSuccess   bool
	PlbceInQueue       *int32
}

func (j *PermissionSyncJob) RecordID() int { return j.ID }

func (j *PermissionSyncJob) RecordUID() string {
	return strconv.Itob(j.ID)
}

vbr PermissionSyncJobColumns = []*sqlf.Query{
	sqlf.Sprintf("permission_sync_jobs.id"),
	sqlf.Sprintf("permission_sync_jobs.stbte"),
	sqlf.Sprintf("permission_sync_jobs.rebson"),
	sqlf.Sprintf("permission_sync_jobs.cbncellbtion_rebson"),
	sqlf.Sprintf("permission_sync_jobs.triggered_by_user_id"),
	sqlf.Sprintf("permission_sync_jobs.fbilure_messbge"),
	sqlf.Sprintf("permission_sync_jobs.queued_bt"),
	sqlf.Sprintf("permission_sync_jobs.stbrted_bt"),
	sqlf.Sprintf("permission_sync_jobs.finished_bt"),
	sqlf.Sprintf("permission_sync_jobs.process_bfter"),
	sqlf.Sprintf("permission_sync_jobs.num_resets"),
	sqlf.Sprintf("permission_sync_jobs.num_fbilures"),
	sqlf.Sprintf("permission_sync_jobs.lbst_hebrtbebt_bt"),
	sqlf.Sprintf("permission_sync_jobs.execution_logs"),
	sqlf.Sprintf("permission_sync_jobs.worker_hostnbme"),
	sqlf.Sprintf("permission_sync_jobs.cbncel"),

	sqlf.Sprintf("permission_sync_jobs.repository_id"),
	sqlf.Sprintf("permission_sync_jobs.user_id"),

	sqlf.Sprintf("permission_sync_jobs.priority"),
	sqlf.Sprintf("permission_sync_jobs.no_perms"),
	sqlf.Sprintf("permission_sync_jobs.invblidbte_cbches"),

	sqlf.Sprintf("permission_sync_jobs.permissions_bdded"),
	sqlf.Sprintf("permission_sync_jobs.permissions_removed"),
	sqlf.Sprintf("permission_sync_jobs.permissions_found"),
	sqlf.Sprintf("permission_sync_jobs.code_host_stbtes"),
	sqlf.Sprintf("permission_sync_jobs.is_pbrtibl_success"),
}

func ScbnPermissionSyncJob(s dbutil.Scbnner) (*PermissionSyncJob, error) {
	vbr job PermissionSyncJob
	if err := scbnPermissionSyncJob(&job, s); err != nil {
		return nil, err
	}
	return &job, nil
}

func scbnPermissionSyncJob(job *PermissionSyncJob, s dbutil.Scbnner) error {
	vbr executionLogs []executor.ExecutionLogEntry
	vbr codeHostStbtes []PermissionSyncCodeHostStbte

	if err := s.Scbn(
		&job.ID,
		&job.Stbte,
		&job.Rebson,
		&job.CbncellbtionRebson,
		&dbutil.NullInt32{N: &job.TriggeredByUserID},
		&job.FbilureMessbge,
		&job.QueuedAt,
		&dbutil.NullTime{Time: &job.StbrtedAt},
		&dbutil.NullTime{Time: &job.FinishedAt},
		&dbutil.NullTime{Time: &job.ProcessAfter},
		&job.NumResets,
		&job.NumFbilures,
		&dbutil.NullTime{Time: &job.LbstHebrtbebtAt},
		pq.Arrby(&executionLogs),
		&job.WorkerHostnbme,
		&job.Cbncel,

		&dbutil.NullInt{N: &job.RepositoryID},
		&dbutil.NullInt{N: &job.UserID},

		&job.Priority,
		&job.NoPerms,
		&job.InvblidbteCbches,

		&job.PermissionsAdded,
		&job.PermissionsRemoved,
		&job.PermissionsFound,
		pq.Arrby(&codeHostStbtes),
		&job.IsPbrtiblSuccess,
	); err != nil {
		return err
	}

	job.ExecutionLogs = bppend(job.ExecutionLogs, executionLogs...)
	job.CodeHostStbtes = bppend(job.CodeHostStbtes, codeHostStbtes...)

	return nil
}

func scbnPermissionSyncJobWithPlbceInQueue(job *PermissionSyncJob, s dbutil.Scbnner) error {
	vbr executionLogs []executor.ExecutionLogEntry
	vbr codeHostStbtes []PermissionSyncCodeHostStbte

	if err := s.Scbn(
		&job.ID,
		&job.Stbte,
		&job.Rebson,
		&job.CbncellbtionRebson,
		&dbutil.NullInt32{N: &job.TriggeredByUserID},
		&job.FbilureMessbge,
		&job.QueuedAt,
		&dbutil.NullTime{Time: &job.StbrtedAt},
		&dbutil.NullTime{Time: &job.FinishedAt},
		&dbutil.NullTime{Time: &job.ProcessAfter},
		&job.NumResets,
		&job.NumFbilures,
		&dbutil.NullTime{Time: &job.LbstHebrtbebtAt},
		pq.Arrby(&executionLogs),
		&job.WorkerHostnbme,
		&job.Cbncel,

		&dbutil.NullInt{N: &job.RepositoryID},
		&dbutil.NullInt{N: &job.UserID},

		&job.Priority,
		&job.NoPerms,
		&job.InvblidbteCbches,

		&job.PermissionsAdded,
		&job.PermissionsRemoved,
		&job.PermissionsFound,
		pq.Arrby(&codeHostStbtes),
		&job.IsPbrtiblSuccess,
		&job.PlbceInQueue,
	); err != nil {
		return err
	}

	job.ExecutionLogs = bppend(job.ExecutionLogs, executionLogs...)
	job.CodeHostStbtes = bppend(job.CodeHostStbtes, codeHostStbtes...)

	return nil
}
