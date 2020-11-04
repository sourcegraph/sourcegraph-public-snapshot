package uploadstore

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/go-cmp/cmp"
)

func TestS3Init(t *testing.T) {
	s3Client := NewMockS3API()
	client := newS3WithClients(s3Client, nil, "test-bucket", time.Hour*24, true)
	if err := client.Init(context.Background()); err != nil {
		t.Fatalf("unexpected error initializing client: %s", err.Error())
	}

	if calls := s3Client.CreateBucketFunc.History(); len(calls) != 1 {
		t.Fatalf("unexpected number of CreateBucket calls. want=%d have=%d", 1, len(calls))
	} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
		t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
	}

	if calls := s3Client.PutBucketLifecycleConfigurationFunc.History(); len(calls) != 1 {
		t.Fatalf("unexpected number of PutBucketLifecycleConfiguration calls. want=%d have=%d", 1, len(calls))
	} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
		t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
	}
}

func TestS3InitBucketExists(t *testing.T) {
	for _, code := range []string{s3.ErrCodeBucketAlreadyExists, s3.ErrCodeBucketAlreadyOwnedByYou} {
		s3Client := NewMockS3API()
		s3Client.CreateBucketFunc.SetDefaultReturn(nil, awserr.New(code, "", nil))

		client := newS3WithClients(s3Client, nil, "test-bucket", time.Hour*24, true)
		if err := client.Init(context.Background()); err != nil {
			t.Fatalf("unexpected error initializing client: %s", err.Error())
		}

		if calls := s3Client.CreateBucketFunc.History(); len(calls) != 1 {
			t.Fatalf("unexpected number of CreateBucket calls. want=%d have=%d", 1, len(calls))
		} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
			t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
		}

		if calls := s3Client.PutBucketLifecycleConfigurationFunc.History(); len(calls) != 1 {
			t.Fatalf("unexpected number of PutBucketLifecycleConfiguration calls. want=%d have=%d", 1, len(calls))
		} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
			t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
		}
	}
}

func TestS3UnmanagedInit(t *testing.T) {
	s3Client := NewMockS3API()
	client := newS3WithClients(s3Client, nil, "test-bucket", time.Hour*24, false)
	if err := client.Init(context.Background()); err != nil {
		t.Fatalf("unexpected error initializing client: %s", err.Error())
	}

	if calls := s3Client.CreateBucketFunc.History(); len(calls) != 0 {
		t.Fatalf("unexpected number of CreateBucket calls. want=%d have=%d", 0, len(calls))
	}
	if calls := s3Client.PutBucketLifecycleConfigurationFunc.History(); len(calls) != 0 {
		t.Fatalf("unexpected number of PutBucketLifecycleConfiguration calls. want=%d have=%d", 0, len(calls))
	}
}

func TestS3Get(t *testing.T) {
	s3Client := NewMockS3API()
	s3Client.GetObjectFunc.SetDefaultReturn(&s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewReader([]byte("TEST PAYLOAD"))),
	}, nil)

	client := newS3WithClients(s3Client, nil, "test-bucket", time.Hour*24, false)
	rc, err := client.Get(context.Background(), "test-key", 0)
	if err != nil {
		t.Fatalf("unexpected error getting key: %s", err.Error())
	}

	defer rc.Close()
	contents, err := ioutil.ReadAll(rc)
	if err != nil {
		t.Fatalf("unexpected error reading object: %s", err.Error())
	}

	if string(contents) != "TEST PAYLOAD" {
		t.Fatalf("unexpected contents. want=%s have=%s", "TEST PAYLOAD", contents)
	}

	if calls := s3Client.GetObjectFunc.History(); len(calls) != 1 {
		t.Fatalf("unexpected number of GetObject calls. want=%d have=%d", 1, len(calls))
	} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
		t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
	} else if value := *calls[0].Arg1.Key; value != "test-key" {
		t.Errorf("unexpected key argument. want=%s have=%s", "test-key", value)
	} else if value := calls[0].Arg1.Range; value != nil {
		t.Errorf("unexpected range argument. want=%v have=%v", nil, value)
	}
}

