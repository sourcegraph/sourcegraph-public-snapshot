// Package config contains all knobs and defaults used to configure features of
// OpenFGA when running as a standalone server.
package config

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/spf13/viper"

	"github.com/openfga/openfga/pkg/logger"
)

const (
	DefaultMaxRPCMessageSizeInBytes         = 512 * 1_204 // 512 KB
	DefaultMaxTuplesPerWrite                = 100
	DefaultMaxTypesPerAuthorizationModel    = 100
	DefaultMaxAuthorizationModelSizeInBytes = 256 * 1_024
	DefaultMaxAuthorizationModelCacheSize   = 100000
	DefaultChangelogHorizonOffset           = 0
	DefaultResolveNodeLimit                 = 25
	DefaultResolveNodeBreadthLimit          = 100
	DefaultListObjectsDeadline              = 3 * time.Second
	DefaultListObjectsMaxResults            = 1000
	DefaultMaxConcurrentReadsForCheck       = math.MaxUint32
	DefaultMaxConcurrentReadsForListObjects = math.MaxUint32
	DefaultListUsersDeadline                = 3 * time.Second
	DefaultListUsersMaxResults              = 1000
	DefaultMaxConcurrentReadsForListUsers   = math.MaxUint32

	DefaultWriteContextByteLimit = 32 * 1_024 // 32KB
	DefaultCheckQueryCacheLimit  = 10000
	DefaultCheckQueryCacheTTL    = 10 * time.Second
	DefaultCheckQueryCacheEnable = false

	// Care should be taken here - decreasing can cause API compatibility problems with Conditions.
	DefaultMaxConditionEvaluationCost = 100
	DefaultInterruptCheckFrequency    = 100

	DefaultCheckDispatchThrottlingEnabled          = false
	DefaultCheckDispatchThrottlingFrequency        = 10 * time.Microsecond
	DefaultCheckDispatchThrottlingDefaultThreshold = 100
	DefaultCheckDispatchThrottlingMaxThreshold     = 0 // 0 means use the default threshold as max

	DefaultListObjectsDispatchThrottlingEnabled          = false
	DefaultListObjectsDispatchThrottlingFrequency        = 10 * time.Microsecond
	DefaultListObjectsDispatchThrottlingDefaultThreshold = 100
	DefaultListObjectsDispatchThrottlingMaxThreshold     = 0 // 0 means use the default threshold as max

	DefaultRequestTimeout = 3 * time.Second

	additionalUpstreamTimeout = 3 * time.Second
)

type DatastoreMetricsConfig struct {
	// Enabled enables export of the Datastore metrics.
	Enabled bool
}

// DatastoreConfig defines OpenFGA server configurations for datastore specific settings.
type DatastoreConfig struct {
	// Engine is the datastore engine to use (e.g. 'memory', 'postgres', 'mysql')
	Engine   string
	URI      string `json:"-"` // private field, won't be logged
	Username string
	Password string `json:"-"` // private field, won't be logged

	// MaxCacheSize is the maximum number of authorization models that will be cached in memory.
	MaxCacheSize int

	// MaxOpenConns is the maximum number of open connections to the database.
	MaxOpenConns int

	// MaxIdleConns is the maximum number of connections to the datastore in the idle connection
	// pool.
	MaxIdleConns int

	// ConnMaxIdleTime is the maximum amount of time a connection to the datastore may be idle.
	ConnMaxIdleTime time.Duration

	// ConnMaxLifetime is the maximum amount of time a connection to the datastore may be reused.
	ConnMaxLifetime time.Duration

	// Metrics is configuration for the Datastore metrics.
	Metrics DatastoreMetricsConfig
}

// GRPCConfig defines OpenFGA server configurations for grpc server specific settings.
type GRPCConfig struct {
	Addr string
	TLS  *TLSConfig
}

// HTTPConfig defines OpenFGA server configurations for HTTP server specific settings.
type HTTPConfig struct {
	Enabled bool
	Addr    string
	TLS     *TLSConfig

	// UpstreamTimeout is the timeout duration for proxying HTTP requests upstream
	// to the grpc endpoint. It cannot be smaller than Config.ListObjectsDeadline.
	UpstreamTimeout time.Duration

	CORSAllowedOrigins []string
	CORSAllowedHeaders []string
}

