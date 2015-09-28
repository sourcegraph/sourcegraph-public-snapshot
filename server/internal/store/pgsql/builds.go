package pgsql

import (
	"fmt"
	"strings"
	"time"

	"github.com/sqs/modl"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	t := Schema.Map.AddTableWithName(dbBuild{}, "repo_build").SetKeys(false, "attempt", "commit_id", "repo")
	t.ColMap("commit_id").SetMaxSize(40)
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE repo_build ALTER COLUMN started_at TYPE timestamp with time zone USING started_at::timestamp with time zone;`,
		`ALTER TABLE repo_build ALTER COLUMN ended_at TYPE timestamp with time zone USING ended_at::timestamp with time zone;`,
		`ALTER TABLE repo_build ALTER COLUMN heartbeat_at TYPE timestamp with time zone USING ended_at::timestamp with time zone;`,
		`CREATE INDEX repo_build_repo ON repo_build(repo);`,
		`CREATE INDEX repo_build_priority ON repo_build(priority);`,
		`create index repo_build_created_at on repo_build(created_at desc nulls last);`,
		`create index repo_build_updated_at on repo_build((greatest(started_at, ended_at, created_at)) desc nulls last);`,
		`create index repo_build_successful on repo_build(repo) where success and not purged;`,

		// Set attempt to 1 + the max previous attempt for the repo and commit ID.
		`CREATE OR REPLACE FUNCTION increment_attempt() RETURNS trigger IMMUTABLE AS $$
         BEGIN
           RAISE WARNING 'Before %', NEW;
           IF NEW.attempt = 0 OR NEW.attempt IS NULL THEN
             NEW.attempt = (SELECT COALESCE(max(b.attempt), 0) + 1 FROM repo_build b WHERE b.repo=NEW.repo AND b.commit_id=NEW.commit_id);
           END IF;
           RAISE WARNING 'After %', NEW;
           RETURN NEW;
         END
         $$ language plpgsql;`,
		`CREATE TRIGGER repo_build_next_attempt BEFORE INSERT ON repo_build FOR EACH ROW EXECUTE PROCEDURE increment_attempt();`,
	)

	Schema.Map.AddTableWithName(dbBuildTask{}, "repo_build_task").SetKeys(true, "taskid")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE repo_build_task ALTER COLUMN started_at TYPE timestamp with time zone USING started_at::timestamp with time zone;`,
		`ALTER TABLE repo_build_task ALTER COLUMN ended_at TYPE timestamp with time zone USING ended_at::timestamp with time zone;`,
		`CREATE INDEX repo_build_task_build ON repo_build_task(repo, attempt, commit_id);`,
	)
}

// dbBuild DB-maps a sourcegraph.Build object.
type dbBuild struct {
	Attempt     uint32
	Repo        string
	CommitID    string     `db:"commit_id"`
	CreatedAt   time.Time  `db:"created_at"`
	StartedAt   *time.Time `db:"started_at"`
	EndedAt     *time.Time `db:"ended_at"`
	HeartbeatAt *time.Time `db:"heartbeat_at"`
	Success     bool
	Failure     bool
	Killed      bool
	Host        string
	Purged      bool
	Import      bool
	Queue       bool
	UseCache    bool
	Priority    int
}

func (b *dbBuild) toBuild() *sourcegraph.Build {
	return &sourcegraph.Build{
		Attempt:     b.Attempt,
		Repo:        b.Repo,
		CommitID:    b.CommitID,
		CreatedAt:   pbtypes.NewTimestamp(b.CreatedAt),
		StartedAt:   ts(b.StartedAt),
		EndedAt:     ts(b.EndedAt),
		HeartbeatAt: ts(b.HeartbeatAt),
		Success:     b.Success,
		Failure:     b.Failure,
		Killed:      b.Killed,
		Host:        b.Host,
		Purged:      b.Purged,
		BuildConfig: sourcegraph.BuildConfig{
			Import:   b.Import,
			Queue:    b.Queue,
			UseCache: b.UseCache,
			Priority: int32(b.Priority),
		},
	}
}

