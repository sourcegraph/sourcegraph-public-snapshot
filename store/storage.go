package store

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

type Storage interface {
	// Get gets the value of a storage object.
	Get(context.Context, *sourcegraph.StorageKey) (*sourcegraph.StorageValue, error)

	// Put puts a value into a storage object.
	Put(context.Context, *sourcegraph.StoragePutOp) (*pbtypes.Void, error)

	// Delete deletes the specific storage object or, if no key is specified, all
	// objects in the bucket.
	Delete(context.Context, *sourcegraph.StorageKey) (*pbtypes.Void, error)

	// Exists tells if the given key exists in the bucket or not.
	Exists(context.Context, *sourcegraph.StorageKey) (*sourcegraph.StorageExists, error)

	// List lists all objects in the bucket. It ignores the 'key' field of the
	// storage name parameter.
	List(context.Context, *sourcegraph.StorageKey) (*sourcegraph.StorageList, error)
}
