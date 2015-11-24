package storage

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// System represents the storage system. It is a simple key / value store.
//
// Conceptually, buckets keys are similiar to directories, keys are similiar to
// files, and values are similiar to the contents of a file.
type System interface {
	// Get reads the value of <bucket>/<key>.
	//
	// If the specified object does not exist, os.ErrNotExist is returned.
	Get(bucket, key string) ([]byte, error)

	// Put puts the given value into <bucket>/<key>.
	Put(bucket, key string, value []byte) error

	// PutNoOverwrite puts the given value into <bucket>/<key> and returns
	// os.ErrExist if the bucket already has such a key.
	PutNoOverwrite(bucket, key string, value []byte) error

	// Delete deletes the entire <bucket> or deletes <bucket>/<key> if key != "".
	//
	// If the given key or bucket does not exist, this function is no-op.
	Delete(bucket, key string) error

	// Exists tells if <bucket/key> exists or not.
	//
	// If the object does not exist, no error is returned, only exists=false is
	// returned.
	Exists(bucket, key string) (bool, error)

	// List lists all objects in <bucket>.
	//
	// If the bucket does not exist, no error is returned, only an empty list is
	// returned.
	List(bucket string) ([]string, error)

	String() string
}

// Namespace returns a storage system for the given application. Because
// creating a new storage system via this function is very lightweight, it is
// acceptable to create one for each request.
//
// appName is the name of the application whose data you are trying to
// read/write, applications may read and write to eachother's buckets assuming
// the admin has not restricted such access.
//
// If the repo is a valid repo URI, storage is considered "local" to that
// repository. Otherwise, storage is considered "global" (i.e. shared across
// all repositories).
func Namespace(ctx context.Context, appName string, repo string) System {
	return &system{
		ctx:     ctx,
		client:  sourcegraph.NewClientFromContext(ctx),
		appName: appName,
		repo:    repo,
	}
}