func (b *dbBuild) fromBuild(b2 *sourcegraph.Build) {
	b.Attempt = b2.Attempt
	b.Repo = b2.Repo
	b.CommitID = b2.CommitID
	b.CreatedAt = b2.CreatedAt.Time()
	b.StartedAt = tm(b2.StartedAt)
	b.EndedAt = tm(b2.EndedAt)
	b.HeartbeatAt = tm(b2.HeartbeatAt)
	b.Success = b2.Success
	b.Failure = b2.Failure
	b.Killed = b2.Killed
	b.Host = b2.Host
	b.Purged = b2.Purged
	b.Import = b2.Import
	b.Queue = b2.Queue
	b.UseCache = b2.UseCache
	b.Priority = int(b2.Priority)
}

func toBuilds(bs []*dbBuild) []*sourcegraph.Build {
	b2s := make([]*sourcegraph.Build, len(bs))
	for i, b := range bs {
		b2s[i] = b.toBuild()
	}
	return b2s
}

// dbBuildTask DB-maps a sourcegraph.BuildTask object.
type dbBuildTask struct {
	TaskID    int64 `db:"taskid"`
	Repo      string
	Attempt   uint32
	CommitID  string `db:"commit_id"`
	UnitType  string `db:"unit_type"`
	Unit      string
	Op        string
	Order     int
	CreatedAt time.Time  `db:"created_at"`
	StartedAt *time.Time `db:"started_at"`
	EndedAt   *time.Time `db:"ended_at"`
	Queue     bool
	Success   bool
	Failure   bool
}

func (t *dbBuildTask) toBuildTask() *sourcegraph.BuildTask {
	return &sourcegraph.BuildTask{
		TaskID:    t.TaskID,
		Repo:      t.Repo,
		Attempt:   t.Attempt,
		CommitID:  t.CommitID,
		UnitType:  t.UnitType,
		Unit:      t.Unit,
		Op:        t.Op,
		Order:     int32(t.Order),
		CreatedAt: pbtypes.NewTimestamp(t.CreatedAt),
		StartedAt: ts(t.StartedAt),
		EndedAt:   ts(t.EndedAt),
		Success:   t.Success,
		Failure:   t.Failure,
	}
}

func (t *dbBuildTask) fromBuildTask(t2 *sourcegraph.BuildTask) {
	t.TaskID = t2.TaskID
	t.Repo = t2.Repo
	t.Attempt = t2.Attempt
	t.CommitID = t2.CommitID
	t.UnitType = t2.UnitType
	t.Unit = t2.Unit
	t.Op = t2.Op
	t.Order = int(t2.Order)
	t.CreatedAt = t2.CreatedAt.Time()
	t.StartedAt = tm(t2.StartedAt)
	t.EndedAt = tm(t2.EndedAt)
	t.Success = t2.Success
	t.Failure = t2.Failure
}

func toBuildTasks(ts []*dbBuildTask) []*sourcegraph.BuildTask {
	t2s := make([]*sourcegraph.BuildTask, len(ts))
	for i, t := range ts {
		t2s[i] = t.toBuildTask()
	}
	return t2s
}

// Builds is a DB-backed implementation of the Builds store.
type Builds struct{}

var _ store.Builds = (*Builds)(nil)

func (s *Builds) Get(ctx context.Context, buildSpec sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
	var builds []*dbBuild
	err := dbh(ctx).Select(&builds, `SELECT * FROM repo_build WHERE attempt=$1 AND commit_id=$2 AND repo=$3 LIMIT 1;`, buildSpec.Attempt, buildSpec.CommitID, buildSpec.Repo.URI)
	if err != nil {
		return nil, err
	}
	if len(builds) != 1 {
		return nil, sourcegraph.ErrBuildNotFound
	}
	return builds[0].toBuild(), nil
}