// TLSConfig defines configuration specific to Transport Layer Security (TLS) settings.
type TLSConfig struct {
	Enabled  bool
	CertPath string `mapstructure:"cert"`
	KeyPath  string `mapstructure:"key"`
}

// AuthnConfig defines OpenFGA server configurations for authentication specific settings.
type AuthnConfig struct {

	// Method is the authentication method that should be enforced (e.g. 'none', 'preshared',
	// 'oidc')
	Method                   string
	*AuthnOIDCConfig         `mapstructure:"oidc"`
	*AuthnPresharedKeyConfig `mapstructure:"preshared"`
}

// AuthnOIDCConfig defines configurations for the 'oidc' method of authentication.
type AuthnOIDCConfig struct {
	Issuer        string
	IssuerAliases []string
	Audience      string
}

// AuthnPresharedKeyConfig defines configurations for the 'preshared' method of authentication.
type AuthnPresharedKeyConfig struct {
	// Keys define the preshared keys to verify authn tokens against.
	Keys []string `json:"-"` // private field, won't be logged
}

// LogConfig defines OpenFGA server configurations for log specific settings. For production we
// recommend using the 'json' log format.
type LogConfig struct {
	// Format is the log format to use in the log output (e.g. 'text' or 'json')
	Format string

	// Level is the log level to use in the log output (e.g. 'none', 'debug', or 'info')
	Level string

	// Format of the timestamp in the log output (e.g. 'Unix'(default) or 'ISO8601')
	TimestampFormat string
}

type TraceConfig struct {
	Enabled     bool
	OTLP        OTLPTraceConfig `mapstructure:"otlp"`
	SampleRatio float64
	ServiceName string
}

type OTLPTraceConfig struct {
	Endpoint string
	TLS      OTLPTraceTLSConfig
}

type OTLPTraceTLSConfig struct {
	Enabled bool
}

// PlaygroundConfig defines OpenFGA server configurations for the Playground specific settings.
type PlaygroundConfig struct {
	Enabled bool
	Port    int
}

// ProfilerConfig defines server configurations specific to pprof profiling.
type ProfilerConfig struct {
	Enabled bool
	Addr    string
}

// MetricConfig defines configurations for serving custom metrics from OpenFGA.
type MetricConfig struct {
	Enabled             bool
	Addr                string
	EnableRPCHistograms bool
}

// CheckQueryCache defines configuration for caching when resolving check.
type CheckQueryCache struct {
	Enabled bool
	Limit   uint32 // (in items)
	TTL     time.Duration
}

// DispatchThrottlingConfig defines configurations for dispatch throttling.
type DispatchThrottlingConfig struct {
	Enabled      bool
	Frequency    time.Duration
	Threshold    uint32
	MaxThreshold uint32
}

