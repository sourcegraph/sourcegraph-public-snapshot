package buildkite

// const cachePluginName = "gencer/cache#v2.4.10"

const cachePluginName = "jhchabran/cache#a006be99a6d5bfbab177ddbf8c988bdce3bd2a0e"

// CacheConfig represents the configuration data for https://github.com/gencer/cache-buildkite-plugin
type CacheConfigPayload struct {
	ID          string   `json:"id"`
	Backend     string   `json:"backend"`
	Key         string   `json:"key"`
	RestoreKeys []string `json:"restore_keys"`
	Compress    bool     `json:"compress,omitempty"`
	TarBall     struct {
		Path string `json:"path,omitempty"`
		Max  int    `json:"max,omitempty"`
	} `json:"tarball,omitempty"`
	Paths []string             `json:"paths"`
	S3    CacheConfigS3Payload `json:"s3"`
}

type CacheConfigS3Payload struct {
	Profile  string `json:"profile,omitempty"`
	Bucket   string `json:"bucket"`
	Class    string `json:"class,omitempty"`
	Args     string `json:"args,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	Region   string `json:"region,omitempty"`
}

type CacheOptions struct {
	ID          string
	Key         string
	RestoreKeys []string
	Paths       []string
}

func Cache(opts *CacheOptions) StepOpt {
	return Plugin(cachePluginName, CacheConfigPayload{
		ID:          opts.ID,
		Key:         opts.Key,
		RestoreKeys: opts.RestoreKeys,
		Paths:       opts.Paths,
		Backend:     "s3",
		S3: CacheConfigS3Payload{
			Bucket:   "sourcegraph_buildkite_cache",
			Profile:  "buildkite",
			Endpoint: "https://storage.googleapis.com",
			Region:   "us-central1",
		},
	})
}
