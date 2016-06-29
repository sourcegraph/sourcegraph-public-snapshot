package yaml

// Config is a typed representation of the
// Yaml configuration file.
type Config struct {
	Cache Plugin `yaml:",omitempty"`
	Clone Plugin `yaml:",omitempty"`
	Build Builds `yaml:",omitempty"`

	Compose Containers `yaml:",omitempty"`
	Publish Plugins    `yaml:",omitempty"`
	Deploy  Plugins    `yaml:",omitempty"`
	Notify  Plugins    `yaml:",omitempty"`
}

// Container is a typed representation of a
// docker step in the Yaml configuration file.
type Container struct {
	Image       string        `yaml:",omitempty"`
	Pull        bool          `yaml:",omitempty"`
	Privileged  bool          `yaml:",omitempty"`
	Environment MapEqualSlice `yaml:",omitempty"`
	Entrypoint  Command       `yaml:",omitempty"`
	Command     Command       `yaml:",omitempty"`
	ExtraHosts  []string      `yaml:"extra_hosts,omitempty"`
	Volumes     []string      `yaml:",omitempty"`
	Net         string        `yaml:",omitempty"`
	AuthConfig  AuthConfig    `yaml:"auth_config,omitempty"`
	Memory      int64         `yaml:"mem_limit,omitempty"`
	CPUSetCPUs  string        `yaml:"cpuset,omitempty"`
}

// Build is a typed representation of the build
// step in the Yaml configuration file.
type Build struct {
	Container `yaml:",inline"`

	Commands []string `yaml:",omitempty"`
	Filter   Filter   `yaml:"when,omitempty"`

	AllowFailure bool `yaml:"allow_failure,omitempty"`
}

// Auth for Docker Image Registry
type AuthConfig struct {
	Username      string `yaml:"username,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Email         string `yaml:"email,omitempty"`
	RegistryToken string `yaml:"registry_token,omitempty"`
}

// Plugin is a typed representation of a
// docker plugin step in the Yaml configuration
// file.
type Plugin struct {
	Container `yaml:",inline"`

	Vargs  Vargs  `yaml:",inline"`
	Filter Filter `yaml:"when,omitempty"`
}

// Vargs holds unstructured arguments, specific
// to the plugin, that are used at runtime when
// executing the plugin.
type Vargs map[string]interface{}

// Filter is a typed representation of filters
// used at runtime to decide if a particular
// plugin should be executed or skipped.
type Filter struct {
	Repo    string            `yaml:",omitempty"`
	Branch  Stringorslice     `yaml:",omitempty"`
	Event   Stringorslice     `yaml:",omitempty"`
	Success string            `yaml:",omitempty"`
	Failure string            `yaml:",omitempty"`
	Change  string            `yaml:",omitempty"`
	Matrix  map[string]string `yaml:",omitempty"`
}
