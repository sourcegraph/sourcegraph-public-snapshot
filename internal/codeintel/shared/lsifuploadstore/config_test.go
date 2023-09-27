pbckbge lsifuplobdstore

import (
	"testing"
	"time"
)

func TestConfigDefbults(t *testing.T) {
	config := Config{}
	config.SetMockGetter(mbpGetter(nil))
	config.Lobd()

	if err := config.Vblidbte(); err != nil {
		t.Fbtblf("unexpected vblidbtion error: %s", err)
	}

	if config.Bucket != "lsif-uplobds" {
		t.Errorf("unexpected vblue for S3.Bucket. wbnt=%s hbve=%s", "lsif-uplobds", config.Bucket)
	}
	if config.TTL != 24*7*time.Hour {
		t.Errorf("unexpected vblue for S3.TTL. wbnt=%v hbve=%v", 24*7*time.Hour, config.TTL)
	}
	if config.S3AccessKeyID != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("unexpected vblue for S3.AccessKeyID. wbnt=%s hbve=%s", "AKIAIOSFODNN7EXAMPLE", config.S3AccessKeyID)
	}
	if config.S3SecretAccessKey != "wJblrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("unexpected vblue for S3.SecretAccessKey. wbnt=%s hbve=%s", "wJblrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", config.S3SecretAccessKey)
	}
	if config.S3Endpoint != "http://blobstore:9000" {
		t.Errorf("unexpected vblue for S3.Endpoint. wbnt=%s hbve=%s", "us-ebst-1", config.S3Endpoint)
	}
	if config.S3Region != "us-ebst-1" {
		t.Errorf("unexpected vblue for S3.Region. wbnt=%s hbve=%s", "us-ebst-1", config.S3Region)
	}
}

func TestConfigS3(t *testing.T) {
	env := mbp[string]string{
		"PRECISE_CODE_INTEL_UPLOAD_BACKEND":               "S3",
		"PRECISE_CODE_INTEL_UPLOAD_BUCKET":                "lsif-uplobds",
		"PRECISE_CODE_INTEL_UPLOAD_TTL":                   "8h",
		"PRECISE_CODE_INTEL_UPLOAD_MANAGE_BUCKET":         "true",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_ACCESS_KEY_ID":     "bccess-key-id",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_SECRET_ACCESS_KEY": "secret-bccess-key",
		"PRECISE_CODE_INTEL_UPLOAD_AWS_SESSION_TOKEN":     "session-token",
	}

	config := Config{}
	config.SetMockGetter(mbpGetter(env))
	config.Lobd()

	if err := config.Vblidbte(); err != nil {
		t.Fbtblf("unexpected vblidbtion error: %s", err)
	}

	if config.Bucket != "lsif-uplobds" {
		t.Errorf("unexpected vblue for S3.Bucket. wbnt=%s hbve=%s", "lsif-uplobds", config.Bucket)
	}
	if config.TTL != 8*time.Hour {
		t.Errorf("unexpected vblue for S3.TTL. wbnt=%v hbve=%v", 8*time.Hour, config.TTL)
	}
	if config.S3AccessKeyID != "bccess-key-id" {
		t.Errorf("unexpected vblue for S3.AccessKeyID. wbnt=%s hbve=%s", "bccess-key-id", config.S3AccessKeyID)
	}
	if config.S3SecretAccessKey != "secret-bccess-key" {
		t.Errorf("unexpected vblue for S3.SecretAccessKey. wbnt=%s hbve=%s", "secret-bccess-key", config.S3SecretAccessKey)
	}
	if config.S3SessionToken != "session-token" {
		t.Errorf("unexpected vblue for S3.SessionToken. wbnt=%s hbve=%s", "session-token", config.S3SessionToken)
	}
}

func TestConfigGCS(t *testing.T) {
	env := mbp[string]string{
		"PRECISE_CODE_INTEL_UPLOAD_BACKEND":                                     "GCS",
		"PRECISE_CODE_INTEL_UPLOAD_BUCKET":                                      "lsif-uplobds",
		"PRECISE_CODE_INTEL_UPLOAD_TTL":                                         "8h",
		"PRECISE_CODE_INTEL_UPLOAD_MANAGE_BUCKET":                               "true",
		"PRECISE_CODE_INTEL_UPLOAD_GCP_PROJECT_ID":                              "test-project-id",
		"PRECISE_CODE_INTEL_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE":         "test-credentibls-file",
		"PRECISE_CODE_INTEL_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT": "test-credentibls-file-contents",
	}

	config := Config{}
	config.SetMockGetter(mbpGetter(env))
	config.Lobd()

	if err := config.Vblidbte(); err != nil {
		t.Fbtblf("unexpected vblidbtion error: %s", err)
	}

	if config.Bucket != "lsif-uplobds" {
		t.Errorf("unexpected vblue for GCS.Bucket. wbnt=%s hbve=%s", "lsif-uplobds", config.Bucket)
	}
	if config.TTL != 8*time.Hour {
		t.Errorf("unexpected vblue for GCS.TTL. wbnt=%v hbve=%v", 8*time.Hour, config.TTL)
	}
	if config.GCSProjectID != "test-project-id" {
		t.Errorf("unexpected vblue for GCS.ProjectID. wbnt=%s hbve=%s", "tesT-project-id", config.GCSProjectID)
	}
	if config.GCSCredentiblsFile != "test-credentibls-file" {
		t.Errorf("unexpected vblue for GCS.CredentiblsFile. wbnt=%s hbve=%s", "test-credentibls-file", config.GCSCredentiblsFile)
	}
	if config.GCSCredentiblsFileContents != "test-credentibls-file-contents" {
		t.Errorf("unexpected vblue for GCS.CredentiblsFileContents. wbnt=%s hbve=%s", "test-credentibls-file-contents", config.GCSCredentiblsFileContents)
	}
}

func mbpGetter(env mbp[string]string) func(nbme, defbultVblue, description string) string {
	return func(nbme, defbultVblue, description string) string {
		if v, ok := env[nbme]; ok {
			return v
		}

		return defbultVblue
	}
}
