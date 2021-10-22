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

func TestConfigDefaults(t *testing.T) {
	config := Config{}
	config.SetMockGetter(mapGetter(nil))
	config.Load()

	if err := config.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %s", err)
	}

	if config.Bucket != "lsif-uploads" {
		t.Errorf("unexpected value for S3.Bucket. want=%s have=%s", "lsif-uploads", config.Bucket)
	}
	if config.TTL != 24*7*time.Hour {
		t.Errorf("unexpected value for S3.TTL. want=%v have=%v", 24*7*time.Hour, config.TTL)
	}
	if config.S3.AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("unexpected value for S3.AccessKeyID. want=%s have=%s", "AKIAIOSFODNN7EXAMPLE", config.S3.AccessKeyID)
	}
	if config.S3.SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("unexpected value for S3.SecretAccessKey. want=%s have=%s", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", config.S3.SecretAccessKey)
	}
	if config.S3.Endpoint != "http://minio:9000" {
		t.Errorf("unexpected value for S3.Endpoint. want=%s have=%s", "us-east-1", config.S3.Endpoint)
	}
	if config.S3.Region != "us-east-1" {
		t.Errorf("unexpected value for S3.Region. want=%s have=%s", "us-east-1", config.S3.Region)
	}
}

func TestConfigS3(t *testing.T) {
	env := map[string]string{
		"PRECISE_CODE_INTEL_UPLOAD_BACKEND":               "S3",
		"PRECISE_CODE_INTEL_UPLOAD_BUCKET":                "lsif-uploads",
		"PRECISE_CODE_INTEL_UPLOAD_TTL":                   "8h",
		"PRECISE_CODE_INTEL_UPLOAD_MANAGE_BUCKET":         "true",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_ACCESS_KEY_ID":     "access-key-id",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_SECRET_ACCESS_KEY": "secret-access-key",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_SESSION_TOKEN":     "session-token",
	}

	config := Config{}
	config.SetMockGetter(mapGetter(env))
	config.Load()

	if err := config.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %s", err)
	}

	if config.Bucket != "lsif-uploads" {
		t.Errorf("unexpected value for S3.Bucket. want=%s have=%s", "lsif-uploads", config.Bucket)
	}
	if config.TTL != 8*time.Hour {
		t.Errorf("unexpected value for S3.TTL. want=%v have=%v", 8*time.Hour, config.TTL)
	}
	if config.S3.AccessKeyID != "access-key-id" {
		t.Errorf("unexpected value for S3.AccessKeyID. want=%s have=%s", "access-key-id", config.S3.AccessKeyID)
	}
	if config.S3.SecretAccessKey != "secret-access-key" {
		t.Errorf("unexpected value for S3.SecretAccessKey. want=%s have=%s", "secret-access-key", config.S3.SecretAccessKey)
	}
	if config.S3.SessionToken != "session-token" {
		t.Errorf("unexpected value for S3.SessionToken. want=%s have=%s", "session-token", config.S3.SessionToken)
	}
}

func TestS3ClientOptions(t *testing.T) {
	config := Config{}
	config.SetMockGetter(mapGetter(nil))
	config.Load()

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
	env := map[string]string{
		"PRECISE_CODE_INTEL_UPLOAD_BACKEND":               "S3",
		"PRECISE_CODE_INTEL_UPLOAD_BUCKET":                "lsif-uploads",
		"PRECISE_CODE_INTEL_UPLOAD_TTL":                   "8h",
		"PRECISE_CODE_INTEL_UPLOAD_MANAGE_BUCKET":         "true",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_REGION":            "us-east-2",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_ACCESS_KEY_ID":     "access-key-id",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_SECRET_ACCESS_KEY": "secret-access-key",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_SESSION_TOKEN":     "session-token",
	}

	config := Config{}
	config.SetMockGetter(mapGetter(env))
	config.Load()

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
	if cfg.EndpointResolver != nil {
		t.Errorf("unexpected endpoint option")
	}
}

func TestConfigGCS(t *testing.T) {
	env := map[string]string{
		"PRECISE_CODE_INTEL_UPLOAD_BACKEND":                                     "GCS",
		"PRECISE_CODE_INTEL_UPLOAD_BUCKET":                                      "lsif-uploads",
		"PRECISE_CODE_INTEL_UPLOAD_TTL":                                         "8h",
		"PRECISE_CODE_INTEL_UPLOAD_MANAGE_BUCKET":                               "true",
		"PRECISE_CODE_INTEL_UPLOAD_GCP_PROJECT_ID":                              "test-project-id",
		"PRECISE_CODE_INTEL_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE":         "test-credentials-file",
		"PRECISE_CODE_INTEL_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT": "test-credentials-file-contents",
	}

	config := Config{}
	config.SetMockGetter(mapGetter(env))
	config.Load()

	if err := config.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %s", err)
	}

	if config.Bucket != "lsif-uploads" {
		t.Errorf("unexpected value for GCS.Bucket. want=%s have=%s", "lsif-uploads", config.Bucket)
	}
	if config.TTL != 8*time.Hour {
		t.Errorf("unexpected value for GCS.TTL. want=%v have=%v", 8*time.Hour, config.TTL)
	}
	if config.GCS.ProjectID != "test-project-id" {
		t.Errorf("unexpected value for GCS.ProjectID. want=%s have=%s", "tesT-project-id", config.GCS.ProjectID)
	}
	if config.GCS.CredentialsFile != "test-credentials-file" {
		t.Errorf("unexpected value for GCS.CredentialsFile. want=%s have=%s", "test-credentials-file", config.GCS.CredentialsFile)
	}
	if config.GCS.CredentialsFileContents != "test-credentials-file-contents" {
		t.Errorf("unexpected value for GCS.CredentialsFileContents. want=%s have=%s", "test-credentials-file-contents", config.GCS.CredentialsFileContents)
	}
}

func mapGetter(env map[string]string) func(name, defaultValue, description string) string {
	return func(name, defaultValue, description string) string {
		if v, ok := env[name]; ok {
			return v
		}

		return defaultValue
	}
}
