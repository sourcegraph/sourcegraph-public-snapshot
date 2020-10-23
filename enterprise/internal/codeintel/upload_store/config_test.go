package uploadstore

import (
	"testing"
	"time"
)

func TestConfigS3(t *testing.T) {
	env := map[string]string{
		"PRECISE_CODE_INTEL_UPLOAD_BACKEND": "S3",
		"PRECISE_CODE_INTEL_UPLOAD_BUCKET":  "lsif-uploads",
		"PRECISE_CODE_INTEL_UPLOAD_TTL":     "8h",
	}

	config := Config{}
	config.SetMockGetter(mapGetter(env))
	config.Load()

	if err := config.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %s", err)
	}

	if config.S3.Bucket != "lsif-uploads" {
		t.Errorf("unexpected value for S3.Bucket. want=%s have=%s", "lsif-uploads", config.S3.Bucket)
	}
	if config.S3.TTL != 8*time.Hour {
		t.Errorf("unexpected value for S3.TTL. want=%v have=%v", 8*time.Hour, config.S3.TTL)
	}
}

func TestConfigGCS(t *testing.T) {
	env := map[string]string{
		"PRECISE_CODE_INTEL_UPLOAD_BACKEND":        "GCS",
		"PRECISE_CODE_INTEL_UPLOAD_BUCKET":         "lsif-uploads",
		"PRECISE_CODE_INTEL_UPLOAD_TTL":            "8h",
		"PRECISE_CODE_INTEL_UPLOAD_GCP_PROJECT_ID": "test",
	}

	config := Config{}
	config.SetMockGetter(mapGetter(env))
	config.Load()

	if err := config.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %s", err)
	}

	if config.GCS.Bucket != "lsif-uploads" {
		t.Errorf("unexpected value for GCS.Bucket. want=%s have=%s", "lsif-uploads", config.GCS.Bucket)
	}
	if config.GCS.TTL != 8*time.Hour {
		t.Errorf("unexpected value for GCS.TTL. want=%v have=%v", 8*time.Hour, config.GCS.TTL)
	}
	if config.GCS.ProjectID != "test" {
		t.Errorf("unexpected value for GCS.ProjectID. want=%s have=%s", "test", config.GCS.ProjectID)
	}
}

func mapGetter(env map[string]string) func(name, defaultValue, description string) string {
	return func(name, defaultValue, description string) string {
		return env[name]
	}
}
