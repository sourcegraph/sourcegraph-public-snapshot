package graphql

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

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

func unmarshalLSIFIndexGQLID(id graphql.ID) (indexID int64, err error) {
	err = relay.UnmarshalSpec(id, &indexID)
	return indexID, err
}

// ConnectionArgs is the common set of arguments to GraphQL fields that return connections (lists).
type ConnectionArgs struct {
	First *int32 // return the first n items
}

const DefaultIndexPageSize = 50

// makeGetIndexesOptions translates the given GraphQL arguments into options defined by the
// store.GetIndexes operations.
func makeGetIndexesOptions(args *resolverstubs.LSIFRepositoryIndexesQueryArgs) (shared.GetIndexesOptions, error) {
	repositoryID, err := resolveRepositoryID(args.RepositoryID)
	if err != nil {
		return shared.GetIndexesOptions{}, err
	}

	offset, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return shared.GetIndexesOptions{}, err
	}

	return shared.GetIndexesOptions{
		RepositoryID: repositoryID,
		State:        strings.ToLower(derefString(args.State, "")),
		Term:         derefString(args.Query, ""),
		Limit:        derefInt32(args.First, DefaultIndexPageSize),
		Offset:       offset,
	}, nil
}

// resolveRepositoryByID gets a repository's internal identifier from a GraphQL identifier.
func resolveRepositoryID(id graphql.ID) (int, error) {
	if id == "" {
		return 0, nil
	}

	repoID, err := UnmarshalRepositoryID(id)
	if err != nil {
		return 0, err
	}

	return int(repoID), nil
}

func UnmarshalRepositoryID(id graphql.ID) (repo api.RepoID, err error) {
	err = relay.UnmarshalSpec(id, &repo)
	return
}

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

// DecodeIntCursor decodes the given integer cursor value. It is assumed to be a value
// previously returned from the function encodeIntCursor. The zero value is returned if
// no cursor is supplied. Invalid cursors return errors.
func DecodeIntCursor(val *string) (int, error) {
	cursor, err := DecodeCursor(val)
	if err != nil || cursor == "" {
		return 0, err
	}

	return strconv.Atoi(cursor)
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

// toInt32 translates the given int pointer into an int32 pointer.
func toInt32(val *int) *int32 {
	if val == nil {
		return nil
	}

	v := int32(*val)
	return &v
}

// derefString returns the underlying value in the given pointer.
// If the pointer is nil, the default value is returned.
func derefString(val *string, defaultValue string) string {
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

// intPtr creates a pointer to the given value.
func intPtr(val int32) *int32 {
	return &val
}

// PageInfo implements the GraphQL type PageInfo.
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

// makeDeleteIndexesOptions translates the given GraphQL arguments into options defined by the
// store.DeleteIndexes operations.
func makeDeleteIndexesOptions(args *resolverstubs.DeleteLSIFIndexesArgs) (shared.DeleteIndexesOptions, error) {
	var repository int
	if args.Repository != nil {
		var err error
		repository, err = resolveRepositoryID(*args.Repository)
		if err != nil {
			return shared.DeleteIndexesOptions{}, err
		}
	}

	return shared.DeleteIndexesOptions{
		States:       []string{strings.ToLower(derefString(args.State, ""))},
		Term:         derefString(args.Query, ""),
		RepositoryID: repository,
	}, nil
}

// makeReindexIndexesOptions translates the given GraphQL arguments into options defined by the
// store.ReindexIndexes operations.
func makeReindexIndexesOptions(args *resolverstubs.ReindexLSIFIndexesArgs) (shared.ReindexIndexesOptions, error) {
	var repository int
	if args.Repository != nil {
		var err error
		repository, err = resolveRepositoryID(*args.Repository)
		if err != nil {
			return shared.ReindexIndexesOptions{}, err
		}
	}

	return shared.ReindexIndexesOptions{
		States:       []string{strings.ToLower(derefString(args.State, ""))},
		Term:         derefString(args.Query, ""),
		RepositoryID: repository,
	}, nil
}