type Config struct {
	// If you change any of these settings, please update the documentation at
	// https://github.com/openfga/openfga.dev/blob/main/docs/content/intro/setup-openfga.mdx

	// ListObjectsDeadline defines the maximum amount of time to accumulate ListObjects results
	// before the server will respond. This is to protect the server from misuse of the
	// ListObjects endpoints. It cannot be larger than HTTPConfig.UpstreamTimeout.
	ListObjectsDeadline time.Duration

	// ListObjectsMaxResults defines the maximum number of results to accumulate
	// before the non-streaming ListObjects API will respond to the client.
	// This is to protect the server from misuse of the ListObjects endpoints.
	ListObjectsMaxResults uint32

	// ListUsersDeadline defines the maximum amount of time to accumulate ListUsers results
	// before the server will respond. This is to protect the server from misuse of the
	// ListUsers endpoints. It cannot be larger than the configured server's request timeout (RequestTimeout or HTTPConfig.UpstreamTimeout).
	ListUsersDeadline time.Duration

	// ListUsersMaxResults defines the maximum number of results to accumulate
	// before the non-streaming ListUsers API will respond to the client.
	// This is to protect the server from misuse of the ListUsers endpoints.
	ListUsersMaxResults uint32

	// MaxTuplesPerWrite defines the maximum number of tuples per Write endpoint.
	MaxTuplesPerWrite int

	// MaxTypesPerAuthorizationModel defines the maximum number of type definitions per
	// authorization model for the WriteAuthorizationModel endpoint.
	MaxTypesPerAuthorizationModel int

	// MaxAuthorizationModelSizeInBytes defines the maximum size in bytes allowed for
	// persisting an Authorization Model.
	MaxAuthorizationModelSizeInBytes int

	// MaxConcurrentReadsForListObjects defines the maximum number of concurrent database reads
	// allowed in ListObjects queries
	MaxConcurrentReadsForListObjects uint32

	// MaxConcurrentReadsForCheck defines the maximum number of concurrent database reads allowed in
	// Check queries
	MaxConcurrentReadsForCheck uint32

	// MaxConcurrentReadsForListUsers defines the maximum number of concurrent database reads
	// allowed in ListUsers queries
	MaxConcurrentReadsForListUsers uint32

	// MaxConditionEvaluationCost defines the maximum cost for CEL condition evaluation before a request returns an error
	MaxConditionEvaluationCost uint64

	// ChangelogHorizonOffset is an offset in minutes from the current time. Changes that occur
	// after this offset will not be included in the response of ReadChanges.
	ChangelogHorizonOffset int

	// Experimentals is a list of the experimental features to enable in the OpenFGA server.
	Experimentals []string

	// ResolveNodeLimit indicates how deeply nested an authorization model can be before a query
	// errors out.
	ResolveNodeLimit uint32

	// ResolveNodeBreadthLimit indicates how many nodes on a given level can be evaluated
	// concurrently in a query
	ResolveNodeBreadthLimit uint32

	// RequestTimeout configures request timeout.  If both HTTP upstream timeout and request timeout are specified,
	// request timeout will be prioritized
	RequestTimeout time.Duration

	Datastore                     DatastoreConfig
	GRPC                          GRPCConfig
	HTTP                          HTTPConfig
	Authn                         AuthnConfig
	Log                           LogConfig
	Trace                         TraceConfig
	Playground                    PlaygroundConfig
	Profiler                      ProfilerConfig
	Metrics                       MetricConfig
	CheckQueryCache               CheckQueryCache
	DispatchThrottling            DispatchThrottlingConfig
	CheckDispatchThrottling       DispatchThrottlingConfig
	ListObjectsDispatchThrottling DispatchThrottlingConfig

	RequestDurationDatastoreQueryCountBuckets []string
	RequestDurationDispatchCountBuckets       []string
}

