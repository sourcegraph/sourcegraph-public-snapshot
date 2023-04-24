package install

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

type ExternalService struct {
	// Type of code host, e.g. GITHUB.
	Kind string `yaml:"kind"`

	// Display name of external service, e.g. sourcegraph-test.
	DisplayName string `yaml:"displayName"`

	// Configuration for code host.
	Config Config `yaml:"config"`

	// Maximum retry attempts when cloning test repositories. Defaults to 5 retries.
	MaxRetries int `yaml:"maxRetries"`

	// Retry timeout in seconds. Defaults to 5 seconds
	RetryTimeoutSeconds int `yaml:"retryTimeoutSeconds"`

	// Delete code host when test is done. Defaults to true.
	DeleteWhenDone bool `yaml:"deleteWhenDone"`
}

// Config for different types of code hosts.
type Config struct {
	GitHub GitHub `yaml:"gitHub"`
}

// GitHub configuration parameters.
type GitHub struct {
	// URL used to access your GitHub instance, e.g. https://github.com.
	URL string `yaml:"url" json:"url"`

	// Auth token used to authenticate to GitHub instance. This should be provided via env var SRC_GITHUB_TOKEN.
	Token string `yaml:"token" json:"token"`

	// List of organizations.
	Orgs []string `yaml:"orgs" json:"orgs"`

	// List of repositories to pull.
	Repos []string `yaml:"repos" json:"repos"`
}

type Insight struct {
	Title          string           `yaml:"title"`
	DataSeries     []map[string]any `yaml:"dataSeries"`
	DeleteWhenDone bool             `yaml:"deleteWhenDone"`
}

type Executor struct {
	Enabled bool `yaml:"enabled"`
	Count   bool `yaml:"count"`
}

type Smtp struct {
	Enabled bool   `yaml:"enabled"`
	To      string `yaml:"to"`
}

type ValidationSpec struct {
	// Search queries used for validation testing, e.g. "repo:^github\\.com/gorilla/mux$ Router".
	SearchQuery []string `yaml:"searchQuery"`

	// External Service configuration.
	ExternalService ExternalService `yaml:"externalService"`

	// Insight used for validation testing.
	Insight Insight `yaml:"insight"`

	// Executor check configuration
	Executor Executor `yaml:"executor"`

	//Test SMTP configuration
	Smtp Smtp `yaml:"smtp"`
}

// DefaultConfig returns a default configuration to be used for testing.
func DefaultConfig() *ValidationSpec {
	return &ValidationSpec{
		SearchQuery: []string{
			"repo:^github.com/sourcegraph/src-cli$ config",
			"repo:^github.com/sourcegraph/src-cli$@4.0.0 config",
			"repo:^github.com/sourcegraph/src-cli$ type:symbol config",
		},
		ExternalService: ExternalService{
			Kind:        "GITHUB",
			DisplayName: "sourcegraph-test",
			Config: Config{
				GitHub: GitHub{
					URL:   "https://github.com",
					Token: "",
					Orgs:  []string{},
					Repos: []string{"sourcegraph/src-cli"},
				},
			},
			MaxRetries:          5,
			RetryTimeoutSeconds: 5,
			DeleteWhenDone:      true,
		},
		Insight: Insight{
			Title: "test insight",
			DataSeries: []map[string]any{
				{
					"query":           "lang:javascript",
					"label":           "javascript",
					"repositoryScope": "",
					"lineColor":       "#6495ED",
					"timeScopeUnit":   "MONTH",
					"timeScopeValue":  1,
				},
				{
					"query":           "lang:typescript",
					"label":           "typescript",
					"lineColor":       "#DE3163",
					"repositoryScope": "",
					"timeScopeUnit":   "MONTH",
					"timeScopeValue":  1,
				},
			},
			DeleteWhenDone: true,
		},
		Executor: Executor{
			Enabled: false,
		},
		Smtp: Smtp{
			Enabled: false,
			To:      "example@domain.com",
		},
	}
}

// LoadYamlConfig will unmarshal a YAML configuration file into a ValidationSpec.
func LoadYamlConfig(userConfig []byte) (*ValidationSpec, error) {
	var config ValidationSpec
	if err := yaml.Unmarshal(userConfig, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadJsonConfig will unmarshal a JSON configuration file into a ValidationSpec.
func LoadJsonConfig(userConfig []byte) (*ValidationSpec, error) {
	var config ValidationSpec
	if err := json.Unmarshal(userConfig, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
