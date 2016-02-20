package inject

import (
	"sort"

	"gopkg.in/yaml.v2"
)

// Inject injects a map of parameters into a raw string and returns
// the resulting string.
//
// Parameters are represented in the string using $$ notation, similar
// to how environment variables are defined in Makefiles.
func Inject(raw string, params map[string]string) string {
	if params == nil || len(params) == 0 {
		return raw
	}
	keys := []string{}
	for k := range params {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	injected := raw
	for _, k := range keys {
		v := params[k]

		for _, substitute := range substitutors {
			injected = substitute(injected, k, v)
		}
	}
	return injected
}

// InjectSafe attempts to safely inject parameters without leaking
// parameters in the Build or Compose section of the yaml file.
//
// The intended use case for this function are public pull requests.
// We want to avoid a malicious pull request that allows someone
// to inject and print private variables.
func InjectSafe(raw string, params map[string]string) (string, error) {
	if params == nil || len(params) == 0 {
		return raw, nil
	}
	before, err := parse(raw)
	if err != nil {
		return raw, err
	}
	after, err := parse(Inject(raw, params))
	if err != nil {
		return raw, err
	}
	after.Build = before.Build
	result, err := yaml.Marshal(after)
	return string(result), err
}

// parse unmarshals the yaml file into an intermediate representation
// that isolates the build section. This allows us to modify the rest
// of the Yaml file while preserving the build section.
func parse(raw string) (*config, error) {
	conf := &config{}
	err := yaml.Unmarshal([]byte(raw), &conf)
	return conf, err
}

type config struct {
	Build map[string]interface{} `yaml:"build"`
	Vargs map[string]interface{} `yaml:"vargs,inline"`
}
