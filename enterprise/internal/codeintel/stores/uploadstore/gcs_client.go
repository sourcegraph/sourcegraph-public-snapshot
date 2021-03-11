package uploadstore

import (
	"context"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type gcsStore struct {
	bucket       string
	ttl          time.Duration
	manageBucket bool
	config       GCSConfig
	client       gcsAPI
	operations   *operations
}

var _ Store = &gcsStore{}

type GCSConfig struct {
	ProjectID               string
	CredentialsFile         string
	CredentialsFileContents string
}

func (c *GCSConfig) load(parent *env.BaseConfig) {
	c.ProjectID = parent.Get("PRECISE_CODE_INTEL_UPLOAD_GCP_PROJECT_ID", "", "The project containing the GCS bucket.")
	c.CredentialsFile = parent.GetOptional("PRECISE_CODE_INTEL_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE", "The path to a service account key file with access to GCS.")
	c.CredentialsFileContents = parent.GetOptional("PRECISE_CODE_INTEL_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT", "The contents of a service account key file with access to GCS.")
}

// newGCSFromConfig creates a new store backed by GCP storage.
func newGCSFromConfig(ctx context.Context, config *Config, operations *operations) (Store, error) {
	client, err := storage.NewClient(ctx, gcsClientOptions(config.GCS)...)
	if err != nil {
		return nil, err
	}

	return newGCSWithClient(&gcsAPIShim{client}, config.Bucket, config.TTL, config.ManageBucket, config.GCS, operations), nil
}

func newGCSWithClient(client gcsAPI, bucket string, ttl time.Duration, manageBucket bool, config GCSConfig, operations *operations) *gcsStore {
	return &gcsStore{
		bucket:       bucket,
		ttl:          ttl,
		config:       config,
		manageBucket: manageBucket,
		client:       client,
		operations:   operations,
	}
}

func (s *gcsStore) Init(ctx context.Context) error {
	if !s.manageBucket {
		return nil
	}

	bucket := s.client.Bucket(s.bucket)

	if _, err := bucket.Attrs(ctx); err != nil {
		if err == storage.ErrBucketNotExist {
			if err := s.create(ctx, bucket); err != nil {
				return errors.Wrap(err, "failed to create bucket")
			}

			return nil
		}

		return errors.Wrap(err, "failed to get bucket attributes")
	}

	if err := s.update(ctx, bucket); err != nil {
		return errors.Wrap(err, "failed to update bucket attributes")
	}

	return nil
}

func (s *gcsStore) Get(ctx context.Context, key string) (_ io.ReadCloser, err error) {
	ctx, endObservation := s.operations.get.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("key", key),
	}})
	defer endObservation(1, observation.Args{})

	rc, err := s.client.Bucket(s.bucket).Object(key).NewRangeReader(ctx, 0, -1)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get object")
	}

	return rc, nil
}

func (s *gcsStore) Upload(ctx context.Context, key string, r io.Reader) (_ int64, err error) {
	ctx, endObservation := s.operations.upload.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("key", key),
	}})
	defer endObservation(1, observation.Args{})

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	writer := s.client.Bucket(s.bucket).Object(key).NewWriter(ctx)
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			err = multierror.Append(err, errors.Wrap(closeErr, "failed to close writer"))
		}

		cancel()
	}()

	n, err := io.Copy(writer, r)
	if err != nil {
		return 0, errors.Wrap(err, "failed to upload object")
	}

	return n, nil
}

func (s *gcsStore) Compose(ctx context.Context, destination string, sources ...string) (_ int64, err error) {
	ctx, endObservation := s.operations.compose.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("destination", destination),
		log.String("sources", strings.Join(sources, ", ")),
	}})
	defer endObservation(1, observation.Args{})

	bucket := s.client.Bucket(s.bucket)

	defer func() {
		if err == nil {
			// Delete sources on success
			if err := s.deleteSources(ctx, bucket, sources); err != nil {
				log15.Error("Failed to delete source objects", "error", err)
			}
		}
	}()

	var handles []gcsObjectHandle
	for _, source := range sources {
		handles = append(handles, bucket.Object(source))
	}

	attrs, err := bucket.Object(destination).ComposerFrom(handles...).Run(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to compose objects")
	}

	return attrs.Size, nil
}

func (s *gcsStore) Delete(ctx context.Context, key string) (err error) {
	ctx, endObservation := s.operations.delete.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("key", key),
	}})
	defer endObservation(1, observation.Args{})

	return errors.Wrap(s.client.Bucket(s.bucket).Object(key).Delete(ctx), "failed to delete object")
}

func (s *gcsStore) create(ctx context.Context, bucket gcsBucketHandle) error {
	return bucket.Create(ctx, s.config.ProjectID, &storage.BucketAttrs{
		Lifecycle: s.lifecycle(),
	})
}

func (s *gcsStore) update(ctx context.Context, bucket gcsBucketHandle) error {
	lifecycle := s.lifecycle()

	return bucket.Update(ctx, storage.BucketAttrsToUpdate{
		Lifecycle: &lifecycle,
	})
}

func (s *gcsStore) lifecycle() storage.Lifecycle {
	return storage.Lifecycle{
		Rules: []storage.LifecycleRule{
			{
				Action: storage.LifecycleAction{
					Type: "Delete",
				},
				Condition: storage.LifecycleCondition{
					AgeInDays: int64(s.ttl / (time.Hour * 24)),
				},
			},
		},
	}
}

func (s *gcsStore) deleteSources(ctx context.Context, bucket gcsBucketHandle, sources []string) error {
	return goroutine.RunWorkersOverStrings(sources, func(index int, source string) error {
		if err := bucket.Object(source).Delete(ctx); err != nil {
			return errors.Wrap(err, "failed to delete source object")
		}

		return nil
	})
}

func gcsClientOptions(config GCSConfig) []option.ClientOption {
	if config.CredentialsFile != "" {
		return []option.ClientOption{option.WithCredentialsFile(config.CredentialsFile)}
	}

	if config.CredentialsFileContents != "" {
		return []option.ClientOption{option.WithCredentialsJSON([]byte(config.CredentialsFileContents))}
	}

	return nil
}
