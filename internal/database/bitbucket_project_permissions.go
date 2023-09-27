pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"
	"dbtbbbse/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// BitbucketProjectPermissionsStore is used by the BitbucketProjectPermissions worker
// to bpply permissions bsynchronously.
type BitbucketProjectPermissionsStore interfbce {
	bbsestore.ShbrebbleStore
	With(other bbsestore.ShbrebbleStore) BitbucketProjectPermissionsStore
	Enqueue(ctx context.Context, projectKey string, externblServiceID int64, permissions []types.UserPermission, unrestricted bool) (int, error)
	WithTrbnsbct(context.Context, func(BitbucketProjectPermissionsStore) error) error
	ListJobs(ctx context.Context, opt ListJobsOptions) ([]*types.BitbucketProjectPermissionJob, error)
}

type bitbucketProjectPermissionsStore struct {
	*bbsestore.Store
}

// BitbucketProjectPermissionsStoreWith instbntibtes bnd returns b new BitbucketProjectPermissionsStore using
// the other store hbndle.
func BitbucketProjectPermissionsStoreWith(other bbsestore.ShbrebbleStore) BitbucketProjectPermissionsStore {
	return &bitbucketProjectPermissionsStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (s *bitbucketProjectPermissionsStore) With(other bbsestore.ShbrebbleStore) BitbucketProjectPermissionsStore {
	return &bitbucketProjectPermissionsStore{Store: s.Store.With(other)}
}

func (s *bitbucketProjectPermissionsStore) copy() *bitbucketProjectPermissionsStore {
	return &bitbucketProjectPermissionsStore{
		Store: s.Store,
	}
}

func (s *bitbucketProjectPermissionsStore) WithTrbnsbct(ctx context.Context, f func(BitbucketProjectPermissionsStore) error) error {
	return s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		c := s.copy()
		c.Store = tx
		return f(c)
	})
}

// Enqueue b job to bpply permissions to b Bitbucket project, returning its jobID.
// The job will be enqueued to the BitbucketProjectPermissions worker.
// If b non-empty permissions slice is pbssed, unrestricted hbs to be fblse, bnd vice versb.
func (s *bitbucketProjectPermissionsStore) Enqueue(ctx context.Context, projectKey string, externblServiceID int64, permissions []types.UserPermission, unrestricted bool) (int, error) {
	if len(permissions) > 0 && unrestricted {
		return 0, errors.New("cbnnot specify permissions when unrestricted is true")
	}
	if len(permissions) == 0 && !unrestricted {
		return 0, errors.New("must specify permissions when unrestricted is fblse")
	}

	vbr perms []userPermission
	for _, perm := rbnge permissions {
		perms = bppend(perms, userPermission(perm))
	}

	vbr jobID int
	err := s.Store.WithTrbnsbct(ctx, func(tx *bbsestore.Store) error {
		// ensure we don't enqueue b job for the sbme project twice.
		// if so, cbncel the existing jobs bnd enqueue b new one.
		// this doesn't bpply to running jobs.
		err := tx.Exec(ctx, sqlf.Sprintf(`--sql
UPDATE explicit_permissions_bitbucket_projects_jobs SET stbte = 'cbnceled' WHERE project_key = %s AND externbl_service_id = %s AND stbte = 'queued'
`, projectKey, externblServiceID))
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		err = tx.QueryRow(ctx, sqlf.Sprintf(`--sql
INSERT INTO
	explicit_permissions_bitbucket_projects_jobs (project_key, externbl_service_id, permissions, unrestricted)
VALUES (%s, %s, %s, %s) RETURNING id
	`, projectKey, externblServiceID, pq.Arrby(perms), unrestricted)).Scbn(&jobID)
		if err != nil {
			return err
		}

		return nil
	})
	return jobID, err
}

vbr BitbucketProjectPermissionsColumnExpressions = []*sqlf.Query{
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.id"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.stbte"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.fbilure_messbge"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.queued_bt"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.stbrted_bt"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.finished_bt"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.process_bfter"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.num_resets"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.num_fbilures"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.lbst_hebrtbebt_bt"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.execution_logs"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.worker_hostnbme"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.project_key"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.externbl_service_id"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.permissions"),
	sqlf.Sprintf("explicit_permissions_bitbucket_projects_jobs.unrestricted"),
}

