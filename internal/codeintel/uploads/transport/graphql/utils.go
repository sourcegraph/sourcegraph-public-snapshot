package graphql

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
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

func unmarshalLSIFUploadGQLID(id graphql.ID) (uploadID int64, err error) {
	// First, try to unmarshal the ID as a string and then convert it to an
	// integer. This is here to maintain backwards compatibility with the
	// src-cli lsif upload command, which constructs its own relay identifier
	// from a the string payload returned by the upload proxy.
	var idString string
	err = relay.UnmarshalSpec(id, &idString)
	if err == nil {
		uploadID, err = strconv.ParseInt(idString, 10, 64)
		return
	}

	// If it wasn't unmarshal-able as a string, it's a new-style int identifier
	err = relay.UnmarshalSpec(id, &uploadID)
	return uploadID, err
}

func marshalLSIFUploadGQLID(uploadID int64) graphql.ID {
	return relay.MarshalID("LSIFUpload", uploadID)
}

func unmarshalRepositoryID(id graphql.ID) (repo api.RepoID, err error) {
	err = relay.UnmarshalSpec(id, &repo)
	return
}

const DefaultUploadPageSize = 50

// makeGetUploadsOptions translates the given GraphQL arguments into options defined by the
// store.GetUploads operations.
func makeGetUploadsOptions(args *LSIFRepositoryUploadsQueryArgs) (types.GetUploadsOptions, error) {
	repositoryID, err := resolveRepositoryID(args.RepositoryID)
	if err != nil {
		return types.GetUploadsOptions{}, err
	}

	var dependencyOf int64
	if args.DependencyOf != nil {
		dependencyOf, err = unmarshalLSIFUploadGQLID(*args.DependencyOf)
		if err != nil {
			return types.GetUploadsOptions{}, err
		}
	}

	var dependentOf int64
	if args.DependentOf != nil {
		dependentOf, err = unmarshalLSIFUploadGQLID(*args.DependentOf)
		if err != nil {
			return types.GetUploadsOptions{}, err
		}
	}

	offset, err := decodeIntCursor(args.After)
	if err != nil {
		return types.GetUploadsOptions{}, err
	}

	return types.GetUploadsOptions{
		RepositoryID:       repositoryID,
		State:              strings.ToLower(derefString(args.State, "")),
		Term:               derefString(args.Query, ""),
		VisibleAtTip:       derefBool(args.IsLatestForRepo, false),
		DependencyOf:       int(dependencyOf),
		DependentOf:        int(dependentOf),
		Limit:              derefInt32(args.First, DefaultUploadPageSize),
		Offset:             offset,
		AllowExpired:       true,
		AllowDeletedUpload: derefBool(args.IncludeDeleted, false),
	}, nil
}

// makeDeleteUploadsOptions translates the given GraphQL arguments into options defined by the
// store.DeleteUploads operations.
func makeDeleteUploadsOptions(args *DeleteLSIFUploadsArgs) types.DeleteUploadsOptions {
	return types.DeleteUploadsOptions{
		State:        strings.ToLower(derefString(args.State, "")),
		Term:         derefString(args.Query, ""),
		VisibleAtTip: derefBool(args.IsLatestForRepo, false),
	}
}

// resolveRepositoryByID gets a repository's internal identifier from a GraphQL identifier.
func resolveRepositoryID(id graphql.ID) (int, error) {
	if id == "" {
		return 0, nil
	}

	repoID, err := unmarshalRepositoryID(id)
	if err != nil {
		return 0, err
	}

	return int(repoID), nil
}

// derefString returns the underlying value in the given pointer.
// If the pointer is nil, the default value is returned.
func derefString(val *string, defaultValue string) string {
	if val != nil {
		return *val
	}
	return defaultValue
}

// derefBool returns the underlying value in the given pointer.
// If the pointer is nil, the default value is returned.
func derefBool(val *bool, defaultValue bool) bool {
	if val != nil {
		return *val
	}
	return defaultValue
}

// derefInt32 returns the underlying value in the given pointer.
// If the pointer is nil, the default value is returned.
func derefInt32(val *int32, defaultValue int) int {
	if val != nil {
		return int(*val)
	}
	return defaultValue
}

// ConnectionArgs is the common set of arguments to GraphQL fields that return connections (lists).
type ConnectionArgs struct {
	First *int32 // return the first n items
}

// Set is a convenience method for setting the DB limit and offset in a DB XyzListOptions struct.
func (a ConnectionArgs) Set(o **database.LimitOffset) {
	if a.First != nil {
		*o = &database.LimitOffset{Limit: int(*a.First)}
	}
}

// GetFirst is a convenience method returning the value of First, defaulting to
// the type's zero value if nil.
func (a ConnectionArgs) GetFirst() int32 {
	if a.First == nil {
		return 0
	}
	return *a.First
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

// EncodeCursor creates a PageInfo object from the given cursor. If the cursor is not
// defined, then an object indicating the end of the result set is returned. The cursor
// is base64 encoded for transfer, and should be decoded using the function decodeCursor.
func EncodeCursor(val *string) *PageInfo {
	if val != nil {
		return NextPageCursor(base64.StdEncoding.EncodeToString([]byte(*val)))
	}

	return HasNextPage(false)
}

// DecodeCursor decodes the given cursor value. It is assumed to be a value previously
// returned from the function encodeCursor. An empty string is returned if no cursor is
// supplied. Invalid cursors return errors.
func DecodeCursor(val *string) (string, error) {
	if val == nil {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*val)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// DecodeIntCursor decodes the given integer cursor value. It is assumed to be a value
// previously returned from the function encodeIntCursor. The zero value is returned if
// no cursor is supplied. Invalid cursors return errors.
func decodeIntCursor(val *string) (int, error) {
	cursor, err := DecodeCursor(val)
	if err != nil || cursor == "" {
		return 0, err
	}

	return strconv.Atoi(cursor)
}

// strPtr creates a pointer to the given value. If the value is an
// empty string, a nil pointer is returned.
func strPtr(val string) *string {
	if val == "" {
		return nil
	}

	return &val
}

// intPtr creates a pointer to the given value.
func intPtr(val int32) *int32 {
	return &val
}

// intPtr creates a pointer to the given value.
func boolPtr(val bool) *bool {
	return &val
}
