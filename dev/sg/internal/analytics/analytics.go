package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"

	_ "modernc.org/sqlite" // pure Go SQLite implementation

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrDBNotInitialized = errors.New("analytics database not initialized")

const schemaVersion = "1"

type key int

const (
	invocationKey key = 0
)

type invocation struct {
	uuid     uuid.UUID
	metadata map[string]any
}

func (i invocation) GetStartTime() *time.Time {
	v, ok := i.metadata["start_time"]
	if !ok {
		return nil
	}
	raw := v.(string)
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &t
}

func (i invocation) GetEndTime() *time.Time {
	v, ok := i.metadata["end_time"]
	if !ok {
		return nil
	}
	raw := v.(string)
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil
	}
	return &t
}

func (i invocation) GetDuration() time.Duration {
	start := i.GetStartTime()
	end := i.GetEndTime()

	if start == nil || end == nil {
		return 0
	}

	return end.Sub(*start)
}

func (i invocation) IsSuccess() bool {
	v, ok := i.metadata["success"]
	if !ok {
		return false
	}
	return v.(bool)
}

func (i invocation) IsCancelled() bool {
	v, ok := i.metadata["cancelled"]
	if !ok {
		return false
	}
	return v.(bool)
}

func (i invocation) IsFailed() bool {
	v, ok := i.metadata["failed"]
	if !ok {
		return false
	}
	return v.(bool)
}

func (i invocation) IsPanicked() bool {
	v, ok := i.metadata["panicked"]
	if !ok {
		return false
	}
	return v.(bool)
}

func (i invocation) GetCommand() string {
	v, ok := i.metadata["command"]
	if !ok {
		return ""
	}
	return v.(string)
}

func (i invocation) GetVersion() string {
	v, ok := i.metadata["version"]
	if !ok {
		return ""
	}
	return v.(string)
}

func (i invocation) GetError() string {
	v, ok := i.metadata["error"]
	if !ok {
		return ""
	}
	return v.(string)
}

func (i invocation) GetUserID() string {
	v, ok := i.metadata["user_id"]
	if !ok {
		return ""
	}
	return v.(string)
}

func (i invocation) GetFlags() map[string]any {
	v, ok := i.metadata["flags"]
	if !ok {
		return nil
	}
	return v.(map[string]any)
}

func (i invocation) GetArgs() []any {
	v, ok := i.metadata["args"]
	if !ok {
		return nil
	}
	return v.([]any)
}

func (i invocation) GetOS() string {
	v, ok := i.metadata["os"]
	if !ok {
		return ""
	}
	return v.(string)
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

func (s analyticsStore) NewInvocation(_ context.Context, uuid uuid.UUID, version string, meta map[string]any) error {
	if s.db == nil {
		return ErrDBNotInitialized
	}

	meta["user_id"] = getEmail()
	meta["version"] = version
	meta["start_time"] = time.Now()
	meta["os"] = getOS()

	b, err := json.Marshal(meta)
	if err != nil {
		return errors.Wrapf(err, "failed to JSON marshal metadata %v")
	}

	_, err = s.db.Exec(`INSERT INTO analytics (event_uuid, schema_version, metadata_json) VALUES (?, ?, ?)`, uuid, schemaVersion, string(b))
	if err != nil {
		return errors.Wrapf(err, "failed to insert sg analytics event")
	}

	return nil
}

func (s analyticsStore) AddMetadata(_ context.Context, uuid uuid.UUID, meta map[string]any) error {
	if s.db == nil {
		return ErrDBNotInitialized
	}

	b, err := json.Marshal(meta)
	if err != nil {
		return errors.Wrapf(err, "failed to JSON marshal metadata %v")
	}

	_, err = s.db.Exec(`UPDATE analytics SET metadata_json = json_patch(metadata_json, ?) WHERE event_uuid = ?`, string(b), uuid)
	if err != nil {
		return errors.Wrapf(err, "failed to update sg analytics event")
	}

	return nil
}

func (s analyticsStore) DeleteInvocation(_ context.Context, uuid string) error {
	if s.db == nil {
		return ErrDBNotInitialized
	}

	_, err := s.db.Exec(`DELETE FROM analytics WHERE event_uuid = ?`, uuid)
	if err != nil {
		return errors.Wrapf(err, "failed to delete sg analytics event")
	}

	return nil
}

func (s analyticsStore) ListCompleted(_ context.Context) ([]invocation, error) {
	if s.db == nil {
		return nil, nil
	}

	res, err := s.db.Query(`SELECT * FROM analytics WHERE json_extract(metadata_json, '$.end_time') IS NOT NULL LIMIT 10`)
	if err != nil {
		return nil, err
	}

	results := []invocation{}

	for res.Next() {
		invc := invocation{metadata: map[string]any{}}
		var eventUUID string
		var schemaVersion string
		var metadata_json string
		if err := res.Scan(&eventUUID, &schemaVersion, &metadata_json); err != nil {
			return nil, err
		}

		invc.uuid, err = uuid.Parse(eventUUID)
		if err != nil {
			return nil, err
		}

		err := json.Unmarshal([]byte(metadata_json), &invc.metadata)
		if err != nil {
			return nil, err
		}

		results = append(results, invc)
	}

	return results, err

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

func osFunc() string {
	userOS := runtime.GOOS
	switch userOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	default:
		// weird case, but catching it instead of throwing an error
		return userOS
	}
}

// Dont invoke this function directly. Use the `getOS` function instead.
var getOS = sync.OnceValue[string](osFunc)

func NewInvocation(ctx context.Context, version string, meta map[string]any) (context.Context, error) {
	// v7 for sortable property (not vital as we also store timestamps, but no harm to have)
	u, _ := uuid.NewV7()
	invc := invocation{u, meta}

	if err := store().NewInvocation(ctx, u, version, meta); err != nil && !errors.Is(err, ErrDBNotInitialized) {
		return nil, err
	}

	return context.WithValue(ctx, invocationKey, invc), nil
}

func AddMeta(ctx context.Context, meta map[string]any) error {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return nil
	}

	return store().AddMetadata(ctx, invc.uuid, meta)
}

func InvocationSucceeded(ctx context.Context) error {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return nil
	}

	return store().AddMetadata(ctx, invc.uuid, map[string]any{
		"success":  true,
		"end_time": time.Now(),
	})

}

func InvocationCancelled(ctx context.Context) error {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return nil
	}

	return store().AddMetadata(ctx, invc.uuid, map[string]any{
		"cancelled": true,
		"end_time":  time.Now(),
	})

}

func InvocationFailed(ctx context.Context, err error) error {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return nil
	}

	return store().AddMetadata(ctx, invc.uuid, map[string]any{
		"failed":   true,
		"error":    err.Error(),
		"end_time": time.Now(),
	})
}

func InvocationPanicked(ctx context.Context, err any) error {
	invc, ok := ctx.Value(invocationKey).(invocation)
	if !ok {
		return nil
	}

	return store().AddMetadata(ctx, invc.uuid, map[string]any{
		"panicked": true,
		"error":    fmt.Sprint(err),
		"end_time": time.Now(),
	})
}
