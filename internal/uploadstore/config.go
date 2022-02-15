package uploadstore

import (
	"strings"
	"time"
)

// Config captures all parameters required for instanciating an uploadstore.
// This struct needs to be passed in in full, there will be no `Load` call.
type Config struct {
	Backend      string
	ManageBucket bool
	Bucket       string
	TTL          time.Duration
	S3           S3Config
	GCS          GCSConfig
}

func normalizeConfig(t Config) Config {
	o := t
	// Normalize the backend name.
	o.Backend = strings.ToLower(o.Backend)

	if o.Backend == "minio" {
		// No manual provisioning on minIO.
		o.ManageBucket = true
	}
	return o
}
