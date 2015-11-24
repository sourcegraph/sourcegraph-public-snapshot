package storage

import (
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// system implements the System interface by wrapping the gRPC Storage service.
type system struct {
	ctx     context.Context
	client  *sourcegraph.Client
	appName string
	repo    string
}

// String implements the fmt.Stringer interface.
func (s *system) String() string {
	return fmt.Sprintf("storage.System(AppName=%q, Repo=%q)", s.appName, s.repo)
}

// storageKey is a utility function which returns a new StorageKey given it's
// bucket and key strings. The AppName and Repo fields are handled for you.
func (s *system) storageKey(bucket, key string) *sourcegraph.StorageKey {
	return &sourcegraph.StorageKey{
		Bucket: &sourcegraph.StorageBucket{
			AppName: s.appName,
			Repo:    s.repo,
			Name:    bucket,
		},
		Key: key,
	}
}

// Get implements the Storage interface.
func (s *system) Get(bucket, key string) ([]byte, error) {
	v, err := s.client.Storage.Get(s.ctx, s.storageKey(bucket, key))
	if err != nil {
		// If the specified object does not exist, os.ErrNotExist is returned.
		if grpc.Code(err) == codes.NotFound {
			err = os.ErrNotExist
		}
		return nil, err
	}
	return v.Value, nil
}

// Put implements the Storage interface.
func (s *system) Put(bucket, key string, value []byte) error {
	_, err := s.client.Storage.Put(s.ctx, &sourcegraph.StoragePutOp{
		Key:   *s.storageKey(bucket, key),
		Value: value,
	})
	return err
}

// PutNoOverwrite implements the Storage interface.
func (s *system) PutNoOverwrite(bucket, key string, value []byte) error {
	_, err := s.client.Storage.PutNoOverwrite(s.ctx, &sourcegraph.StoragePutOp{
		Key:   *s.storageKey(bucket, key),
		Value: value,
	})
	if err != nil && grpc.Code(err) == codes.AlreadyExists {
		return os.ErrExist
	}
	return err
}

// Delete implements the Storage interface.
func (s *system) Delete(bucket, key string) error {
	_, err := s.client.Storage.Delete(s.ctx, s.storageKey(bucket, key))
	return err
}

// Exists implements the Storage interface.
func (s *system) Exists(bucket, key string) bool {
	exists, err := s.client.Storage.Exists(s.ctx, s.storageKey(bucket, key))
	if err != nil {
		return false
	}
	return exists.Exists
}

// List implements the Storage interface.
func (s *system) List(bucket string) ([]string, error) {
	list, err := s.client.Storage.List(s.ctx, s.storageKey(bucket, ""))
	if err != nil {
		return nil, err
	}
	return list.Keys, nil
}
