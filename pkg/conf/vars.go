package conf

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"

	"github.com/sourcegraph/sourcegraph/pkg/env"
)

// configVarsEnabled (see description below) affects *how* configuration JSON is parsed, so it is
// specified as an env var (not in experimentalFeatures) to avoid a circular dependency ("needing to
// parse the JSON to determine how to parse the JSON").
var configVarsEnabled, _ = strconv.ParseBool(env.Get("SOURCEGRAPH_EXPAND_CONFIG_VARS", "false", "expand ${var} and $var in site configuration JSON based on the environment (env vars)"))

// expandEnv replaces $var or ${var} based on environment variable values. See the docstring of the
// expand func for details.
//
// If SOURCEGRAPH_EXPAND_CONFIG_VARS is disabled, it returns the input unchanged and a nil error.
func expandEnv(data []byte) (interpolatedData []byte, seenVars []string, err error) {
	if !configVarsEnabled {
		return data, nil, nil
	}
	return expand(data, os.Getenv)
}

// expand replaces $var or ${var} in all JSON strings (except JSON object property names)
// based on the mapping function.
//
// This func implements SOURCEGRAPH_EXPAND_CONFIG_VARS. Callers should use expandEnv instead of this
// low-level func.
func expand(data []byte, mapping func(string) string) (interpolatedData []byte, seenVars []string, err error) {
	if len(data) == 0 {
		return nil, nil, nil
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, nil, err
	}

	// Record the vars we've seen, to help with debugging.
	seenVarsMap := map[string]struct{}{}
	expand := func(s string) string {
		return os.Expand(s, func(name string) string {
			if name == "$" {
				return "$"
			}
			seenVarsMap[name] = struct{}{}
			return mapping(name)
		})
	}

	var walk func(vp *interface{})
	walk = func(vp *interface{}) {
		switch v := (*vp).(type) {
		case string:
			*vp = expand(v)
		case map[string]interface{}:
			for name, value := range v {
				walk(&value)
				v[name] = value
			}
		case []interface{}:
			for i, value := range v {
				walk(&value)
				v[i] = value
			}
		}
	}
	walk(&v)

	seenVars = make([]string, 0, len(seenVarsMap))
	for k := range seenVarsMap {
		seenVars = append(seenVars, k)
	}
	sort.Strings(seenVars)

	interpolatedData, err = json.Marshal(v)
	return interpolatedData, seenVars, err
}
