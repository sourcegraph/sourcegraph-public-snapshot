package uploadstore

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type diskStore struct {
	config Config
}

type DiskConfig struct {
	BaseFolder string
}

var _ Store = &gcsStore{}

// newDiskStoreFromConfig creates a new store backed by a folder on disk.
func newDiskStoreFromConfig(_ context.Context, config Config, _ *Operations) (Store, error) {
	return &diskStore{config: config}, nil
}

func (s *diskStore) Init(ctx context.Context) error {
	return nil
}

func (s *diskStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	p := s.pathForKey(key)
	return os.Open(p)
}

func (s *diskStore) Upload(ctx context.Context, key string, r io.Reader) (int64, error) {
	p := s.pathForKey(key)

	// TODO: should this be Create or Open?
	f, err := os.Create(p)
	if err != nil {
		return 0, err
	}

	return io.Copy(f, r)
}

func (s *diskStore) Compose(ctx context.Context, destination string, sources ...string) (int64, error) {
	destPath := s.pathForKey(destination)

	dest, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}

	var count int64
	for _, src := range sources {
		f, err := s.Get(ctx, src)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to open %s", src)
		}

		n, err := io.Copy(dest, f)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to copy %s to %s", src, destination)
		}
		count += n
	}

	return count, nil
}

func (s *diskStore) Delete(ctx context.Context, key string) error {
	p := s.pathForKey(key)
	return os.Remove(p)
}

func (s *diskStore) pathForKey(key string) string {
	return filepath.Join(s.config.Disk.BaseFolder, key)
}