func (cfg *Config) Verify() error {
	configuredTimeout := DefaultContextTimeout(cfg)

	if cfg.ListObjectsDeadline > configuredTimeout {
		return fmt.Errorf(
			"configured request timeout (%s) cannot be lower than 'listObjectsDeadline' config (%s)",
			configuredTimeout,
			cfg.ListObjectsDeadline,
		)
	}
	if cfg.ListUsersDeadline > configuredTimeout {
		return fmt.Errorf(
			"configured request timeout (%s) cannot be lower than 'listUsersDeadline' config (%s)",
			configuredTimeout,
			cfg.ListUsersDeadline,
		)
	}

	if cfg.MaxConcurrentReadsForListUsers == 0 {
		return fmt.Errorf("config 'maxConcurrentReadsForListUsers' cannot be 0")
	}

	if cfg.Log.Format != "text" && cfg.Log.Format != "json" {
		return fmt.Errorf("config 'log.format' must be one of ['text', 'json']")
	}

	if cfg.Log.Level != "none" &&
		cfg.Log.Level != "debug" &&
		cfg.Log.Level != "info" &&
		cfg.Log.Level != "warn" &&
		cfg.Log.Level != "error" &&
		cfg.Log.Level != "panic" &&
		cfg.Log.Level != "fatal" {
		return fmt.Errorf(
			"config 'log.level' must be one of ['none', 'debug', 'info', 'warn', 'error', 'panic', 'fatal']",
		)
	}

	if cfg.Log.TimestampFormat != "Unix" && cfg.Log.TimestampFormat != "ISO8601" {
		return fmt.Errorf("config 'log.TimestampFormat' must be one of ['Unix', 'ISO8601']")
	}

	if cfg.Playground.Enabled {
		if !cfg.HTTP.Enabled {
			return errors.New("the HTTP server must be enabled to run the openfga playground")
		}

		if !(cfg.Authn.Method == "none" || cfg.Authn.Method == "preshared") {
			return errors.New("the playground only supports authn methods 'none' and 'preshared'")
		}
	}

	if cfg.HTTP.TLS.Enabled {
		if cfg.HTTP.TLS.CertPath == "" || cfg.HTTP.TLS.KeyPath == "" {
			return errors.New("'http.tls.cert' and 'http.tls.key' configs must be set")
		}
	}

	if cfg.GRPC.TLS.Enabled {
		if cfg.GRPC.TLS.CertPath == "" || cfg.GRPC.TLS.KeyPath == "" {
			return errors.New("'grpc.tls.cert' and 'grpc.tls.key' configs must be set")
		}
	}

	if len(cfg.RequestDurationDatastoreQueryCountBuckets) == 0 {
		return errors.New("request duration datastore query count buckets must not be empty")
	}
	for _, val := range cfg.RequestDurationDatastoreQueryCountBuckets {
		valInt, err := strconv.Atoi(val)
		if err != nil || valInt < 0 {
			return errors.New(
				"request duration datastore query count bucket items must be non-negative integer",
			)
		}
	}

	if len(cfg.RequestDurationDispatchCountBuckets) == 0 {
		return errors.New("request duration datastore dispatch count buckets must not be empty")
	}
	for _, val := range cfg.RequestDurationDispatchCountBuckets {
		valInt, err := strconv.Atoi(val)
		if err != nil || valInt < 0 {
			return errors.New(
				"request duration dispatch count bucket items must be non-negative integer",
			)
		}
	}

	// Tha validation ensures we are picking the right values for Check dispatch throttling
	err := cfg.VerifyCheckDispatchThrottlingConfig()
	if err != nil {
		return err
	}

	if cfg.ListObjectsDispatchThrottling.Enabled {
		if cfg.ListObjectsDispatchThrottling.Frequency <= 0 {
			return errors.New("'listObjectsDispatchThrottling.frequency' must be non-negative time duration")
		}
		if cfg.ListObjectsDispatchThrottling.Threshold <= 0 {
			return errors.New("'listObjectsDispatchThrottling.threshold' must be non-negative integer")
		}
		if cfg.ListObjectsDispatchThrottling.MaxThreshold != 0 && cfg.ListObjectsDispatchThrottling.Threshold > cfg.ListObjectsDispatchThrottling.MaxThreshold {
			return errors.New("'listObjectsDispatchThrottling.threshold' must be less than or equal to 'listObjectsDispatchThrottling.maxThreshold'")
		}
	}

	if cfg.RequestTimeout < 0 {
		return errors.New("requestTimeout must be a non-negative time duration")
	}

	if cfg.RequestTimeout == 0 && cfg.HTTP.Enabled && cfg.HTTP.UpstreamTimeout < 0 {
		return errors.New("http.upstreamTimeout must be a non-negative time duration")
	}

	if cfg.ListObjectsDeadline < 0 {
		return errors.New("listObjectsDeadline must be non-negative time duration")
	}

	if cfg.MaxConditionEvaluationCost < 100 {
		return errors.New("maxConditionsEvaluationCosts less than 100 can cause API compatibility problems with Conditions")
	}

	return nil
}

// DefaultContextTimeout returns the runtime DefaultContextTimeout.
// If requestTimeout > 0, we should let the middleware take care of the timeout and the
// runtime.DefaultContextTimeout is used as last resort.
// Otherwise, use the http upstream timeout if http is enabled.
func DefaultContextTimeout(config *Config) time.Duration {
	if config.RequestTimeout > 0 {
		return config.RequestTimeout + additionalUpstreamTimeout
	}
	if config.HTTP.Enabled && config.HTTP.UpstreamTimeout > 0 {
		return config.HTTP.UpstreamTimeout
	}
	return 0
}

