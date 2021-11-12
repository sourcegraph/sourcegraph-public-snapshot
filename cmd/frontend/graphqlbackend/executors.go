package graphqlbackend

import (
	"context"
	"encoding/base64"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type executorResolver struct {
	executor types.Executor
}

type executorConnectionResolver struct {
	resolvers  []*executorResolver
	totalCount int
	nextOffset *int
}

const DefaultExecutorsLimit = 50

func (r *schemaResolver) Executors(ctx context.Context, args *struct {
	Query  *string
	Active *bool
	First  *int32
	After  *string
}) (*executorConnectionResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	query := ""
	if args.Query != nil {
		query = *args.Query
	}

	active := false
	if args.Active != nil {
		active = *args.Active
	}

	offset, err := decodeIntCursor(args.After)
	if err != nil {
		return nil, err
	}

	limit := DefaultExecutorsLimit
	if args.First != nil {
		limit = int(*args.First)
	}

	executors, totalCount, err := r.db.Executors().List(ctx, database.ExecutorStoreListOptions{
		Query:  query,
		Active: active,
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	resolvers := make([]*executorResolver, 0, len(executors))
	for _, executor := range executors {
		resolvers = append(resolvers, &executorResolver{executor: executor})
	}

	return &executorConnectionResolver{
		resolvers:  resolvers,
		totalCount: totalCount,
		nextOffset: nextOffset(offset, len(executors), totalCount),
	}, nil
}

func (r *executorConnectionResolver) Nodes(ctx context.Context) []*executorResolver {
	return r.resolvers
}

func (r *executorConnectionResolver) TotalCount(ctx context.Context) int32 {
	return int32(r.totalCount)
}

func (r *executorConnectionResolver) PageInfo(ctx context.Context) *graphqlutil.PageInfo {
	return encodeIntCursor(toInt32(r.nextOffset))
}

func marshalExecutorID(id int64) graphql.ID {
	return relay.MarshalID("Executor", id)
}

func unmarshalExecutorID(id graphql.ID) (executorID int64, err error) {
	err = relay.UnmarshalSpec(id, &executorID)
	return
}

func executorByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*executorResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	id, err := unmarshalExecutorID(gqlID)
	if err != nil {
		return nil, err
	}

	executor, ok, err := db.Executors().GetByID(ctx, int(id))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &executorResolver{executor: executor}, nil
}

func (e *executorResolver) ID() graphql.ID {
	return marshalExecutorID(int64(e.executor.ID))
}

func (e *executorResolver) Hostname() string {
	return e.executor.Hostname
}

func (e *executorResolver) QueueName() string {
	return e.executor.QueueName
}

func (e *executorResolver) Os() string {
	return e.executor.OS
}

func (e *executorResolver) Architecture() string {
	return e.executor.Architecture
}

func (e *executorResolver) ExecutorVersion() string {
	return e.executor.ExecutorVersion
}

func (e *executorResolver) SrcCliVersion() string {
	return e.executor.SrcCliVersion
}

func (e *executorResolver) DockerVersion() string {
	return e.executor.DockerVersion
}

func (e *executorResolver) GitVersion() string {
	return e.executor.GitVersion
}

func (e *executorResolver) IgniteVersion() string {
	return e.executor.IgniteVersion
}

func (e *executorResolver) FirstSeenAt() DateTime {
	return DateTime{e.executor.FirstSeenAt}
}

func (e *executorResolver) LastSeenAt() DateTime {
	return DateTime{e.executor.LastSeenAt}
}

//
// TODO:
// Deduplicate from codeintel/util.go

// toInt32 translates the given int pointer into an int32 pointer.
func toInt32(val *int) *int32 {
	if val == nil {
		return nil
	}

	v := int32(*val)
	return &v
}

//
// TODO:
// Deduplicate from codeintel/util.go

// nextOffset determines the offset that should be used for a subsequent request.
// If there are no more results in the paged result set, this function returns nil.
func nextOffset(offset, count, totalCount int) *int {
	if offset+count < totalCount {
		val := offset + count
		return &val
	}

	return nil
}

//
// TODO:
// Deduplicate from codeintel/cursors.go

// encodeCursor creates a PageInfo object from the given cursor. If the cursor is not
// defined, then an object indicating the end of the result set is returned. The cursor
// is base64 encoded for transfer, and should be decoded using the function decodeCursor.
func encodeCursor(val *string) *graphqlutil.PageInfo {
	if val != nil {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(*val)))
	}

	return graphqlutil.HasNextPage(false)
}

// decodeCursor decodes the given cursor value. It is assumed to be a value previously
// returned from the function encodeCursor. An empty string is returned if no cursor is
// supplied. Invalid cursors return errors.
func decodeCursor(val *string) (string, error) {
	if val == nil {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*val)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// encodeIntCursor creates a PageInfo object from the given new offset value. If the
// new offset value, then an object indicating the end of the result set is returned.
// The cursor is base64 encoded for transfer, and should be decoded using the function
// decodeIntCursor.
func encodeIntCursor(val *int32) *graphqlutil.PageInfo {
	if val == nil {
		return encodeCursor(nil)
	}

	str := strconv.FormatInt(int64(*val), 10)
	return encodeCursor(&str)
}

// decodeIntCursor decodes the given integer cursor value. It is assumed to be a value
// previously returned from the function encodeIntCursor. The zero value is returned if
// no cursor is supplied. Invalid cursors return errors.
func decodeIntCursor(val *string) (int, error) {
	cursor, err := decodeCursor(val)
	if err != nil || cursor == "" {
		return 0, err
	}

	return strconv.Atoi(cursor)
}
