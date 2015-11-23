package store

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

type Storage interface {
	// Get gets the value of a storage object.
	//
	// If the specified object does not exist, an error detectable with
	// os.IsNotExist is returned.
	Get(context.Context, *sourcegraph.StorageKey) (*sourcegraph.StorageValue, error)

	// Put puts a value into a storage object.
	Put(context.Context, *sourcegraph.StoragePutOp) (*pbtypes.Void, error)

	// PutNoOverwrite puts an object into storage, returning an error detectable
	// with os.IsNotExist if the object already exists.
	PutNoOverwrite(context.Context, *sourcegraph.StoragePutOp) (*pbtypes.Void, error)

	// Delete deletes the specific storage object or, if no key is specified, all
	// objects in the bucket.
	//
	// If the given key or bucket does not exist, this function is no-op.
	Delete(context.Context, *sourcegraph.StorageKey) (*pbtypes.Void, error)

	// Exists tells if the given key exists in the bucket or not.
	//
	// If the object does not exist, no error is returned, only exists=false is
	// returned.
	Exists(context.Context, *sourcegraph.StorageKey) (*sourcegraph.StorageExists, error)

	// List lists all objects in the bucket. It ignores the 'key' field of the
	// storage name parameter.
	//
	// If the bucket does not exist, no error is returned, only an empty list is
	// returned.
	List(context.Context, *sourcegraph.StorageKey) (*sourcegraph.StorageList, error)
}
