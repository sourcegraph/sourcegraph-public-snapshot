package localstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func init() {
	AppSchema.Map.AddTableWithName(dbThread{}, "threads").SetKeys(true, "ID")
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		"CREATE INDEX ON threads(local_repo_id, file);",
	)
}

// dbThread DB-maps a sourcegraph.Thread object.
type dbThread struct {
	ID             int64
	LocalRepoID    int64 `db:"local_repo_id"`
	File           string
	Revision       string
	StartLine      int32     `db:"start_line"`
	EndLine        int32     `db:"end_line"`
	StartCharacter int32     `db:"start_character"`
	EndCharacter   int32     `db:"end_character"`
	CreatedAt      time.Time `db:"created_at"`
}

func (t *dbThread) fromThread(t2 *sourcegraph.Thread) {
	t.ID = int64(t2.ID)
	t.LocalRepoID = int64(t2.LocalRepoID)
	t.File = t2.File
	t.Revision = t2.Revision
	t.StartLine = int32(t2.StartLine)
	t.EndLine = int32(t2.EndLine)
	t.StartCharacter = int32(t2.StartCharacter)
	t.EndCharacter = int32(t2.EndCharacter)
}

func (t *dbThread) toThread() *sourcegraph.Thread {
	t2 := &sourcegraph.Thread{}
	t2.ID = int32(t.ID)
	t2.LocalRepoID = int32(t.LocalRepoID)
	t2.File = t.File
	t2.Revision = t.Revision
	t2.StartLine = t.StartLine
	t2.EndLine = (t.EndLine)
	t2.StartCharacter = (t.StartCharacter)
	t2.EndCharacter = (t.EndCharacter)
	t2.CreatedAt = t.CreatedAt
	return t2
}

type threads struct{}

func (*threads) Create(ctx context.Context, newThread *sourcegraph.Thread) (*sourcegraph.Thread, error) {
	if Mocks.Threads.Create != nil {
		return Mocks.Threads.Create(ctx, newThread)
	}

	var t dbThread
	t.fromThread(newThread)
	t.CreatedAt = time.Now()
	err := appDBH(ctx).Insert(&t)
	if err != nil {
		return nil, err
	}

	return t.toThread(), nil
}

func (*threads) Get(ctx context.Context, id int64) (*sourcegraph.Thread, error) {
	if Mocks.Threads.Get != nil {
		return Mocks.Threads.Get(ctx, id)
	}

	t := dbThread{}
	err := appDBH(ctx).SelectOne(&t, "SELECT * FROM threads WHERE (id=$1)", id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("thread does not exist")
	} else if err != nil {
		return nil, err
	}

	return t.toThread(), nil
}

func (*threads) GetAllForFile(ctx context.Context, repoID int64, file string) ([]*sourcegraph.Thread, error) {
	ts := []*dbThread{}
	_, err := appDBH(ctx).Select(&ts, "SELECT * FROM threads WHERE (local_repo_id=$1 AND file=$2)", repoID, file)
	if err != nil {
		return nil, err
	}

	threads := []*sourcegraph.Thread{}
	for _, t := range ts {
		threads = append(threads, t.toThread())
	}
	return threads, nil
}