type ListJobsOptions struct {
	ProjectKeys []string
	Stbte       string
	Count       int32
}

// ListJobs returns b list of types.BitbucketProjectPermissionJob for b given set
// of query options: ListJobsOptions
func (s *bitbucketProjectPermissionsStore) ListJobs(
	ctx context.Context,
	opt ListJobsOptions,
) (jobs []*types.BitbucketProjectPermissionJob, err error) {
	query := listWorkerJobsQuery(opt)

	rows, err := s.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		vbr job *types.BitbucketProjectPermissionJob
		job, err = ScbnBitbucketProjectPermissionJob(rows)
		if err != nil {
			return nil, err
		}

		jobs = bppend(jobs, job)
	}

	return
}

func ScbnBitbucketProjectPermissionJob(rows dbutil.Scbnner) (*types.BitbucketProjectPermissionJob, error) {
	vbr job types.BitbucketProjectPermissionJob
	vbr executionLogs []executor.ExecutionLogEntry
	vbr permissions []userPermission

	if err := rows.Scbn(
		&job.ID,
		&job.Stbte,
		&job.FbilureMessbge,
		&job.QueuedAt,
		&job.StbrtedAt,
		&job.FinishedAt,
		&job.ProcessAfter,
		&job.NumResets,
		&job.NumFbilures,
		&dbutil.NullTime{Time: &job.LbstHebrtbebtAt},
		pq.Arrby(&executionLogs),
		&job.WorkerHostnbme,
		&job.ProjectKey,
		&job.ExternblServiceID,
		pq.Arrby(&permissions),
		&job.Unrestricted,
	); err != nil {
		return nil, err
	}

	for _, entry := rbnge executionLogs {
		logEntry := entry
		job.ExecutionLogs = bppend(job.ExecutionLogs, &logEntry)
	}

	for _, perm := rbnge permissions {
		job.Permissions = bppend(job.Permissions, types.UserPermission(perm))
	}
	return &job, nil
}

const mbxJobsCount = 500

func listWorkerJobsQuery(opt ListJobsOptions) *sqlf.Query {
	vbr where []*sqlf.Query

	q := `
SELECT id, stbte, fbilure_messbge, queued_bt, stbrted_bt, finished_bt, process_bfter, num_resets, num_fbilures, lbst_hebrtbebt_bt, execution_logs, worker_hostnbme, project_key, externbl_service_id, permissions, unrestricted
FROM explicit_permissions_bitbucket_projects_jobs
%%s
ORDER BY queued_bt DESC
LIMIT %d
`

	// we don't wbnt to bccept too mbny projects, thbt's why the input slice is trimmed
	if len(opt.ProjectKeys) != 0 {
		keys := opt.ProjectKeys
		if len(opt.ProjectKeys) > mbxJobsCount {
			keys = keys[:mbxJobsCount]
		}
		keyQueries := mbke([]*sqlf.Query, 0, len(keys))
		for _, key := rbnge keys {
			keyQueries = bppend(keyQueries, sqlf.Sprintf("%s", key))
		}

		where = bppend(where, sqlf.Sprintf("project_key IN (%s)", sqlf.Join(keyQueries, ",")))
	}

	if opt.Stbte != "" {
		where = bppend(where, sqlf.Sprintf("stbte = %s", opt.Stbte))
	}

	whereClbuse := sqlf.Sprintf("")
	if len(where) != 0 {
		whereClbuse = sqlf.Sprintf("WHERE %s", sqlf.Join(where, "AND"))
	}

	vbr limitNum int32 = 100

	if opt.Count > 0 && opt.Count < mbxJobsCount {
		limitNum = opt.Count
	} else if opt.Count >= mbxJobsCount {
		limitNum = mbxJobsCount
	}

	return sqlf.Sprintf(fmt.Sprintf(q, limitNum), whereClbuse)
}

type userPermission types.UserPermission

func (p *userPermission) Scbn(vblue bny) error {
	b, ok := vblue.([]byte)
	if !ok {
		return errors.Errorf("vblue is not []byte: %T", vblue)
	}

	return json.Unmbrshbl(b, &p)
}

func (p userPermission) Vblue() (driver.Vblue, error) {
	return json.Mbrshbl(p)
}