func (s *Builds) List(ctx context.Context, opt *sourcegraph.BuildListOptions) ([]*sourcegraph.Build, error) {
	if opt == nil {
		opt = &sourcegraph.BuildListOptions{}
	}

	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return modl.PostgresDialect{}.BindVar(len(args) - 1)
	}

	var conds []string
	if opt.Repo != "" {
		conds = append(conds, "b.repo="+arg(opt.Repo))
	}
	if opt.Queued {
		conds = append(conds, "b.started_at IS NULL AND b.queue")
	} else {
		if opt.Active {
			conds = append(conds, "b.started_at IS NOT NULL AND b.ended_at IS NULL AND (NOT b.failure)")
		}
		if opt.Failed {
			conds = append(conds, "b.failure")
		}
		if opt.Ended {
			conds = append(conds, "b.ended_at IS NOT NULL")
		}
		if opt.Succeeded {
			conds = append(conds, "b.success")
		}
	}
	if opt.Purged {
		conds = append(conds, "b.purged")
	} else {
		conds = append(conds, "NOT b.purged")
	}
	if opt.CommitID != "" {
		if len(opt.CommitID) == 40 {
			conds = append(conds, "b.commit_id="+arg(opt.CommitID))
		} else {
			conds = append(conds, "b.commit_id LIKE "+arg(opt.CommitID+"%"))
		}
	}
	condsSQL := strings.Join(conds, " AND ")
	if condsSQL != "" {
		condsSQL = "WHERE " + condsSQL
	}

	// Sort and paginate
	sort := opt.Sort
	if sort == "" {
		sort = "build"
	}
	direction := opt.Direction
	if direction == "" {
		direction = "asc"
	}
	sortKeyToCol := map[string]string{
		"build":      "b.repo %(direction)s, b.commit_id %(direction)s, b.attempt %(direction)s",
		"created_at": "b.created_at %(direction)s NULLS LAST",
		"started_at": "b.started_at %(direction)s NULLS LAST",
		"ended_at":   "b.ended_at %(direction)s NULLS LAST",
		"updated_at": "greatest(b.started_at, b.ended_at, b.created_at) %(direction)s NULLS LAST",
		"priority":   "b.priority %(direction)s, greatest(b.started_at, b.ended_at, b.created_at) ASC NULLS LAST",
	}
	if sortCol, valid := sortKeyToCol[sort]; valid {
		sort = sortCol
	} else {
		return nil, &sourcegraph.InvalidOptionsError{Reason: "invalid sort: " + sort}
	}
	if direction != "asc" && direction != "desc" {
		return nil, &sourcegraph.InvalidOptionsError{Reason: "invalid direction: " + direction}
	}

	orderSQL := fmt.Sprintf(" ORDER BY %s", strings.Replace(sort, "%(direction)s", direction, -1))
	limitSQL := `LIMIT ` + arg(opt.Limit()) + ` OFFSET ` + arg(opt.Offset())

	sql := `
WITH builds AS (
  SELECT * FROM repo_build b
  ` + condsSQL + `
  ` + orderSQL + ` ` + limitSQL + `
)
SELECT b.* FROM builds b
` + orderSQL

	var builds []*dbBuild
	if err := dbh(ctx).Select(&builds, sql, args...); err != nil {
		return nil, err
	}

	return toBuilds(builds), nil
}

