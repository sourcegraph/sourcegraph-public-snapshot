package modelconfig

import (
	"encoding/json"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// deepCopy returns a deep copy of the entire ModelConfiguration data structure.
func deepCopy(source *types.ModelConfiguration) (*types.ModelConfiguration, error) {
	// Rather than manage all the boiler plage by hand, or resorting to reflection
	// we just round-trip the configuration data through JSON.
	//
	// This means that ALL fields in the types package MUST be exported, since
	// unexported fields will silently be dropped by JSON marshalling. But that's
	// not a problem in-practice.
	bytes, err := json.Marshal(source)
	if err != nil {
		return nil, errors.Wrap(err, "marshalling source config")
	}

	var cfgCopy types.ModelConfiguration
	if err = json.Unmarshal(bytes, &cfgCopy); err != nil {
		return nil, errors.Wrap(err, "unmarshalling source config")
	}

	return &cfgCopy, nil
}

// filterListMatches returns whether or not any of the patterns match the supplied
// mref. Assumes the supplied patterns are well-formed. Any asterisks can only be
// in the first or last character of the pattern.
func filterListMatches(mref types.ModelRef, patterns []string) bool {
	s := string(mref)
	for _, pattern := range patterns {
		if pattern == "*" {
			return true
		}

		pLen := len(pattern)
		if pLen < 3 {
			continue // Invalid pattern.
		}
		hasLeadingAsterisk := pattern[0] == '*'
		hasTrailingAsterisk := pattern[pLen-1] == '*'

		// e.g. "*latest"
		if hasLeadingAsterisk && !hasTrailingAsterisk {
			if strings.HasSuffix(s, pattern[1:]) {
				return true
			}
		}
		// e.g. "openai::*"
		if !hasLeadingAsterisk && hasTrailingAsterisk {
			if strings.HasPrefix(s, pattern[:pLen-1]) {
				return true
			}
		}
		// e.g. "anthropic::2023-06-01::claude-3-sonnet"
		if !hasLeadingAsterisk && !hasTrailingAsterisk {
			if s == pattern {
				return true
			}
		}
		// e.g. "*gpt*"
		if hasLeadingAsterisk && hasTrailingAsterisk {
			if strings.Contains(s, pattern[1:pLen-1]) {
				return true
			}
		}
	}

	return false
}
