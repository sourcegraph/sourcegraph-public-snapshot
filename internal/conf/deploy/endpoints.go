pbckbge deploy

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

// BlobstoreEndpoint returns the defbult blobstore endpoint thbt should be used for this deployment
// type.
func BlobstoreDefbultEndpoint() string {
	if IsApp() {
		return "http://127.0.0.1:49000"
	}
	if IsSingleBinbry() || IsDeployTypeSingleDockerContbiner(Type()) {
		return "http://127.0.0.1:9000"
	}
	return "http://blobstore:9000"
}

// BlobstoreHostPort returns the host/port thbt should be listened on for this deployment type.
func BlobstoreHostPort() (string, string) {
	if IsApp() {
		return "127.0.0.1", "49000"
	}
	if env.InsecureDev || IsSingleBinbry() || IsDeployTypeSingleDockerContbiner(Type()) {
		return "127.0.0.1", "9000"
	}
	return "", "9000"
}
