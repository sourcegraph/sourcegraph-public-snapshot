package pgsql

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/sqs/modl"
	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/dbutil"
)

func init() {
	t := Schema.Map.AddTableWithName(dbBuild{}, "repo_build").SetKeys(false, "repo", "id")
	t.ColMap("commit_id").SetMaxSize(40)
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE repo_build ALTER COLUMN started_at TYPE timestamp with time zone USING started_at::timestamp with time zone;`,
		`ALTER TABLE repo_build ALTER COLUMN ended_at TYPE timestamp with time zone USING ended_at::timestamp with time zone;`,
		`ALTER TABLE repo_build ALTER COLUMN heartbeat_at TYPE timestamp with time zone USING ended_at::timestamp with time zone;`,
		`ALTER TABLE repo_build ALTER COLUMN builder_config TYPE text;`,
		`CREATE INDEX repo_build_repo ON repo_build(repo);`,
		`CREATE INDEX repo_build_priority ON repo_build(priority);`,
		`create index repo_build_created_at on repo_build(created_at desc nulls last);`,
		`create index repo_build_updated_at on repo_build((greatest(started_at, ended_at, created_at)) desc nulls last);`,
		`create index repo_build_successful on repo_build(repo, commit_id) where success and not purged;`,

		// Set id to 1 + the max previous build ID for the repo.
		`CREATE OR REPLACE FUNCTION increment_build_id() RETURNS trigger IMMUTABLE AS $$
         BEGIN
           IF NEW.id = 0 OR NEW.id IS NULL THEN
             NEW.id = (SELECT COALESCE(max(b.id), 0) + 1 FROM repo_build b WHERE b.repo=NEW.repo);
           END IF;
           RETURN NEW;
         END
         $$ language plpgsql;`,
		`CREATE TRIGGER repo_build_next_id BEFORE INSERT ON repo_build FOR EACH ROW EXECUTE PROCEDURE increment_build_id();`,
	)

	Schema.Map.AddTableWithName(dbBuildTask{}, "repo_build_task").SetKeys(false, "repo", "build_id", "id")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`ALTER TABLE repo_build_task ALTER COLUMN started_at TYPE timestamp with time zone USING started_at::timestamp with time zone;`,
		`ALTER TABLE repo_build_task ALTER COLUMN ended_at TYPE timestamp with time zone USING ended_at::timestamp with time zone;`,

		// Set id to 1 + the max previous task ID for the build.
		`CREATE OR REPLACE FUNCTION increment_build_task_id() RETURNS trigger IMMUTABLE AS $$
         BEGIN
           IF NEW.id = 0 OR NEW.id IS NULL THEN
             NEW.id = (SELECT COALESCE(max(t.id), 0) + 1 FROM repo_build_task t WHERE t.repo=NEW.repo AND t.build_id=NEW.build_id);
           END IF;
           RETURN NEW;
         END
         $$ language plpgsql;`,
		`CREATE TRIGGER repo_build_task_next_id BEFORE INSERT ON repo_build_task FOR EACH ROW EXECUTE PROCEDURE increment_build_task_id();`,
	)
}

// dbBuild DB-maps a sourcegraph.Build object.
type dbBuild struct {
	ID            uint64
	Repo          string
	CommitID      string `db:"commit_id"`
	Branch        string
	Tag           string
	CreatedAt     time.Time  `db:"created_at"`
	StartedAt     *time.Time `db:"started_at"`
	EndedAt       *time.Time `db:"ended_at"`
	HeartbeatAt   *time.Time `db:"heartbeat_at"`
	Success       bool
	Failure       bool
	Killed        bool
	Host          string
	Purged        bool
	Queue         bool
	Priority      int
	BuilderConfig string `db:"builder_config"`
}

func (b *dbBuild) toBuild() *sourcegraph.Build {
	return &sourcegraph.Build{
		ID:          b.ID,
		Repo:        b.Repo,
		CommitID:    b.CommitID,
		Branch:      b.Branch,
		Tag:         b.Tag,
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
			Queue:         b.Queue,
			Priority:      int32(b.Priority),
			BuilderConfig: b.BuilderConfig,
		},
	}
}

func (b *dbBuild) fromBuild(b2 *sourcegraph.Build) {
	b.ID = b2.ID
	b.Repo = b2.Repo
	b.CommitID = b2.CommitID
	b.Branch = b2.Branch
	b.Tag = b2.Tag
	b.CreatedAt = b2.CreatedAt.Time()
	b.StartedAt = tm(b2.StartedAt)
	b.EndedAt = tm(b2.EndedAt)
	b.HeartbeatAt = tm(b2.HeartbeatAt)
	b.Success = b2.Success
	b.Failure = b2.Failure
	b.Killed = b2.Killed
	b.Host = b2.Host
	b.Purged = b2.Purged
	b.Queue = b2.Queue
	b.Priority = int(b2.Priority)
	b.BuilderConfig = b2.BuilderConfig
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
	ID        uint64
	Repo      string
	BuildID   uint64 `db:"build_id"`
	ParentID  uint64 `db:"parent_id"`
	Label     string
	CreatedAt time.Time  `db:"created_at"`
	StartedAt *time.Time `db:"started_at"`
	EndedAt   *time.Time `db:"ended_at"`
	Success   bool
	Failure   bool
	Skipped   bool
	Warnings  bool
}

func (t *dbBuildTask) toBuildTask() *sourcegraph.BuildTask {
	return &sourcegraph.BuildTask{
		ID:        t.ID,
		Build:     sourcegraph.BuildSpec{Repo: sourcegraph.RepoSpec{URI: t.Repo}, ID: t.BuildID},
		ParentID:  t.ParentID,
		Label:     t.Label,
		CreatedAt: pbtypes.NewTimestamp(t.CreatedAt),
		StartedAt: ts(t.StartedAt),
		EndedAt:   ts(t.EndedAt),
		Success:   t.Success,
		Failure:   t.Failure,
		Skipped:   t.Skipped,
		Warnings:  t.Warnings,
	}
}

func (t *dbBuildTask) fromBuildTask(t2 *sourcegraph.BuildTask) {
	t.ID = t2.ID
	t.Repo = t2.Build.Repo.URI
	t.BuildID = t2.Build.ID
	t.ParentID = t2.ParentID
	t.Label = t2.Label
	t.CreatedAt = t2.CreatedAt.Time()
	t.StartedAt = tm(t2.StartedAt)
	t.EndedAt = tm(t2.EndedAt)
	t.Success = t2.Success
	t.Failure = t2.Failure
	t.Skipped = t2.Skipped
	t.Warnings = t2.Warnings
}

func toBuildTasks(ts []*dbBuildTask) []*sourcegraph.BuildTask {
	t2s := make([]*sourcegraph.BuildTask, len(ts))
	for i, t := range ts {
		t2s[i] = t.toBuildTask()
	}
	return t2s
}

// builds is a DB-backed implementation of the Builds store.
type builds struct{}

var _ store.Builds = (*builds)(nil)

func (s *builds) Get(ctx context.Context, buildSpec sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
	var builds []*dbBuild
	err := dbh(ctx).Select(&builds, `SELECT * FROM repo_build WHERE id=$1 AND repo=$2 LIMIT 1;`, buildSpec.ID, buildSpec.Repo.URI)
	if err != nil {
		return nil, err
	}
	if len(builds) != 1 {
		return nil, grpc.Errorf(codes.NotFound, "build %s not found", buildSpec.IDString())
	}
	return builds[0].toBuild(), nil
}

func (s *builds) List(ctx context.Context, opt *sourcegraph.BuildListOptions) ([]*sourcegraph.Build, error) {
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
		"build":      "b.repo %(direction)s, b.commit_id %(direction)s, b.id %(direction)s",
		"created_at": "b.created_at %(direction)s NULLS LAST",
		"started_at": "b.started_at %(direction)s NULLS LAST",
		"ended_at":   "b.ended_at %(direction)s NULLS LAST",
		"updated_at": "greatest(b.started_at, b.ended_at, b.created_at) %(direction)s NULLS LAST",
		"priority":   "b.priority %(direction)s, greatest(b.started_at, b.ended_at, b.created_at) ASC NULLS LAST",
	}
	if sortCol, valid := sortKeyToCol[sort]; valid {
		sort = sortCol
	} else {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid sort: "+sort)
	}
	if direction != "asc" && direction != "desc" {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid direction: "+direction)
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

func (s *builds) GetFirstInCommitOrder(ctx context.Context, repo string, commitIDs []string, successfulOnly bool) (build *sourcegraph.Build, nth int, err error) {
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

func (s *builds) Create(ctx context.Context, newBuild *sourcegraph.Build) (*sourcegraph.Build, error) {
	var b dbBuild
	b.fromBuild(newBuild)

	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return modl.PostgresDialect{}.BindVar(len(args) - 1)
	}

	// Construct SQL manually so we can retrieve the id # from
	// the DB trigger.
	sql := `INSERT INTO repo_build(id, repo, commit_id, branch, tag, created_at, started_at, ended_at, heartbeat_at,
                                   success, failure, killed, host, purged, queue, priority, builder_config)
            VALUES(` + arg(b.ID) + `, ` + arg(b.Repo) + `, ` + arg(b.CommitID) + `, ` + arg(b.Branch) + `, ` + arg(b.Tag) + `, ` + arg(b.CreatedAt) + `, ` + arg(b.StartedAt) + `,` +
		arg(b.EndedAt) + `,` + arg(b.HeartbeatAt) + `, ` + arg(b.Success) + `, ` + arg(b.Failure) + `, ` + arg(b.Killed) + `, ` +
		arg(b.Host) + `, ` + arg(b.Purged) + `, ` + arg(b.Queue) + `, ` + arg(b.Priority) + `, ` + arg(b.BuilderConfig) + `)
            RETURNING id;`
	id, err := dbutil.SelectInt(dbh(ctx), sql, args...)
	if err != nil {
		return nil, err
	}
	b.ID = uint64(id)
	return b.toBuild(), nil
}

func (s *builds) Update(ctx context.Context, build sourcegraph.BuildSpec, info sourcegraph.BuildUpdate) error {
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
		sql := fmt.Sprintf(`UPDATE repo_build SET %s WHERE id=%s AND repo=%s`, strings.Join(updates, ", "), arg(build.ID), arg(build.Repo.URI))

		if _, err := dbh(ctx).Exec(sql, args...); err != nil {
			return err
		}
	}

	return nil
}

func (s *builds) CreateTasks(ctx context.Context, tasks []*sourcegraph.BuildTask) ([]*sourcegraph.BuildTask, error) {
	created := make([]*dbBuildTask, len(tasks))
	for i, task := range tasks {
		var args []interface{}
		arg := func(v interface{}) string {
			args = append(args, v)
			return modl.PostgresDialect{}.BindVar(len(args) - 1)
		}

		created[i] = &dbBuildTask{}
		created[i].fromBuildTask(task)

		// Construct SQL manually so we can retrieve the id # from
		// the DB trigger.
		t := created[i] // shorter alias
		sql := `INSERT INTO repo_build_task(id, repo, build_id, parent_id, label, created_at, started_at, ended_at, success, failure, skipped, warnings)
            VALUES(` + arg(t.ID) + `, ` + arg(t.Repo) + `, ` + arg(t.BuildID) + `, ` + arg(t.ParentID) + `, ` + arg(t.Label) + `, ` + arg(t.CreatedAt) + `, ` + arg(t.StartedAt) + `,` + arg(t.EndedAt) + `,` + arg(t.Success) + `, ` + arg(t.Failure) + `, ` + arg(t.Skipped) + `, ` + arg(t.Warnings) + `)
            RETURNING id;`
		id, err := dbutil.SelectInt(dbh(ctx), sql, args...)
		if err != nil {
			return nil, err
		}
		created[i].ID = uint64(id)
	}
	return toBuildTasks(created), nil
}

func (s *builds) UpdateTask(ctx context.Context, task sourcegraph.TaskSpec, info sourcegraph.TaskUpdate) error {
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
	if info.Skipped {
		updates = append(updates, "skipped="+arg(info.Skipped))
	}
	if info.Warnings {
		updates = append(updates, "warnings="+arg(info.Warnings))
	}

	if len(updates) != 0 {
		sql := `UPDATE repo_build_task SET ` + strings.Join(updates, ", ") + ` WHERE id=` + arg(task.ID)
		if _, err := dbh(ctx).Exec(sql, args...); err != nil {
			return err
		}
	}

	return nil
}

func (s *builds) ListBuildTasks(ctx context.Context, build sourcegraph.BuildSpec, opt *sourcegraph.BuildTaskListOptions) ([]*sourcegraph.BuildTask, error) {
	if opt == nil {
		opt = &sourcegraph.BuildTaskListOptions{}
	}

	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return modl.PostgresDialect{}.BindVar(len(args) - 1)
	}

	conds := []string{"build_id=" + arg(build.ID), "repo=" + arg(build.Repo.URI)}
	condsSQL := strings.Join(conds, " AND ")

	sql := `-- Builds.ListBuildTasks
SELECT * FROM repo_build_task
WHERE ` + condsSQL + `
ORDER BY id ASC
LIMIT ` + arg(opt.Limit()) + ` OFFSET ` + arg(opt.Offset()) + `;`
	var tasks []*dbBuildTask
	if err := dbh(ctx).Select(&tasks, sql, args...); err != nil {
		return nil, err
	}
	return toBuildTasks(tasks), nil
}

func (s *builds) DequeueNext(ctx context.Context) (*sourcegraph.Build, error) {
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
WHERE repo_build.repo = to_dequeue.repo AND repo_build.id = to_dequeue.id
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

func (s *builds) GetTask(ctx context.Context, task sourcegraph.TaskSpec) (*sourcegraph.BuildTask, error) {
	var tasks []*dbBuildTask
	sql := `SELECT * FROM repo_build_task WHERE repo=$1 AND build_id=$2 AND id=$3;`
	if err := dbh(ctx).Select(&tasks, sql, task.Build.Repo.URI, task.Build.ID, task.ID); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks[0].toBuildTask(), nil
}
