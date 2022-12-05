package shared

import (
	"os"
	"path/filepath"

	sglog "github.com/sourcegraph/log"
)

func maybeBlobstore(logger sglog.Logger) []string {
	if os.Getenv("DISABLE_BLOBSTORE") != "" {
		logger.Warn("running with blobstore disabled")
		return []string{""}
	}

	// Point at local blobstore endpoint.
	SetDefaultEnv("PRECISE_CODE_INTEL_UPLOAD_AWS_ENDPOINT", "http://127.0.0.1:9000")
	SetDefaultEnv("PRECISE_CODE_INTEL_UPLOAD_BACKEND", "blobstore")

	// cmd/server-specific blobstore env vars
	// Note: SetDefaultEnv should be called only once for the same env var, so
	// ensure these do not conflict with the default list set below.
	dataDir := filepath.Join(os.Getenv("DATA_DIR"), "blobstore")
	SetDefaultEnv("JCLOUDS_FILESYSTEM_BASEDIR", dataDir)

	// Default blobstore env vars copied from the Dockerfile
	// https://github.com/sourcegraph/sourcegraph/blob/main/docker-images/blobstore/Dockerfile
	SetDefaultEnv("LOG_LEVEL", "info")
	SetDefaultEnv("S3PROXY_AUTHORIZATION", "none")
	SetDefaultEnv("S3PROXY_ENDPOINT", "http://0.0.0.0:9000")
	SetDefaultEnv("S3PROXY_IDENTITY", "local-identity")
	SetDefaultEnv("S3PROXY_CREDENTIAL", "local-credential")
	SetDefaultEnv("S3PROXY_VIRTUALHOST", "")
	SetDefaultEnv("S3PROXY_CORS_ALLOW_ALL", "false")
	SetDefaultEnv("S3PROXY_CORS_ALLOW_ORIGINS", "")
	SetDefaultEnv("S3PROXY_CORS_ALLOW_METHODS", "")
	SetDefaultEnv("S3PROXY_CORS_ALLOW_HEADERS", "")
	SetDefaultEnv("S3PROXY_IGNORE_UNKNOWN_HEADERS", "false")
	SetDefaultEnv("S3PROXY_ENCRYPTED_BLOBSTORE", "")
	SetDefaultEnv("S3PROXY_ENCRYPTED_BLOBSTORE_PASSWORD", "")
	SetDefaultEnv("S3PROXY_ENCRYPTED_BLOBSTORE_SALT", "")
	SetDefaultEnv("S3PROXY_V4_MAX_NON_CHUNKED_REQ_SIZE", "33554432")
	SetDefaultEnv("JCLOUDS_PROVIDER", "filesystem")
	SetDefaultEnv("JCLOUDS_ENDPOINT", "")
	SetDefaultEnv("JCLOUDS_REGION", "")
	SetDefaultEnv("JCLOUDS_REGIONS", "us-east-1")
	SetDefaultEnv("JCLOUDS_IDENTITY", "remote-identity")
	SetDefaultEnv("JCLOUDS_CREDENTIAL", "remote-credential")
	SetDefaultEnv("JCLOUDS_KEYSTONE_VERSION", "")
	SetDefaultEnv("JCLOUDS_KEYSTONE_SCOPE", "")
	SetDefaultEnv("JCLOUDS_KEYSTONE_PROJECT_DOMAIN_NAME", "")
	// SetDefaultEnv("JCLOUDS_FILESYSTEM_BASEDIR", dataDir) // overridden above; here for equality with Dockerfile

	// Configure blobstore service
	procline := `blobstore: /opt/s3proxy/run-docker-container.sh >> /var/opt/sourcegraph/blobstore.log 2>&1`
	return []string{procline}
}
