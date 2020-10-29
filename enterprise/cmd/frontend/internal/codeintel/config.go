package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	BundleManagerURL            string
	HunkCacheSize               int
	BackgroundTaskInterval      time.Duration
	IndexBatchSize              int
	MinimumTimeSinceLastEnqueue time.Duration
	MinimumSearchCount          int
	MinimumSearchRatio          int
	MinimumPreciseCount         int
	UploadTimeout               time.Duration
	DataTTL                     time.Duration
}

var config = &Config{}

func init() {
	config.BundleManagerURL = config.Get("PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL", "", "TODO")
	config.HunkCacheSize = config.GetInt("PRECISE_CODE_INTEL_HUNK_CACHE_SIZE", "1000", "TODO")
	config.BackgroundTaskInterval = config.GetInterval("PRECISE_CODE_INTEL_BACKGROUND_TASK_INTERVAL", "1m", "TODO")
	config.IndexBatchSize = config.GetInt("PRECISE_CODE_INTEL_INDEX_BATCH_SIZE", "100", "TODO")
	config.MinimumTimeSinceLastEnqueue = config.GetInterval("PRECISE_CODE_INTEL_MINIMUM_TIME_SINCE_LAST_ENQUEUE", "24h", "TODO")
	config.MinimumSearchCount = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_COUNT", "50", "TODO")
	config.MinimumSearchRatio = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_SEARCH_RATIO", "50", "TODO")
	config.MinimumPreciseCount = config.GetInt("PRECISE_CODE_INTEL_MINIMUM_PRECISE_COUNT", "1", "TODO")
	config.UploadTimeout = config.GetInterval("PRECISE_CODE_INTEL_UPLOAD_TIMEOUT", "24h", "TODO")
	config.DataTTL = config.GetInterval("PRECISE_CODE_INTEL_DATA_TTL", "720h", "TODO")
}

// func (c *Config) APIWorkerOptions(transport http.RoundTripper) apiworker.Options {
// 	return apiworker.Options{
// 		// QueueName:          c.QueueName,
// 		// HeartbeatInterval:  c.HeartbeatInterval,
// 		// WorkerOptions:      c.WorkerOptions(),
// 		// FirecrackerOptions: c.FirecrackerOptions(),
// 		// ResourceOptions:    c.ResourceOptions(),
// 		// GitServicePath:     "/.executors/git",
// 		// ClientOptions:      c.ClientOptions(transport),
// 	}
// }
