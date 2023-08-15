package lsifuploadstore

import (
	"testing"
	"time"
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
	if config.S3AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("unexpected value for S3.AccessKeyID. want=%s have=%s", "AKIAIOSFODNN7EXAMPLE", config.S3AccessKeyID)
	}
	if config.S3SecretAccessKey != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("unexpected value for S3.SecretAccessKey. want=%s have=%s", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", config.S3SecretAccessKey)
	}
	if config.S3Endpoint != "http://blobstore:9000" {
		t.Errorf("unexpected value for S3.Endpoint. want=%s have=%s", "us-east-1", config.S3Endpoint)
	}
	if config.S3Region != "us-east-1" {
		t.Errorf("unexpected value for S3.Region. want=%s have=%s", "us-east-1", config.S3Region)
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
	if config.S3AccessKeyID != "access-key-id" {
		t.Errorf("unexpected value for S3.AccessKeyID. want=%s have=%s", "access-key-id", config.S3AccessKeyID)
	}
	if config.S3SecretAccessKey != "secret-access-key" {
		t.Errorf("unexpected value for S3.SecretAccessKey. want=%s have=%s", "secret-access-key", config.S3SecretAccessKey)
	}
	if config.S3SessionToken != "session-token" {
		t.Errorf("unexpected value for S3.SessionToken. want=%s have=%s", "session-token", config.S3SessionToken)
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
	if config.GCSProjectID != "test-project-id" {
		t.Errorf("unexpected value for GCS.ProjectID. want=%s have=%s", "tesT-project-id", config.GCSProjectID)
	}
	if config.GCSCredentialsFile != "test-credentials-file" {
		t.Errorf("unexpected value for GCS.CredentialsFile. want=%s have=%s", "test-credentials-file", config.GCSCredentialsFile)
	}
	if config.GCSCredentialsFileContents != "test-credentials-file-contents" {
		t.Errorf("unexpected value for GCS.CredentialsFileContents. want=%s have=%s", "test-credentials-file-contents", config.GCSCredentialsFileContents)
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
