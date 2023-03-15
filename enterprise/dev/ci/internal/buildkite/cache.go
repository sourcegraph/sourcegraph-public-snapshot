package buildkite

// Follow-up to INC-101, this fork of 'gencer/cache#v2.4.10' uses bsdtar instead of tar.
const cachePluginName = "https://github.com/sourcegraph/cache-buildkite-plugin.git#master"

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
	PR    string               `json:"pr,omitempty"`
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
	ID                string
	Key               string
	RestoreKeys       []string
	Paths             []string
	Compress          bool
	IgnorePullRequest bool
}

func Cache(opts *CacheOptions) StepOpt {
	var cachePR string
	if opts.IgnorePullRequest {
		cachePR = "off"
	}
	return flattenStepOpts(
		// Overrides the aws command configuration to use the buildkite cache
		// configuration instead.
		Env("AWS_CONFIG_FILE", "/buildkite/.aws/config"),
		Env("AWS_SHARED_CREDENTIALS_FILE", "/buildkite/.aws/credentials"),
		Plugin(cachePluginName, CacheConfigPayload{
			ID:          opts.ID,
			Key:         opts.Key,
			RestoreKeys: opts.RestoreKeys,
			Paths:       opts.Paths,
			Compress:    opts.Compress,
			Backend:     "s3",
			PR:          cachePR,
			S3: CacheConfigS3Payload{
				Bucket:   "sourcegraph_buildkite_cache",
				Profile:  "buildkite",
				Endpoint: "https://storage.googleapis.com",
				Region:   "us-central1",
			},
		}),
	)
}
