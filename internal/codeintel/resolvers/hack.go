package resolvers

import (
	"encoding/base64"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

// TODO: This file should not exist but it breaks the dependency of codeintel resolvers on the cmd/frontend
// package. This should be removed once codeintel resolvers are moved into cmd/frontend.

// TODO: Duplicated. Can be removed once the resolvers for code intel live in the
// cmd/frontend package.
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

// EncodeCursor creates a PageInfo object from the given cursor. If the cursor is not
// defined, then an object indicating the end of the result set is returned. The cursor
// is base64 encoded for transfer, and should be decoded using the function decodeCursor.
func EncodeCursor(val *string) PageInfo {
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

// EncodeIntCursor creates a PageInfo object from the given new offset value. If the
// new offset value, then an object indicating the end of the result set is returned.
// The cursor is base64 encoded for transfer, and should be decoded using the function
// decodeIntCursor.
func EncodeIntCursor(val *int32) PageInfo {
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

// NextOffset determines the offset that should be used for a subsequent request.
// If there are no more results in the paged result set, this function returns nil.
func NextOffset(offset, count, totalCount int) *int32 {
	if offset+count < totalCount {
		val := int32(offset + count)
		return &val
	}

	return nil
}

// pageInfo implements the GraphQL type PageInfo.
type pageInfo struct {
	endCursor   *string
	hasNextPage bool
}

// HasNextPage returns a new PageInfo with the given hasNextPage value.
func HasNextPage(hasNextPage bool) *pageInfo {
	return &pageInfo{hasNextPage: hasNextPage}
}

// NextPageCursor returns a new PageInfo indicating there is a next page with
// the given end cursor.
func NextPageCursor(endCursor string) *pageInfo {
	return &pageInfo{endCursor: &endCursor, hasNextPage: true}
}

func (r *pageInfo) EndCursor() *string { return r.endCursor }
func (r *pageInfo) HasNextPage() bool  { return r.hasNextPage }
