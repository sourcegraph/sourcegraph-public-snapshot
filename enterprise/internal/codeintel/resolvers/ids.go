package resolvers

import (
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

func marshalLSIFUploadGQLID(lsifUploadID int64) graphql.ID {
	return relay.MarshalID("LSIFUpload", lsifUploadID)
}

func unmarshalLSIFUploadGQLID(id graphql.ID) (lsifUploadID int64, err error) {
	// First, try to unmarshal the ID as a string and then convert it to an
	// integer. This is here to maintain backwards compatibility with the
	// src-cli lsif upload command, which constructs its own relay identifier
	// from a the string payload returned by the upload proxy.

	var lsifUploadIDString string
	err = relay.UnmarshalSpec(id, &lsifUploadIDString)
	if err == nil {
		lsifUploadID, err = strconv.ParseInt(lsifUploadIDString, 10, 64)
		return
	}

	// If it wasn't unmarshal-able as a string, it's a new-style int identifier
	err = relay.UnmarshalSpec(id, &lsifUploadID)
	return lsifUploadID, err
}

func marshalLSIFIndexGQLID(lsifIndexID int64) graphql.ID {
	return relay.MarshalID("LSIFIndex", lsifIndexID)
}

func unmarshalLSIFIndexGQLID(id graphql.ID) (lsifIndexID int64, err error) {
	err = relay.UnmarshalSpec(id, &lsifIndexID)
	return lsifIndexID, err
}
