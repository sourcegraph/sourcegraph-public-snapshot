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
	Get(bucket, key string) ([]byte, error)

	// Put puts the given value into <bucket>/<key>.
	Put(bucket, key string, value []byte) error

	// Delete deletes the entire <bucket> or deletes <bucket>/<key> if key != "".
	Delete(bucket, key string) error

	// Exists tells if <bucket/key> exists or not.
	Exists(bucket, key string) bool

	// List lists all objects in <bucket>.
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
