package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/google/uuid"

	_ "modernc.org/sqlite" // pure Go SQLite implementation

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

const schemaVersion = "1"

type key int

const (
	invocationKey key = 0
)

type invocation struct {
	uuid uuid.UUID
}

var store = sync.OnceValue(func() analyticsStore {
	db, err := newDiskStore()
	if err != nil {
		std.Out.WriteWarningf("Failed to create sg analytics store: %s", err)
	}
	return analyticsStore{db: db}
})

func newDiskStore() (Execer, error) {
	sghome, err := root.GetSGHomePath()
	if err != nil {
		return nil, err
	}

	// this will create the file if it doesnt exist
	db, err := sql.Open("sqlite", "file://"+path.Join(sghome, "analytics.sqlite"))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	rdb := retryableConn{db}

	_, err = rdb.Exec(`CREATE TABLE IF NOT EXISTS analytics (
		event_uuid TEXT PRIMARY KEY,
		-- invocation_uuid TEXT,
		schema_version TEXT NOT NULL,
		metadata_json TEXT
	)`)
	if err != nil {
		return nil, err
	}

	return &rdb, nil
}

type analyticsStore struct {
	db Execer
}

func (s analyticsStore) NewInvocation(ctx context.Context, uuid uuid.UUID, version string, meta map[string]any) {
	if s.db == nil {
		return
	}

	meta["user_id"] = getEmail()
	meta["start_time"] = time.Now()

	b, err := json.Marshal(meta)
	if err != nil {
		std.Out.WriteWarningf("Invalid json generated for sg analytics metadata %v: %s", meta, err)
	}

	_, err = s.db.Exec(`INSERT INTO analytics (event_uuid, schema_version, metadata_json) VALUES (?, ?, ?)`, uuid, schemaVersion, string(b))
	if err != nil {
		std.Out.WriteWarningf("Failed to insert sg analytics event: %s", err)
	}
}

func (s analyticsStore) AddMetadata(ctx context.Context, uuid uuid.UUID, meta map[string]any) {
	if s.db == nil {
		return
	}

	b, err := json.Marshal(meta)
	if err != nil {
		std.Out.WriteWarningf("Invalid json generated for sg analytics metadata %v: %s", meta, err)
		return
	}

	_, err = s.db.Exec(`UPDATE analytics SET metadata_json = json_patch(metadata_json, ?) WHERE event_uuid = ?`, string(b), uuid)
	if err != nil {
		std.Out.WriteWarningf("Failed to update sg analytics event: %s", err)
	}
}

// Dont invoke this function directly. Use the `getEmail` function instead.
func emailfunc() string {
	sgHome, err := root.GetSGHomePath()
	if err != nil {
		return "anonymous"
	}

	b, err := os.ReadFile(path.Join(sgHome, "whoami.json"))
	if err != nil {
		return "anonymous"
	}
	var whoami struct {
		Email string
	}
	if err := json.Unmarshal(b, &whoami); err != nil {
		return "anonymous"
	}
	return whoami.Email
}

var getEmail = sync.OnceValue[string](emailfunc)

func NewInvocation(ctx context.Context, version string, meta map[string]any) context.Context {
	// v7 for sortable property (not vital as we also store timestamps, but no harm to have)
	u, _ := uuid.NewV7()
	invc := invocation{u}

	store().NewInvocation(ctx, u, version, meta)

	return context.WithValue(ctx, invocationKey, invc)
}

func AddMeta(ctx context.Context, meta map[string]any) {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return
	}

	store().AddMetadata(ctx, invc.uuid, meta)
}

func InvocationSucceeded(ctx context.Context) {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return
	}

	store().AddMetadata(ctx, invc.uuid, map[string]any{
		"success":  true,
		"end_time": time.Now(),
	})
}

func InvocationCancelled(ctx context.Context) {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return
	}

	store().AddMetadata(ctx, invc.uuid, map[string]any{
		"cancelled": true,
		"end_time":  time.Now(),
	})
}

func InvocationFailed(ctx context.Context, err error) {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return
	}

	store().AddMetadata(ctx, invc.uuid, map[string]any{
		"failed":   true,
		"error":    err.Error(),
		"end_time": time.Now(),
	})
}

func InvocationPanicked(ctx context.Context, err any) {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return
	}

	store().AddMetadata(ctx, invc.uuid, map[string]any{
		"panicked": true,
		"error":    fmt.Sprint(err),
		"end_time": time.Now(),
	})
}