// GetCheckDispatchThrottlingConfig is used to get the DispatchThrottlingConfig value for Check. To avoid breaking change
// we will try to get the value from config.DispatchThrottling but override it with config.CheckDispatchThrottling if
// a non-zero value exists there.
func GetCheckDispatchThrottlingConfig(logger logger.Logger, config *Config) DispatchThrottlingConfig {
	checkDispatchThrottlingEnabled := config.CheckDispatchThrottling.Enabled
	checkDispatchThrottlingFrequency := config.CheckDispatchThrottling.Frequency
	checkDispatchThrottlingDefaultThreshold := config.CheckDispatchThrottling.Threshold
	checkDispatchThrottlingMaxThreshold := config.CheckDispatchThrottling.MaxThreshold

	if viper.IsSet("dispatchThrottling.enabled") && !viper.IsSet("checkDispatchThrottling.enabled") {
		if logger != nil {
			logger.Warn("'dispatchThrottling.enabled' is deprecated. Please use 'checkDispatchThrottling.enabled'")
		}
		checkDispatchThrottlingEnabled = config.DispatchThrottling.Enabled
	}
	if viper.IsSet("dispatchThrottling.frequency") && !viper.IsSet("checkDispatchThrottling.frequency") {
		if logger != nil {
			logger.Warn("'dispatchThrottling.frequency' is deprecated. Please use 'checkDispatchThrottling.frequency'")
		}
		checkDispatchThrottlingFrequency = config.DispatchThrottling.Frequency
	}
	if viper.IsSet("dispatchThrottling.threshold") && !viper.IsSet("checkDispatchThrottling.threshold") {
		if logger != nil {
			logger.Warn("'dispatchThrottling.threshold' is deprecated. Please use 'checkDispatchThrottling.threshold'")
		}
		checkDispatchThrottlingDefaultThreshold = config.DispatchThrottling.Threshold
	}
	if viper.IsSet("dispatchThrottling.maxThreshold") && !viper.IsSet("checkDispatchThrottling.maxThreshold") {
		if logger != nil {
			logger.Warn("'dispatchThrottling.maxThreshold' is deprecated. Please use 'checkDispatchThrottling.maxThreshold'")
		}
		checkDispatchThrottlingMaxThreshold = config.DispatchThrottling.MaxThreshold
	}

	return DispatchThrottlingConfig{
		Enabled:      checkDispatchThrottlingEnabled,
		Frequency:    checkDispatchThrottlingFrequency,
		Threshold:    checkDispatchThrottlingDefaultThreshold,
		MaxThreshold: checkDispatchThrottlingMaxThreshold,
	}
}

// VerifyCheckDispatchThrottlingConfig ensures GetCheckDispatchThrottlingConfig is called so that the right values are verified.
func (cfg *Config) VerifyCheckDispatchThrottlingConfig() error {
	checkDispatchThrottlingConfig := GetCheckDispatchThrottlingConfig(nil, cfg)
	if checkDispatchThrottlingConfig.Enabled {
		if checkDispatchThrottlingConfig.Frequency <= 0 {
			return errors.New("'dispatchThrottling.frequency (deprecated)' or 'checkDispatchThrottling.frequency' must be non-negative time duration")
		}
		if checkDispatchThrottlingConfig.Threshold <= 0 {
			return errors.New("'dispatchThrottling.threshold (deprecated)' or 'checkDispatchThrottling.threshold' must be non-negative integer")
		}
		if checkDispatchThrottlingConfig.MaxThreshold != 0 && checkDispatchThrottlingConfig.Threshold > checkDispatchThrottlingConfig.MaxThreshold {
			return errors.New("'dispatchThrottling.threshold (deprecated)' or 'checkDispatchThrottling.threshold' must be less than or equal to 'dispatchThrottling.maxThreshold (deprecated)' or 'checkDispatchThrottling.maxThreshold' respectively")
		}
	}
	return nil
}

// MaxConditionEvaluationCost ensures a safe value for CEL evaluation cost.
func MaxConditionEvaluationCost() uint64 {
	return max(DefaultMaxConditionEvaluationCost, viper.GetUint64("maxConditionEvaluationCost"))
}

