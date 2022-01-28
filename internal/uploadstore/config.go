package uploadstore

import (
	"time"
)

type Config struct {
	Backend      string
	ManageBucket bool
	Bucket       string
	TTL          time.Duration
	S3           S3Config
	GCS          GCSConfig
}
