package uploadstore

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3Store struct {
	bucket       string
	ttl          time.Duration
	manageBucket bool
	client       *s3.S3
	uploader     *s3manager.Uploader
}

var _ Store = &s3Store{}

// newS3FromConfig creates a new store backed by AWS Simple Storage Service.
func newS3FromConfig(ctx context.Context, config *Config) (Store, error) {
	return newS3(config.S3.Bucket, config.S3.TTL, config.ManageBucket)
}

// newS3 creates a new store backed by AWS Simple Storage Service.
func newS3(bucket string, ttl time.Duration, manageBucket bool) (Store, error) {
	sess, err := session.NewSessionWithOptions(awsSessionOptions())
	if err != nil {
		return nil, err
	}

	client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)

	return &s3Store{
		bucket:       bucket,
		ttl:          ttl,
		manageBucket: manageBucket,
		client:       client,
		uploader:     uploader,
	}, nil
}

func (s *s3Store) Init(ctx context.Context) error {
	if !s.manageBucket {
		return nil
	}

	if err := s.create(ctx); err != nil {
		return err
	}

	return s.update(ctx)
}

func (s *s3Store) Get(ctx context.Context, key string, skipBytes int64) (io.ReadCloser, error) {
	var bytesRange *string
	if skipBytes > 0 {
		bytesRange = aws.String(fmt.Sprintf("bytes=%d-", skipBytes))
	}

	resp, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Range:  bytesRange,
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (s *s3Store) Upload(ctx context.Context, key string, r io.Reader) error {
	_, err := s.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	})

	return err
}

func (s *s3Store) create(ctx context.Context) error {
	createRequest := &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	}

	if _, err := s.client.CreateBucketWithContext(ctx, createRequest); err != nil {
		aerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}

		codes := []string{
			s3.ErrCodeBucketAlreadyExists,
			s3.ErrCodeBucketAlreadyOwnedByYou,
		}

		found := false
		for _, code := range codes {
			if aerr.Code() == code {
				found = true
			}
		}
		if !found {
			return err
		}
	}

	return nil
}

func (s *s3Store) update(ctx context.Context) error {
	configureRequest := &s3.PutBucketLifecycleConfigurationInput{
		Bucket:                 aws.String(s.bucket),
		LifecycleConfiguration: s.lifecycle(),
	}

	_, err := s.client.PutBucketLifecycleConfiguration(configureRequest)
	return err
}

func (s *s3Store) lifecycle() *s3.BucketLifecycleConfiguration {
	return &s3.BucketLifecycleConfiguration{
		Rules: []*s3.LifecycleRule{
			{
				ID: aws.String("Expiration Rule"),
				Filter: &s3.LifecycleRuleFilter{
					Prefix: aws.String(""),
				},
				Expiration: &s3.LifecycleExpiration{
					Days: aws.Int64(int64(time.Duration(s.ttl) / (time.Hour * 24))),
				},
				Status: aws.String("Enabled"),
			},
		},
	}
}

// awsSessionOptions returns the session used to configure the AWS SDK client.
//
// Authentication of the client will first prefer environment variables, then will
// fall back to a credentials file on disk. The following envvars specify an access
// and secret key, respectively.
//
// - AWS_ACCESS_KEY_ID or AWS_ACCESS_KEY
// - AWS_SECRET_ACCESS_KEY or AWS_SECRET_KEY
//
// If these variables are unset, then the client will read the credentails file at
// the path specified by AWS_SHARED_CREDENTIALS_FILE, or ~/.aws/credentials if not
// specified. The envvar AWS_PROFILE can be used to specify a non-default profile
// within the credentails file.
//
// To specify a non-default region or endpoint, supply the envvars AWS_REGION and
// AWS_ENDPOINT, respectively.
func awsSessionOptions() session.Options {
	return session.Options{
		Config: aws.Config{
			Credentials: credentials.NewCredentials(&credentials.ChainProvider{
				Providers: []credentials.Provider{
					&credentials.EnvProvider{},
					&credentials.SharedCredentialsProvider{},
				},
				VerboseErrors: true,
			}),
			Endpoint:         awsEnv("AWS_ENDPOINT"),
			Region:           awsEnv("AWS_REGION"),
			S3ForcePathStyle: aws.Bool(os.Getenv("AWS_S3_FORCE_PATH_STYLE") != ""),
		},
	}
}

func awsEnv(name string) *string {
	if value := os.Getenv(name); value != "" {
		return aws.String(value)
	}

	return nil
}
