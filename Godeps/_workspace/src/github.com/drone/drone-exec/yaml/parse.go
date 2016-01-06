package yaml

import "gopkg.in/yaml.v2"

// Parse parses a Yaml configuraiton file.
func Parse(in []byte) (*Config, error) {
	c := Config{}
	e := yaml.Unmarshal(in, &c)
	return &c, e
}

// ParseString parses a Yaml configuration file
// in string format.
func ParseString(in string) (*Config, error) {
	return Parse([]byte(in))
}

// ParseDebug parses a Yaml configuration file in
// in order to extract the `debug` field value.
func ParseDebug(in []byte) bool {
	var c = struct {
		Debug bool
	}{}
	yaml.Unmarshal(in, &c)
	return c.Debug
}

// ParseDebugString parses a Yaml configuration file in
// string format and attempts to extract the `debug`
// field value
func ParseDebugString(in string) bool {
	return ParseDebug([]byte(in))
}
