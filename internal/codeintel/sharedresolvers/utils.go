package sharedresolvers

import (
	"encoding/base64"
	"encoding/json"
	"io/fs"
	"os"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
	autoindexingShared "github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DateTime implements the DateTime GraphQL scalar type.
type DateTime struct{ time.Time }

// DateTimeOrNil is a helper function that returns nil for time == nil and otherwise wraps time in
// DateTime.
func DateTimeOrNil(time *time.Time) *DateTime {
	if time == nil {
		return nil
	}
	return &DateTime{Time: *time}
}

func (DateTime) ImplementsGraphQLType(name string) bool {
	return name == "DateTime"
}

func (v DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Time.Format(time.RFC3339))
}

func (v *DateTime) UnmarshalGraphQL(input any) error {
	s, ok := input.(string)
	if !ok {
		return errors.Errorf("invalid GraphQL DateTime scalar value input (got %T, expected string)", input)
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*v = DateTime{Time: t}
	return nil
}

// strPtr creates a pointer to the given value. If the value is an
// empty string, a nil pointer is returned.
func strPtr(val string) *string {
	if val == "" {
		return nil
	}

	return &val
}

func marshalLSIFIndexGQLID(indexID int64) graphql.ID {
	return relay.MarshalID("LSIFIndex", indexID)
}

// toInt32 translates the given int pointer into an int32 pointer.
func toInt32(val *int) *int32 {
	if val == nil {
		return nil
	}

	v := int32(*val)
	return &v
}

func marshalLSIFUploadGQLID(uploadID int64) graphql.ID {
	return relay.MarshalID("LSIFUpload", uploadID)
}

func unmarshalConfigurationPolicyGQLID(id graphql.ID) (configurationPolicyID int64, err error) {
	err = relay.UnmarshalSpec(id, &configurationPolicyID)
	return configurationPolicyID, err
}

type PageInfo struct {
	endCursor   *string
	hasNextPage bool
}

// HasNextPage returns a new PageInfo with the given hasNextPage value.
func HasNextPage(hasNextPage bool) *PageInfo {
	return &PageInfo{hasNextPage: hasNextPage}
}

// NextPageCursor returns a new PageInfo indicating there is a next page with
// the given end cursor.
func NextPageCursor(endCursor string) *PageInfo {
	return &PageInfo{endCursor: &endCursor, hasNextPage: true}
}

func (r *PageInfo) EndCursor() *string { return r.endCursor }
func (r *PageInfo) HasNextPage() bool  { return r.hasNextPage }

// EncodeIntCursor creates a PageInfo object from the given new offset value. If the
// new offset value, then an object indicating the end of the result set is returned.
// The cursor is base64 encoded for transfer, and should be decoded using the function
// decodeIntCursor.
func EncodeIntCursor(val *int32) *PageInfo {
	if val == nil {
		return EncodeCursor(nil)
	}

	str := strconv.FormatInt(int64(*val), 10)
	return EncodeCursor(&str)
}

// EncodeCursor creates a PageInfo object from the given cursor. If the cursor is not
// defined, then an object indicating the end of the result set is returned. The cursor
// is base64 encoded for transfer, and should be decoded using the function decodeCursor.
func EncodeCursor(val *string) *PageInfo {
	if val != nil {
		return NextPageCursor(base64.StdEncoding.EncodeToString([]byte(*val)))
	}

	return HasNextPage(false)
}

// NextOffset determines the offset that should be used for a subsequent request.
// If there are no more results in the paged result set, this function returns nil.
func NextOffset(offset, count, totalCount int) *int {
	if offset+count < totalCount {
		val := offset + count
		return &val
	}

	return nil
}

func CreateFileInfo(path string, isDir bool) fs.FileInfo {
	return fileInfo{path: path, isDir: isDir}
}

type fileInfo struct {
	path  string
	size  int64
	isDir bool
}

func (f fileInfo) Name() string { return f.path }
func (f fileInfo) Size() int64  { return f.size }
func (f fileInfo) IsDir() bool  { return f.isDir }
func (f fileInfo) Mode() os.FileMode {
	if f.IsDir() {
		return os.ModeDir
	}
	return 0
}
func (f fileInfo) ModTime() time.Time { return time.Now() }
func (f fileInfo) Sys() any           { return any(nil) }

func convertSharedIndexesWithRepositoryNamespaceToDBStoreIndexesWithRepositoryNamespace(shared autoindexingShared.IndexesWithRepositoryNamespace) dbstore.IndexesWithRepositoryNamespace {
	indexes := make([]dbstore.Index, 0, len(shared.Indexes))
	for _, index := range shared.Indexes {
		indexes = append(indexes, convertSharedIndexToDBStoreIndex(index))
	}

	return dbstore.IndexesWithRepositoryNamespace{
		Root:    shared.Root,
		Indexer: shared.Indexer,
		Indexes: indexes,
	}
}

func convertSharedIndexToDBStoreIndex(index types.Index) store.Index {
	dockerSteps := make([]store.DockerStep, 0, len(index.DockerSteps))
	for _, step := range index.DockerSteps {
		dockerSteps = append(dockerSteps, store.DockerStep(step))
	}

	executionLogs := make([]workerutil.ExecutionLogEntry, 0, len(index.ExecutionLogs))
	for _, log := range index.ExecutionLogs {
		executionLogs = append(executionLogs, workerutil.ExecutionLogEntry(log))
	}

	return store.Index{
		ID:                 index.ID,
		Commit:             index.Commit,
		QueuedAt:           index.QueuedAt,
		State:              index.State,
		FailureMessage:     index.FailureMessage,
		StartedAt:          index.StartedAt,
		FinishedAt:         index.FinishedAt,
		ProcessAfter:       index.ProcessAfter,
		NumResets:          index.NumResets,
		NumFailures:        index.NumFailures,
		RepositoryID:       index.RepositoryID,
		LocalSteps:         index.LocalSteps,
		RepositoryName:     index.RepositoryName,
		DockerSteps:        dockerSteps,
		Root:               index.Root,
		Indexer:            index.Indexer,
		IndexerArgs:        index.IndexerArgs,
		Outfile:            index.Outfile,
		ExecutionLogs:      executionLogs,
		Rank:               index.Rank,
		AssociatedUploadID: index.AssociatedUploadID,
	}
}

func UnmarshalRepositoryID(id graphql.ID) (repo api.RepoID, err error) {
	err = relay.UnmarshalSpec(id, &repo)
	return
}

// gitCommitGQLID is a type used for marshaling and unmarshaling a Git commit's
// GraphQL ID.
type gitCommitGQLID struct {
	Repository graphql.ID  `json:"r"`
	CommitID   GitObjectID `json:"c"`
}

func marshalGitCommitID(repo graphql.ID, commitID GitObjectID) graphql.ID {
	return relay.MarshalID("GitCommit", gitCommitGQLID{Repository: repo, CommitID: commitID})
}
