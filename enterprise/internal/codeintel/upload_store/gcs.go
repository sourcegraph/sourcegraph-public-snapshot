package uploadstore

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/api/option"
)

type gcsStore struct {
	projectID string
	bucket    string
	ttl       time.Duration
	client    *storage.Client
}

var _ Store = &gcsStore{}

// newGCSFromConfig creates a new store backed by GCP storage.
func newGCSFromConfig(ctx context.Context, config *Config) (Store, error) {
	return newGCS(ctx, config.GCS.ProjectID, config.GCS.Bucket, config.GCS.TTL)
}

// newGCS creates a new store backed by GCP storage.
func newGCS(ctx context.Context, projectID, bucket string, ttl time.Duration) (Store, error) {
	client, err := storage.NewClient(ctx, gcsClientOptions()...)
	if err != nil {
		return nil, err
	}

	return &gcsStore{
		projectID: projectID,
		bucket:    bucket,
		ttl:       ttl,
		client:    client,
	}, nil
}

func (s *gcsStore) Init(ctx context.Context) error {
	if _, err := s.client.Bucket(s.bucket).Attrs(ctx); err != nil {
		if err != storage.ErrBucketNotExist {
			return err
		}

		return s.create(ctx)
	}

	return s.update(ctx)
}

func (s *gcsStore) Get(ctx context.Context, key string, skipBytes int64) (io.ReadCloser, error) {
	return s.client.Bucket(s.bucket).Object(key).NewRangeReader(ctx, skipBytes, -1)
}

func (s *gcsStore) Upload(ctx context.Context, key string, r io.Reader) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	writer := s.client.Bucket(s.bucket).Object(key).NewWriter(ctx)
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			err = multierror.Append(err, closeErr)
			fmt.Printf("B %s\n", err)
		}

		cancel()
	}()

	_, err = io.Copy(writer, r)
	fmt.Printf("A %s\n", err)
	return err
}

func (s *gcsStore) create(ctx context.Context) error {
	return s.client.Bucket(s.bucket).Create(ctx, s.projectID, &storage.BucketAttrs{
		Lifecycle: s.lifecycle(),
	})
}

func (s *gcsStore) update(ctx context.Context) error {
	lifecycle := s.lifecycle()
	_, err := s.client.Bucket(s.bucket).Update(ctx, storage.BucketAttrsToUpdate{
		Lifecycle: &lifecycle,
	})
	return err
}

func (s *gcsStore) lifecycle() storage.Lifecycle {
	return storage.Lifecycle{
		Rules: []storage.LifecycleRule{
			{
				Action: storage.LifecycleAction{
					Type: "Delete",
				},
				Condition: storage.LifecycleCondition{
					AgeInDays: int64(time.Duration(s.ttl) / (time.Hour * 24)),
				},
			},
		},
	}
}

// gcsClientOptions returns options used to configure a GCS storage client. If the
// envvar GOOGLE_APPLICATION_CREDENTIALS is set, it will be used as a path to the GCP
// credentials file. If the envvar GOOGLE_APPLICATION_CREDENTIALS_JSON is set, it will
// be used as the JSON payload of a GCP credentials file.
//
// See https://cloud.google.com/docs/authentication/production.
func gcsClientOptions() []option.ClientOption {
	if value := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); value != "" {
		return []option.ClientOption{option.WithCredentialsFile(value)}
	}

	if value := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON"); value != "" {
		return []option.ClientOption{option.WithCredentialsJSON([]byte(value))}
	}

	return nil
}
