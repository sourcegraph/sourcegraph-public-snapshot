package resolvers

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

type ConnectionResolver[T any] interface {
	Nodes(ctx context.Context) ([]T, error)
}

type PagedConnectionResolver[T any] interface {
	ConnectionResolver[T]
	PageInfo() PageInfo
}

type PagedConnectionResolverWithCount[T any] interface {
	PagedConnectionResolver[T]
	TotalCount() *int32
}

type PageInfo interface {
	HasNextPage() bool
	EndCursor() *string
}

type ConnectionArgs struct {
	First *int32
}

type PagedConnectionArgs struct {
	ConnectionArgs
	After *string
}

type EmptyResponse struct{}

func (er *EmptyResponse) AlwaysNil() *string {
	return nil
}

func UnmarshalLSIFUploadGQLID(id graphql.ID) (uploadID int64, err error) {
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
