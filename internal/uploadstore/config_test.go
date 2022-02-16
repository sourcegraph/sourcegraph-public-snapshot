package uploadstore

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestS3ClientOptions(t *testing.T) {
	config := Config{
		S3: S3Config{
			Endpoint: "http://minio:9000",
		},
	}

	// minIO
	{
		options := &s3.Options{}
		s3ClientOptions("minio", config.S3)(options)

		if options.EndpointResolver == nil {
			t.Fatalf("unexpected endpoint option")
		}
		endpoint, err := options.EndpointResolver.ResolveEndpoint("us-east-2", s3.EndpointResolverOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if endpoint.URL != "http://minio:9000" {
			t.Errorf("unexpected endpoint. want=%s have=%s", "http://minio:9000", endpoint.URL)
		}

		if !options.UsePathStyle {
			t.Errorf("invalid UsePathStyle setting for S3Options")
		}
	}

	// S3
	{
		options := &s3.Options{}
		s3ClientOptions("s3", config.S3)(options)

		if diff := cmp.Diff(&s3.Options{}, options, cmpopts.IgnoreUnexported(s3.Options{})); diff != "" {
			t.Fatalf("invalid s3 options returned for S3: %s", diff)
		}
	}
}

func TestS3ClientConfig(t *testing.T) {
	config := Config{
		Backend:      "s3",
		Bucket:       "lsif-uploads",
		ManageBucket: true,
		TTL:          8 * time.Hour,
		S3: S3Config{
			Region:          "us-east-2",
			AccessKeyID:     "access-key-id",
			SecretAccessKey: "secret-access-key",
			SessionToken:    "session-token",
		},
	}

	cfg, err := s3ClientConfig(context.Background(), config.S3)
	if err != nil {
		t.Fatal(err)
	}

	if value := cfg.Region; value != "us-east-2" {
		t.Errorf("unexpected region. want=%s have=%s", "us-east-2", value)
	}
	cred, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(aws.Credentials{
		AccessKeyID:     config.S3.AccessKeyID,
		SecretAccessKey: config.S3.SecretAccessKey,
		SessionToken:    config.S3.SessionToken,
		Source:          "StaticCredentials",
	}, cred); diff != "" {
		t.Errorf("invalid credential returned: %s", diff)
	}
	if cfg.EndpointResolverWithOptions != nil {
		t.Errorf("unexpected endpoint option")
	}
}
