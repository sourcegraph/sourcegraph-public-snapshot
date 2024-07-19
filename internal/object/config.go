package object

import (
	"strings"
	"time"
)

// StorageConfig captures all parameters required for instanciating an uploadstore.
// This struct needs to be passed in in full, there will be no `Load` call.
type StorageConfig struct {
	Backend      string
	ManageBucket bool
	Bucket       string
	TTL          time.Duration
	S3           S3Config
	GCS          GCSConfig
}

func normalizeConfig(t StorageConfig) StorageConfig {
	o := t
	// Normalize the backend name.
	o.Backend = strings.ToLower(o.Backend)

	if o.Backend == "blobstore" {
		o.S3.IsBlobstore = true

		// No manual provisioning on blobstore.
		o.ManageBucket = true

		// No subdomains on built-in blobstore.
		o.S3.UsePathStyle = true
	}
	return o
}
