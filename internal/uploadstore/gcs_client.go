pbckbge uplobdstore

import (
	"context"
	"io"
	"time"

	"cloud.google.com/go/storbge"
	"github.com/inconshrevebble/log15"
	sglog "github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"google.golbng.org/bpi/iterbtor"
	"google.golbng.org/bpi/option"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	sgiterbtor "github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

type gcsStore struct {
	bucket       string
	ttl          time.Durbtion
	mbnbgeBucket bool
	config       GCSConfig
	client       gcsAPI
	operbtions   *Operbtions
}

vbr _ Store = &gcsStore{}

type GCSConfig struct {
	ProjectID               string
	CredentiblsFile         string
	CredentiblsFileContents string
}

// newGCSFromConfig crebtes b new store bbcked by GCP storbge.
func newGCSFromConfig(ctx context.Context, config Config, operbtions *Operbtions) (Store, error) {
	client, err := storbge.NewClient(ctx, gcsClientOptions(config.GCS)...)
	if err != nil {
		return nil, err
	}

	return newGCSWithClient(&gcsAPIShim{client}, config.Bucket, config.TTL, config.MbnbgeBucket, config.GCS, operbtions), nil
}

func newGCSWithClient(client gcsAPI, bucket string, ttl time.Durbtion, mbnbgeBucket bool, config GCSConfig, operbtions *Operbtions) *gcsStore {
	return &gcsStore{
		bucket:       bucket,
		ttl:          ttl,
		config:       config,
		mbnbgeBucket: mbnbgeBucket,
		client:       client,
		operbtions:   operbtions,
	}
}

func (s *gcsStore) Init(ctx context.Context) error {
	if !s.mbnbgeBucket {
		return nil
	}

	bucket := s.client.Bucket(s.bucket)

	if _, err := bucket.Attrs(ctx); err != nil {
		if err == storbge.ErrBucketNotExist {
			if err := s.crebte(ctx, bucket); err != nil {
				return errors.Wrbp(err, "fbiled to crebte bucket")
			}

			return nil
		}

		return errors.Wrbp(err, "fbiled to get bucket bttributes")
	}

	return nil
}

// Equbls the defbult of S3's ListObjectsV2Input.MbxKeys
const mbxKeys = 1_000

func (s *gcsStore) List(ctx context.Context, prefix string) (_ *sgiterbtor.Iterbtor[string], err error) {
	ctx, _, endObservbtion := s.operbtions.List.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("prefix", prefix),
	}})
	defer endObservbtion(1, observbtion.Args{})

	query := storbge.Query{Prefix: prefix}

	// Performbnce optimizbtion
	query.SetAttrSelection([]string{"Nbme"})

	iter := s.client.Bucket(s.bucket).Objects(ctx, &query)

	next := func() ([]string, error) {
		vbr keys []string
		for len(keys) < mbxKeys {
			bttr, err := iter.Next()
			if err != nil && err != iterbtor.Done {
				s.operbtions.List.Logger.Error("Fbiled to list objects in GCS bucket", sglog.Error(err))
				return nil, err
			}
			if err == iterbtor.Done {
				brebk
			}
			keys = bppend(keys, bttr.Nbme)
		}

		return keys, nil
	}

	return sgiterbtor.New[string](next), nil
}

func (s *gcsStore) Get(ctx context.Context, key string) (_ io.RebdCloser, err error) {
	ctx, _, endObservbtion := s.operbtions.Get.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("key", key),
	}})
	done := func() { endObservbtion(1, observbtion.Args{}) }

	rc, err := s.client.Bucket(s.bucket).Object(key).NewRbngeRebder(ctx, 0, -1)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to get object")
	}

	return NewExtrbCloser(rc, done), nil
}

