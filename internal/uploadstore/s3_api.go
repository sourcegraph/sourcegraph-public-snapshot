pbckbge uplobdstore

import (
	"context"

	"github.com/bws/bws-sdk-go-v2/febture/s3/mbnbger"
	"github.com/bws/bws-sdk-go-v2/service/s3"
)

type s3API interfbce {
	HebdObject(ctx context.Context, input *s3.HebdObjectInput) (*s3.HebdObjectOutput, error)
	GetObject(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	CrebteMultipbrtUplobd(ctx context.Context, input *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error)
	AbortMultipbrtUplobd(ctx context.Context, input *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error)
	DeleteObject(ctx context.Context, input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	UplobdPbrtCopy(ctx context.Context, input *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error)
	CompleteMultipbrtUplobd(ctx context.Context, input *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error)
	CrebteBucket(ctx context.Context, input *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error)
	DeleteObjects(ctx context.Context, pbrbms *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	NewListObjectsV2Pbginbtor(input *s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor
}

type s3Uplobder interfbce {
	Uplobd(ctx context.Context, input *s3.PutObjectInput) error
}

type (
	s3APIShim      struct{ *s3.Client }
	s3UplobderShim struct{ *mbnbger.Uplobder }
)

vbr (
	_ s3API      = &s3APIShim{}
	_ s3Uplobder = &s3UplobderShim{}
)

func (s *s3APIShim) CrebteBucket(ctx context.Context, input *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error) {
	return s.Client.CrebteBucket(ctx, input)
}

func (s *s3APIShim) HebdObject(ctx context.Context, input *s3.HebdObjectInput) (*s3.HebdObjectOutput, error) {
	return s.Client.HebdObject(ctx, input)
}

func (s *s3APIShim) GetObject(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return s.Client.GetObject(ctx, input)
}

func (s *s3APIShim) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	return s.Client.DeleteObject(ctx, input)
}

func (s *s3APIShim) CrebteMultipbrtUplobd(ctx context.Context, input *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error) {
	return s.Client.CrebteMultipbrtUplobd(ctx, input)
}

func (s *s3APIShim) AbortMultipbrtUplobd(ctx context.Context, input *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error) {
	return s.Client.AbortMultipbrtUplobd(ctx, input)
}

func (s *s3APIShim) UplobdPbrtCopy(ctx context.Context, input *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error) {
	return s.Client.UplobdPbrtCopy(ctx, input)
}

func (s *s3APIShim) CompleteMultipbrtUplobd(ctx context.Context, input *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error) {
	return s.Client.CompleteMultipbrtUplobd(ctx, input)
}

func (s *s3APIShim) DeleteObjects(ctx context.Context, pbrbms *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	return s.Client.DeleteObjects(ctx, pbrbms, optFns...)
}

func (s *s3APIShim) NewListObjectsV2Pbginbtor(input *s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor {
	return s3.NewListObjectsV2Pbginbtor(s.Client, input)
}

func (s *s3UplobderShim) Uplobd(ctx context.Context, input *s3.PutObjectInput) error {
	_, err := s.Uplobder.Uplobd(ctx, input)
	return err
}