func TestS3GetSkipBytes(t *testing.T) {
	s3Client := NewMockS3API()
	s3Client.GetObjectFunc.SetDefaultReturn(&s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewReader([]byte("TEST PAYLOAD"))),
	}, nil)

	client := newS3WithClients(s3Client, nil, "test-bucket", time.Hour*24, false)
	rc, err := client.Get(context.Background(), "test-key", 20)
	if err != nil {
		t.Fatalf("unexpected error getting key: %s", err.Error())
	}

	defer rc.Close()
	contents, err := ioutil.ReadAll(rc)
	if err != nil {
		t.Fatalf("unexpected error reading object: %s", err.Error())
	}

	if string(contents) != "TEST PAYLOAD" {
		t.Fatalf("unexpected contents. want=%s have=%s", "TEST PAYLOAD", contents)
	}

	if calls := s3Client.GetObjectFunc.History(); len(calls) != 1 {
		t.Fatalf("unexpected number of GetObject calls. want=%d have=%d", 1, len(calls))
	} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
		t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
	} else if value := *calls[0].Arg1.Key; value != "test-key" {
		t.Errorf("unexpected key argument. want=%s have=%s", "test-key", value)
	} else if value := *calls[0].Arg1.Range; value != "bytes=20-" {
		t.Errorf("unexpected range argument. want=%s have=%s", "", value)
	}
}

func TestS3Upload(t *testing.T) {
	s3Client := NewMockS3API()
	uploaderClient := NewMockS3Uploader()
	uploaderClient.UploadFunc.SetDefaultHook(func(ctx context.Context, input *s3manager.UploadInput) error {
		// Synchronously read the reader so that we trigger the
		// counting reader inside the Upload method and test the
		// count.
		contents, err := ioutil.ReadAll(input.Body)
		if err != nil {
			return err
		}

		if string(contents) != "TEST PAYLOAD" {
			t.Fatalf("unexpected contents. want=%s have=%s", "TEST PAYLOAD", contents)
		}

		return nil
	})

	client := newS3WithClients(s3Client, uploaderClient, "test-bucket", time.Hour*24, false)

	size, err := client.Upload(context.Background(), "test-key", bytes.NewReader([]byte("TEST PAYLOAD")))
	if err != nil {
		t.Fatalf("unexpected error getting key: %s", err.Error())
	} else if size != 12 {
		t.Errorf("unexpected size. want=%d have=%d", 12, size)
	}

	if calls := uploaderClient.UploadFunc.History(); len(calls) != 1 {
		t.Fatalf("unexpected number of Upload calls. want=%d have=%d", 1, len(calls))
	} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
		t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
	} else if value := *calls[0].Arg1.Key; value != "test-key" {
		t.Errorf("unexpected key argument. want=%s have=%s", "test-key", value)
	}
}

func TestS3Combine(t *testing.T) {
	s3Client := NewMockS3API()
	s3Client.CreateMultipartUploadFunc.SetDefaultReturn(&s3.CreateMultipartUploadOutput{
		Bucket:   aws.String("test-bucket"),
		Key:      aws.String("test-key"),
		UploadId: aws.String("uid"),
	}, nil)

	s3Client.UploadPartCopyFunc.SetDefaultHook(func(ctx context.Context, input *s3.UploadPartCopyInput) (*s3.UploadPartCopyOutput, error) {
		return &s3.UploadPartCopyOutput{
			CopyPartResult: &s3.CopyPartResult{
				ETag: aws.String(fmt.Sprintf("etag-%s", *input.CopySource)),
			},
		}, nil
	})

	s3Client.HeadObjectFunc.SetDefaultReturn(&s3.HeadObjectOutput{ContentLength: aws.Int64(42)}, nil)

	client := newS3WithClients(s3Client, nil, "test-bucket", time.Hour*24, false)

	size, err := client.Compose(context.Background(), "test-key", "test-src1", "test-src2", "test-src3")
	if err != nil {
		t.Fatalf("unexpected error getting key: %s", err.Error())
	} else if size != 42 {
		t.Errorf("unexpected size. want=%d have=%d", 42, size)
	}

	if calls := s3Client.UploadPartCopyFunc.History(); len(calls) != 3 {
		t.Fatalf("unexpected number of UploadPartCopy calls. want=%d have=%d", 3, len(calls))
	} else {
		parts := map[int64]string{}
		for _, call := range calls {
			if value := *call.Arg1.Bucket; value != "test-bucket" {
				t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
			}
			if value := *call.Arg1.Key; value != "test-key" {
				t.Errorf("unexpected key argument. want=%s have=%s", "test-key", value)
			}
			if value := *call.Arg1.UploadId; value != "uid" {
				t.Errorf("unexpected key argument. want=%s have=%s", "uid", value)
			}

			parts[*call.Arg1.PartNumber] = *call.Arg1.CopySource
		}

		expectedParts := map[int64]string{
			1: "test-bucket/test-src1",
			2: "test-bucket/test-src2",
			3: "test-bucket/test-src3",
		}
		if diff := cmp.Diff(expectedParts, parts); diff != "" {
			t.Fatalf("unexpected parts payloads (-want, +got):\n%s", diff)
		}
	}

	if calls := s3Client.CreateMultipartUploadFunc.History(); len(calls) != 1 {
		t.Fatalf("unexpected number of CreateMultipartUpload calls. want=%d have=%d", 1, len(calls))
	} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
		t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
	} else if value := *calls[0].Arg1.Key; value != "test-key" {
		t.Errorf("unexpected key argument. want=%s have=%s", "test-key", value)
	}

	if calls := s3Client.CompleteMultipartUploadFunc.History(); len(calls) != 1 {
		t.Fatalf("unexpected number of CompleteMultipartUpload calls. want=%d have=%d", 1, len(calls))
	} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
		t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
	} else if value := *calls[0].Arg1.Key; value != "test-key" {
		t.Errorf("unexpected key argument. want=%s have=%s", "test-key", value)
	} else if value := *calls[0].Arg1.UploadId; value != "uid" {
		t.Errorf("unexpected uploadId argument. want=%s have=%s", "uid", value)
	} else {
		parts := map[int64]string{}
		for _, part := range calls[0].Arg1.MultipartUpload.Parts {
			parts[*part.PartNumber] = *part.ETag
		}

		expectedParts := map[int64]string{
			1: "etag-test-bucket/test-src1",
			2: "etag-test-bucket/test-src2",
			3: "etag-test-bucket/test-src3",
		}
		if diff := cmp.Diff(expectedParts, parts); diff != "" {
			t.Fatalf("unexpected parts payloads (-want, +got):\n%s", diff)
		}
	}

	if calls := s3Client.AbortMultipartUploadFunc.History(); len(calls) != 0 {
		t.Fatalf("unexpected number of AbortMultipartUpload calls. want=%d have=%d", 0, len(calls))
	}

	if calls := s3Client.DeleteObjectFunc.History(); len(calls) != 3 {
		t.Fatalf("unexpected number of DeleteObject calls. want=%d have=%d", 3, len(calls))
	} else {
		var keys []string
		for _, call := range calls {
			if value := *call.Arg1.Bucket; value != "test-bucket" {
				t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
			}
			keys = append(keys, *call.Arg1.Key)
		}
		sort.Strings(keys)

		expectedKeys := []string{
			"test-src1",
			"test-src2",
			"test-src3",
		}
		if diff := cmp.Diff(expectedKeys, keys); diff != "" {
			t.Fatalf("unexpected keys (-want, +got):\n%s", diff)
		}
	}
}

