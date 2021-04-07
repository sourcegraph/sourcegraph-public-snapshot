package uploadstore

import (
	"context"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/inconshreveable/log15"
)

type s3API interface {
	HeadObject(ctx context.Context, input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error)
	GetObject(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	CreateMultipartUpload(ctx context.Context, input *s3.CreateMultipartUploadInput) (*s3.CreateMultipartUploadOutput, error)
	AbortMultipartUpload(ctx context.Context, input *s3.AbortMultipartUploadInput) (*s3.AbortMultipartUploadOutput, error)
	DeleteObject(ctx context.Context, input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	UploadPartCopy(ctx context.Context, input *s3.UploadPartCopyInput) (*s3.UploadPartCopyOutput, error)
	CompleteMultipartUpload(ctx context.Context, input *s3.CompleteMultipartUploadInput) (*s3.CompleteMultipartUploadOutput, error)
	CreateBucket(ctx context.Context, input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error)
	EnforceBucketLifecycle(ctx context.Context, bucket string, ttl time.Duration) error
}

type s3Uploader interface {
	Upload(ctx context.Context, input *s3manager.UploadInput) error
}

type s3APIShim struct{ *s3.S3 }
type s3UploaderShim struct{ *s3manager.Uploader }

var _ s3API = &s3APIShim{}
var _ s3Uploader = &s3UploaderShim{}

func (s *s3APIShim) CreateBucket(ctx context.Context, input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	return s.S3.CreateBucketWithContext(ctx, input)
}

func (s *s3APIShim) HeadObject(ctx context.Context, input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	return s.S3.HeadObjectWithContext(ctx, input)
}

func (s *s3APIShim) GetObject(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return s.S3.GetObjectWithContext(ctx, input)
}

func (s *s3APIShim) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return s.S3.DeleteObjectWithContext(ctx, input)
}

func (s *s3APIShim) CreateMultipartUpload(ctx context.Context, input *s3.CreateMultipartUploadInput) (*s3.CreateMultipartUploadOutput, error) {
	return s.S3.CreateMultipartUploadWithContext(ctx, input)
}

func (s *s3APIShim) AbortMultipartUpload(ctx context.Context, input *s3.AbortMultipartUploadInput) (*s3.AbortMultipartUploadOutput, error) {
	return s.S3.AbortMultipartUploadWithContext(ctx, input)
}

func (s *s3APIShim) UploadPartCopy(ctx context.Context, input *s3.UploadPartCopyInput) (*s3.UploadPartCopyOutput, error) {
	return s.S3.UploadPartCopyWithContext(ctx, input)
}

func (s *s3APIShim) CompleteMultipartUpload(ctx context.Context, input *s3.CompleteMultipartUploadInput) (*s3.CompleteMultipartUploadOutput, error) {
	return s.S3.CompleteMultipartUploadWithContext(ctx, input)
}

func (s *s3UploaderShim) Upload(ctx context.Context, input *s3manager.UploadInput) error {
	_, err := s.Uploader.UploadWithContext(ctx, input)
	return err
}

func (s *s3APIShim) EnforceBucketLifecycle(ctx context.Context, bucket string, ttl time.Duration) error {
	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket:                 aws.String(bucket),
		LifecycleConfiguration: lifecycle(ttl),
	}

	_, err := s.S3.PutBucketLifecycleConfigurationWithContext(ctx, input)
	return err
}

func lifecycle(ttl time.Duration) *s3.BucketLifecycleConfiguration {
	days := aws.Int64(int64(ttl / (time.Hour * 24)))

	return &s3.BucketLifecycleConfiguration{
		Rules: []*s3.LifecycleRule{
			{
				ID:         aws.String("Expiration Rule"),
				Status:     aws.String("Enabled"),
				Filter:     &s3.LifecycleRuleFilter{Prefix: aws.String("")},
				Expiration: &s3.LifecycleExpiration{Days: days},
			},
			{
				ID:                             aws.String("Abort Incomplete Multipart Upload Rule"),
				Status:                         aws.String("Enabled"),
				Filter:                         &s3.LifecycleRuleFilter{Prefix: aws.String("")},
				AbortIncompleteMultipartUpload: &s3.AbortIncompleteMultipartUpload{DaysAfterInitiation: days},
			},
		},
	}
}

// minioAPIShim is a decorator around s3APIShim that intercepts invocations of the
// EnforceBucketLifecycle method. There is a protocol mismatch between the
// AWS SDK and MinIO, which we distribute as an embedded object store.
//
// Instead of nerfing this function in GCS and S3 (where lifecycles are respected),
// we change the behavior only for the MinIO backend. Instead of pushing a lifecycle
// configuration to the bucket, we'll periodically scan through the files in the
// target bucket and delete those older than the threshold age.
type minioAPIShim struct{ s3APIShim }

// EnforceBucketLifecycle starts two goroutines that communicate via a shared channel. The
// writer goroutine periodically scans the target bucket for objects older than the given
// max age, then writes matching keys to the channel. The reader goroutine will delete each
// key as it is written.
func (s *minioAPIShim) EnforceBucketLifecycle(ctx context.Context, bucket string, ttl time.Duration) error {
	// Small buffer allows us to deal with short "bursts" of expired keys without
	// having to synchronously delete all of the keys before the scan can end.
	expiredKeys := make(chan *string, 128)

	go func() {
		for key := range expiredKeys {
			if _, err := s.S3.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: key}); err != nil {
				log15.Error("Failed to delete expired upload file", "error", err, "bucket", bucket, "key", *key)
			}
		}
	}()

	go func() {
		defer close(expiredKeys)

		for {
			pager := func(output *s3.ListObjectsOutput, lastPage bool) bool {
				for _, object := range output.Contents {
					if object.LastModified != nil && object.LastModified.Before(time.Now().Add(-ttl)) {
						expiredKeys <- object.Key
					} else {
					}
				}

				return true
			}

			if err := s.S3.ListObjectsPages(&s3.ListObjectsInput{Bucket: aws.String(bucket)}, pager); err != nil {
				log15.Error("Failed to list upload files to check expiry", "error", err, "bucket", bucket)
			}

			// Wait 30 +/- 5 minutes between scans of the object store. Expiration is
			// on the matter of days, so there's no reason to this very frequently.
			<-time.After(time.Minute*30 + time.Duration(rand.Intn(600)-300)*time.Second)
		}
	}()

	return nil
}
