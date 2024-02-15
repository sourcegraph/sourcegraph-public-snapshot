package deploy

import (
	"github.com/sourcegraph/sourcegraph/internal/env"
)

// BlobstoreEndpoint returns the default blobstore endpoint that should be used for this deployment
// type.
func BlobstoreDefaultEndpoint() string {
	if IsDeployTypeSingleDockerContainer(Type()) {
		return "http://127.0.0.1:9000"
	}
	return "http://blobstore:9000"
}

// BlobstoreHostPort returns the host/port that should be listened on for this deployment type.
func BlobstoreHostPort() (string, string) {
	if env.InsecureDev || IsDeployTypeSingleDockerContainer(Type()) {
		return "127.0.0.1", "9000"
	}
	return "", "9000"
}