func TestS3Delete(t *testing.T) {
	s3Client := NewMockS3API()
	s3Client.GetObjectFunc.SetDefaultReturn(&s3.GetObjectOutput{
		Body: ioutil.NopCloser(bytes.NewReader([]byte("TEST PAYLOAD"))),
	}, nil)

	client := newS3WithClients(s3Client, nil, "test-bucket", time.Hour*24, false)
	if err := client.Delete(context.Background(), "test-key"); err != nil {
		t.Fatalf("unexpected error getting key: %s", err.Error())
	}

	if calls := s3Client.DeleteObjectFunc.History(); len(calls) != 1 {
		t.Fatalf("unexpected number of DeleteObject calls. want=%d have=%d", 1, len(calls))
	} else if value := *calls[0].Arg1.Bucket; value != "test-bucket" {
		t.Errorf("unexpected bucket argument. want=%s have=%s", "test-bucket", value)
	} else if value := *calls[0].Arg1.Key; value != "test-key" {
		t.Errorf("unexpected key argument. want=%s have=%s", "test-key", value)
	}
}

func TestS3Lifecycle(t *testing.T) {
	s3Client := NewMockS3API()
	client := newS3WithClients(s3Client, nil, "test-bucket", time.Hour*24*3, true)

	if lifecycle := client.lifecycle(); lifecycle == nil || len(lifecycle.Rules) != 2 {
		t.Fatalf("unexpected lifecycle rules")
	} else {
		var objectExpiration *int64
		for _, rule := range lifecycle.Rules {
			if rule.Expiration != nil {
				if value := rule.Expiration.Days; value != nil {
					objectExpiration = value
				}
			}
		}
		if objectExpiration == nil {
			t.Fatalf("expected object expiration to be configured")
		} else if *objectExpiration != 3 {
			t.Errorf("unexpected ttl for object expiration. want=%d have=%d", 3, *objectExpiration)
		}

		var multipartExpiration *int64
		for _, rule := range lifecycle.Rules {
			if rule.AbortIncompleteMultipartUpload != nil {
				if value := rule.AbortIncompleteMultipartUpload.DaysAfterInitiation; value != nil {
					multipartExpiration = value
				}
			}
		}
		if multipartExpiration == nil {
			t.Fatalf("expected multipart upload expiration to be configured")
		} else if *multipartExpiration != 3 {
			t.Errorf("unexpected ttl for multipart upload expiration. want=%d have=%d", 3, *multipartExpiration)
		}
	}
}