func (s *Builds) GetFirstInCommitOrder(ctx context.Context, repo string, commitIDs []string, successfulOnly bool) (build *sourcegraph.Build, nth int, err error) {
	if len(commitIDs) == 0 {
		return nil, -1, nil
	}

	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return modl.PostgresDialect{}.BindVar(len(args) - 1)
	}

	sortCases := make([]string, len(commitIDs))
	for i, commitID := range commitIDs {
		sortCases[i] = fmt.Sprintf("WHEN commit_id=%s THEN %d", arg(commitID), i)
	}
	sortFn := "CASE " + strings.Join(sortCases, " ") + " END ASC NULLS LAST, started_at DESC NULLS LAST"

	var successCond string
	if successfulOnly {
		successCond = " AND success "
	}

	sql := `-- Builds.GetFirstInCommitOrder
SELECT * FROM repo_build
WHERE repo=` + arg(repo) + ` AND (NOT purged) ` + successCond + `
      AND commit_id=ANY(` + arg(&dbutil.StringSlice{Slice: commitIDs}) + `)
ORDER BY ` + sortFn + `
LIMIT 1
`

	var builds []*dbBuild
	if err := dbh(ctx).Select(&builds, sql, args...); err != nil {
		return nil, 0, err
	}
	if len(builds) == 1 {
		// Found it!
		build := builds[0].toBuild()

		// Determine which commit ID position this belongs to.
		nth := -1
		for i, c := range commitIDs {
			if build.CommitID == c {
				nth = i
			}
		}
		if nth == -1 {
			panic("build commit ID " + build.CommitID + " was not in arg list")
		}

		return build, nth, nil
	}
	return nil, -1, nil
}

func (s *Builds) Create(ctx context.Context, newBuild *sourcegraph.Build) (*sourcegraph.Build, error) {
	var b dbBuild
	b.fromBuild(newBuild)

	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return modl.PostgresDialect{}.BindVar(len(args) - 1)
	}

	// Construct SQL manually so we can retrieve the attempt # from
	// the DB trigger.
	sql := `INSERT INTO repo_build(attempt, repo, commit_id, created_at, started_at, ended_at, heartbeat_at,
                                   success, failure, killed, host, purged, import, queue, usecache, priority)
            VALUES(` + arg(b.Attempt) + `, ` + arg(b.Repo) + `, ` + arg(b.CommitID) + `, ` + arg(b.CreatedAt) + `, ` + arg(b.StartedAt) + `,` +
		arg(b.EndedAt) + `,` + arg(b.HeartbeatAt) + `, ` + arg(b.Success) + `, ` + arg(b.Failure) + `, ` + arg(b.Killed) + `, ` +
		arg(b.Host) + `, ` + arg(b.Purged) + `, ` + arg(b.Import) + `, ` + arg(b.Queue) + `, ` + arg(b.UseCache) + `,` + arg(b.Priority) + `)
            RETURNING attempt;`
	attempt, err := dbutil.SelectInt(dbh(ctx), sql, args...)
	if err != nil {
		return nil, err
	}
	b.Attempt = uint32(attempt)
	return b.toBuild(), nil
}

func (s *Builds) Update(ctx context.Context, build sourcegraph.BuildSpec, info sourcegraph.BuildUpdate) error {
	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return modl.PostgresDialect{}.BindVar(len(args) - 1)
	}

	var updates []string
	if info.StartedAt != nil {
		updates = append(updates, "started_at="+arg(info.StartedAt.Time()))
	}
	if info.EndedAt != nil {
		updates = append(updates, "ended_at="+arg(info.EndedAt.Time()))
	}
	if info.HeartbeatAt != nil {
		updates = append(updates, "heartbeat_at="+arg(info.HeartbeatAt.Time()))
	}
	updates = append(updates, "host="+arg(info.Host))
	updates = append(updates, "purged="+arg(info.Purged))
	updates = append(updates, "success="+arg(info.Success))
	updates = append(updates, "failure="+arg(info.Failure))
	updates = append(updates, "priority="+arg(info.Priority))
	updates = append(updates, "killed="+arg(info.Killed))

	if len(updates) != 0 {
		sql := fmt.Sprintf(`UPDATE repo_build SET %s WHERE attempt=%s AND commit_id=%s AND repo=%s`, strings.Join(updates, ", "), arg(build.Attempt), arg(build.CommitID), arg(build.Repo.URI))

		if _, err := dbh(ctx).Exec(sql, args...); err != nil {
			return err
		}
	}

	return nil
}