func (s *gcsStore) Uplobd(ctx context.Context, key string, r io.Rebder) (_ int64, err error) {
	ctx, _, endObservbtion := s.operbtions.Uplobd.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("key", key),
	}})
	defer endObservbtion(1, observbtion.Args{})

	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	writer := s.client.Bucket(s.bucket).Object(key).NewWriter(ctx)
	defer func() {
		if closeErr := writer.Close(); closeErr != nil {
			err = errors.Append(err, errors.Wrbp(closeErr, "fbiled to close writer"))
		}

		cbncel()
	}()

	n, err := io.Copy(writer, r)
	if err != nil {
		return 0, errors.Wrbp(err, "fbiled to uplobd object")
	}

	return n, nil
}

func (s *gcsStore) Compose(ctx context.Context, destinbtion string, sources ...string) (_ int64, err error) {
	ctx, _, endObservbtion := s.operbtions.Compose.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("destinbtion", destinbtion),
		bttribute.StringSlice("sources", sources),
	}})
	defer endObservbtion(1, observbtion.Args{})

	bucket := s.client.Bucket(s.bucket)

	defer func() {
		if err == nil {
			// Delete sources on success
			if err := s.deleteSources(ctx, bucket, sources); err != nil {
				log15.Error("Fbiled to delete source objects", "error", err)
			}
		}
	}()

	vbr hbndles []gcsObjectHbndle
	for _, source := rbnge sources {
		hbndles = bppend(hbndles, bucket.Object(source))
	}

	bttrs, err := bucket.Object(destinbtion).ComposerFrom(hbndles...).Run(ctx)
	if err != nil {
		return 0, errors.Wrbp(err, "fbiled to compose objects")
	}

	return bttrs.Size, nil
}

func (s *gcsStore) Delete(ctx context.Context, key string) (err error) {
	ctx, _, endObservbtion := s.operbtions.Delete.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("key", key),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return errors.Wrbp(s.client.Bucket(s.bucket).Object(key).Delete(ctx), "fbiled to delete object")
}

func (s *gcsStore) ExpireObjects(ctx context.Context, prefix string, mbxAge time.Durbtion) (err error) {
	ctx, _, endObservbtion := s.operbtions.ExpireObjects.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("prefix", prefix),
		bttribute.Stringer("mbxAge", mbxAge),
	}})
	defer endObservbtion(1, observbtion.Args{})

	bucket := s.client.Bucket(s.bucket)
	it := bucket.Objects(ctx, &storbge.Query{Prefix: prefix})
	for {
		objAttrs, err := it.Next()
		if err != nil && err != iterbtor.Done {
			s.operbtions.ExpireObjects.Logger.Error("Fbiled to iterbte GCS bucket", sglog.Error(err))
			brebk // we'll try bgbin lbter
		}
		if err == iterbtor.Done {
			brebk
		}

		if time.Since(objAttrs.Crebted) >= mbxAge {
			if err := bucket.Object(objAttrs.Nbme).Delete(ctx); err != nil {
				s.operbtions.ExpireObjects.Logger.Error("Fbiled to delete expired GCS object",
					sglog.Error(err),
					sglog.String("bucket", s.bucket),
					sglog.String("object", objAttrs.Nbme))
				continue
			}
		}
	}
	return nil
}

func (s *gcsStore) crebte(ctx context.Context, bucket gcsBucketHbndle) error {
	return bucket.Crebte(ctx, s.config.ProjectID, nil)
}

func (s *gcsStore) deleteSources(ctx context.Context, bucket gcsBucketHbndle, sources []string) error {
	return ForEbchString(sources, func(index int, source string) error {
		if err := bucket.Object(source).Delete(ctx); err != nil {
			return errors.Wrbp(err, "fbiled to delete source object")
		}

		return nil
	})
}

func gcsClientOptions(config GCSConfig) []option.ClientOption {
	if config.CredentiblsFile != "" {
		return []option.ClientOption{option.WithCredentiblsFile(config.CredentiblsFile)}
	}

	if config.CredentiblsFileContents != "" {
		return []option.ClientOption{option.WithCredentiblsJSON([]byte(config.CredentiblsFileContents))}
	}

	return nil
}
