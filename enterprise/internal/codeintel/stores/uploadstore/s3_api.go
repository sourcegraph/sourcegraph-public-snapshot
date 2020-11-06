package uploadstore

import (
	"context"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	PutBucketLifecycleConfiguration(ctx context.Context, input *s3.PutBucketLifecycleConfigurationInput) (*s3.PutBucketLifecycleConfigurationOutput, error)
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

func (s *s3APIShim) PutBucketLifecycleConfiguration(ctx context.Context, input *s3.PutBucketLifecycleConfigurationInput) (*s3.PutBucketLifecycleConfigurationOutput, error) {
	return s.S3.PutBucketLifecycleConfigurationWithContext(ctx, input)
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