func (s *Builds) CreateTasks(ctx context.Context, tasks []*sourcegraph.BuildTask) ([]*sourcegraph.BuildTask, error) {
	created := make([]*dbBuildTask, len(tasks))
	for i, task := range tasks {
		created[i] = &dbBuildTask{}
		created[i].fromBuildTask(task)
		if err := dbh(ctx).Insert(created[i]); err != nil {
			return nil, err
		}
	}
	return toBuildTasks(created), nil
}

func (s *Builds) UpdateTask(ctx context.Context, task sourcegraph.TaskSpec, info sourcegraph.TaskUpdate) error {
	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return modl.PostgresDialect{}.BindVar(len(args) - 1)
	}

	var updates []string
	if info.StartedAt != nil {
		updates = append(updates, "started_at="+arg(info.StartedAt.Time()))
	}
	if info.EndedAt != nil {
		updates = append(updates, "ended_at="+arg(info.EndedAt.Time()))
	}
	if info.Success {
		updates = append(updates, "success="+arg(info.Success))
	}
	if info.Failure {
		updates = append(updates, "failure="+arg(info.Failure))
	}

	if len(updates) != 0 {
		sql := `UPDATE repo_build_task SET ` + strings.Join(updates, ", ") + ` WHERE taskid=` + arg(task.TaskID)
		if _, err := dbh(ctx).Exec(sql, args...); err != nil {
			return err
		}
	}

	return nil
}

func (s *Builds) ListBuildTasks(ctx context.Context, build sourcegraph.BuildSpec, opt *sourcegraph.BuildTaskListOptions) ([]*sourcegraph.BuildTask, error) {
	if opt == nil {
		opt = &sourcegraph.BuildTaskListOptions{}
	}

	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return modl.PostgresDialect{}.BindVar(len(args) - 1)
	}

	conds := []string{"attempt=" + arg(build.Attempt), "commit_id=" + arg(build.CommitID), "repo=" + arg(build.Repo.URI)}
	condsSQL := strings.Join(conds, " AND ")

	sql := `-- Builds.ListBuildTasks
SELECT * FROM repo_build_task
WHERE ` + condsSQL + `
ORDER BY taskid ASC
LIMIT ` + arg(opt.Limit()) + ` OFFSET ` + arg(opt.Offset()) + `;`
	var tasks []*dbBuildTask
	if err := dbh(ctx).Select(&tasks, sql, args...); err != nil {
		return nil, err
	}
	return toBuildTasks(tasks), nil
}

func (s *Builds) DequeueNext(ctx context.Context) (*sourcegraph.Build, error) {
	sql := `-- Builds.DequeueNext
WITH
to_dequeue AS (
  SELECT * FROM repo_build
  WHERE started_at IS NULL AND queue AND (NOT purged)
  ORDER BY priority desc, greatest(started_at, ended_at, created_at) ASC NULLS LAST
  LIMIT 1
  FOR UPDATE
)
UPDATE repo_build
SET started_at = clock_timestamp()
FROM to_dequeue
WHERE repo_build.repo = to_dequeue.repo AND repo_build.commit_id = to_dequeue.commit_id AND repo_build.attempt = to_dequeue.attempt
RETURNING repo_build.*;
`
	var nextBuilds []*dbBuild
	if err := dbh(ctx).Select(&nextBuilds, sql); err != nil {
		return nil, err
	}
	if len(nextBuilds) == 0 {
		return nil, nil
	}
	return nextBuilds[0].toBuild(), nil
}

func (s *Builds) GetTask(ctx context.Context, task sourcegraph.TaskSpec) (*sourcegraph.BuildTask, error) {
	var tasks []*dbBuildTask
	sql := `SELECT * FROM repo_build_task WHERE repo=$1 AND attempt=$2 AND commit_id=$3 AND taskid=$4;`
	if err := dbh(ctx).Select(&tasks, sql, task.Repo.URI, task.Attempt, task.CommitID, task.TaskID); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks[0].toBuildTask(), nil
}
