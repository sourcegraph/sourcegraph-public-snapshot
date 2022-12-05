package uploadstore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	NewListObjectsV2Paginator(input *s3.ListObjectsV2Input) *s3.ListObjectsV2Paginator
}

type s3Uploader interface {
	Upload(ctx context.Context, input *s3.PutObjectInput) error
}

type (
	s3APIShim      struct{ *s3.Client }
	s3UploaderShim struct{ *manager.Uploader }
)

var (
	_ s3API      = &s3APIShim{}
	_ s3Uploader = &s3UploaderShim{}
)

func (s *s3APIShim) CreateBucket(ctx context.Context, input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	return s.Client.CreateBucket(ctx, input)
}

func (s *s3APIShim) HeadObject(ctx context.Context, input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	return s.Client.HeadObject(ctx, input)
}

func (s *s3APIShim) GetObject(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return s.Client.GetObject(ctx, input)
}

func (s *s3APIShim) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return s.Client.DeleteObject(ctx, input)
}

func (s *s3APIShim) CreateMultipartUpload(ctx context.Context, input *s3.CreateMultipartUploadInput) (*s3.CreateMultipartUploadOutput, error) {
	return s.Client.CreateMultipartUpload(ctx, input)
}

func (s *s3APIShim) AbortMultipartUpload(ctx context.Context, input *s3.AbortMultipartUploadInput) (*s3.AbortMultipartUploadOutput, error) {
	return s.Client.AbortMultipartUpload(ctx, input)
}

func (s *s3APIShim) UploadPartCopy(ctx context.Context, input *s3.UploadPartCopyInput) (*s3.UploadPartCopyOutput, error) {
	return s.Client.UploadPartCopy(ctx, input)
}

func (s *s3APIShim) CompleteMultipartUpload(ctx context.Context, input *s3.CompleteMultipartUploadInput) (*s3.CompleteMultipartUploadOutput, error) {
	return s.Client.CompleteMultipartUpload(ctx, input)
}

func (s *s3APIShim) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return s.Client.DeleteObjects(ctx, params, optFns...)
}

func (s *s3APIShim) NewListObjectsV2Paginator(input *s3.ListObjectsV2Input) *s3.ListObjectsV2Paginator {
	return s3.NewListObjectsV2Paginator(s.Client, input)
}

func (s *s3UploaderShim) Upload(ctx context.Context, input *s3.PutObjectInput) error {
	_, err := s.Uploader.Upload(ctx, input)
	return err
}
