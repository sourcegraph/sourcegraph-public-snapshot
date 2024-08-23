package client

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ApiKey string

	OpsGenieAPIURL ApiUrl

	apiUrl string

	ProxyConfiguration *ProxyConfiguration

	RequestTimeout time.Duration

	HttpClient *http.Client

	Backoff retryablehttp.Backoff

	RetryPolicy retryablehttp.CheckRetry

	RetryCount int

	LogLevel logrus.Level

	Logger logrus.FieldLogger
}

type ApiUrl string

const (
	API_URL         ApiUrl = "api.opsgenie.com"
	API_URL_EU      ApiUrl = "api.eu.opsgenie.com"
	API_URL_SANDBOX ApiUrl = "api.sandbox.opsgenie.com"
)

func (conf Config) Validate() error {

	if conf.ApiKey == "" {
		return errors.New("API key cannot be blank.")
	}
	if conf.RetryCount < 0 {
		return errors.New("Retry count cannot be less than 1.")
	}
	return nil
}

func Default() *Config {
	return &Config{}
}

type ProxyConfiguration struct {
	Username string
	Password string
	Host     string
	Protocol Protocol
	Port     int
}

type Protocol string

const (
	Http   Protocol = "http"
	Https  Protocol = "https"
	Socks5 Protocol = "socks5"
)

func (conf *Config) ConfigureLogLevel(level string) {
	var logLevel logrus.Level
	switch level {
	case "panic":
		logLevel = logrus.PanicLevel
	case "fatal":
		logLevel = logrus.FatalLevel
	case "error":
		logLevel = logrus.ErrorLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "debug":
		logLevel = logrus.DebugLevel
	case "trace":
		logLevel = logrus.TraceLevel
	default:
		logLevel = logrus.InfoLevel
	}
	conf.LogLevel = logLevel
}