// DefaultConfig is the OpenFGA server default configurations.
func DefaultConfig() *Config {
	return &Config{
		MaxTuplesPerWrite:                         DefaultMaxTuplesPerWrite,
		MaxTypesPerAuthorizationModel:             DefaultMaxTypesPerAuthorizationModel,
		MaxAuthorizationModelSizeInBytes:          DefaultMaxAuthorizationModelSizeInBytes,
		MaxConcurrentReadsForCheck:                DefaultMaxConcurrentReadsForCheck,
		MaxConcurrentReadsForListObjects:          DefaultMaxConcurrentReadsForListObjects,
		MaxConcurrentReadsForListUsers:            DefaultMaxConcurrentReadsForListUsers,
		MaxConditionEvaluationCost:                DefaultMaxConditionEvaluationCost,
		ChangelogHorizonOffset:                    DefaultChangelogHorizonOffset,
		ResolveNodeLimit:                          DefaultResolveNodeLimit,
		ResolveNodeBreadthLimit:                   DefaultResolveNodeBreadthLimit,
		Experimentals:                             []string{},
		ListObjectsDeadline:                       DefaultListObjectsDeadline,
		ListObjectsMaxResults:                     DefaultListObjectsMaxResults,
		ListUsersMaxResults:                       DefaultListUsersMaxResults,
		ListUsersDeadline:                         DefaultListUsersDeadline,
		RequestDurationDatastoreQueryCountBuckets: []string{"50", "200"},
		RequestDurationDispatchCountBuckets:       []string{"50", "200"},
		Datastore: DatastoreConfig{
			Engine:       "memory",
			MaxCacheSize: DefaultMaxAuthorizationModelCacheSize,
			MaxIdleConns: 10,
			MaxOpenConns: 30,
		},
		GRPC: GRPCConfig{
			Addr: "0.0.0.0:8081",
			TLS:  &TLSConfig{Enabled: false},
		},
		HTTP: HTTPConfig{
			Enabled:            true,
			Addr:               "0.0.0.0:8080",
			TLS:                &TLSConfig{Enabled: false},
			UpstreamTimeout:    5 * time.Second,
			CORSAllowedOrigins: []string{"*"},
			CORSAllowedHeaders: []string{"*"},
		},
		Authn: AuthnConfig{
			Method:                  "none",
			AuthnPresharedKeyConfig: &AuthnPresharedKeyConfig{},
			AuthnOIDCConfig:         &AuthnOIDCConfig{},
		},
		Log: LogConfig{
			Format:          "text",
			Level:           "info",
			TimestampFormat: "Unix",
		},
		Trace: TraceConfig{
			Enabled: false,
			OTLP: OTLPTraceConfig{
				Endpoint: "0.0.0.0:4317",
				TLS: OTLPTraceTLSConfig{
					Enabled: false,
				},
			},
			SampleRatio: 0.2,
			ServiceName: "openfga",
		},
		Playground: PlaygroundConfig{
			Enabled: true,
			Port:    3000,
		},
		Profiler: ProfilerConfig{
			Enabled: false,
			Addr:    ":3001",
		},
		Metrics: MetricConfig{
			Enabled:             true,
			Addr:                "0.0.0.0:2112",
			EnableRPCHistograms: false,
		},
		CheckQueryCache: CheckQueryCache{
			Enabled: DefaultCheckQueryCacheEnable,
			Limit:   DefaultCheckQueryCacheLimit,
			TTL:     DefaultCheckQueryCacheTTL,
		},
		DispatchThrottling: DispatchThrottlingConfig{
			Enabled:      DefaultCheckDispatchThrottlingEnabled,
			Frequency:    DefaultCheckDispatchThrottlingFrequency,
			Threshold:    DefaultCheckDispatchThrottlingDefaultThreshold,
			MaxThreshold: DefaultCheckDispatchThrottlingMaxThreshold,
		},
		CheckDispatchThrottling: DispatchThrottlingConfig{
			Enabled:      DefaultCheckDispatchThrottlingEnabled,
			Frequency:    DefaultCheckDispatchThrottlingFrequency,
			Threshold:    DefaultCheckDispatchThrottlingDefaultThreshold,
			MaxThreshold: DefaultCheckDispatchThrottlingMaxThreshold,
		},
		ListObjectsDispatchThrottling: DispatchThrottlingConfig{
			Enabled:      DefaultListObjectsDispatchThrottlingEnabled,
			Frequency:    DefaultListObjectsDispatchThrottlingFrequency,
			Threshold:    DefaultListObjectsDispatchThrottlingDefaultThreshold,
			MaxThreshold: DefaultListObjectsDispatchThrottlingMaxThreshold,
		},
		RequestTimeout: DefaultRequestTimeout,
	}
}

// MustDefaultConfig returns default server config with the playground, tracing and metrics turned off.
func MustDefaultConfig() *Config {
	config := DefaultConfig()

	config.Playground.Enabled = false
	config.Metrics.Enabled = false

	return config
}
