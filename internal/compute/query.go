package compute

import (
	"regexp"

	"github.com/sourcegraph/sourcegraph/internal/search/query"

	"github.com/cockroachdb/errors"
)

// RegexpFromQuery extracts a single valid regular expression from a query. If
// it can't, due to presence of, e.g., operators it fails.
func RegexpFromQuery(q string) (*regexp.Regexp, error) {
	plan, err := query.Pipeline(query.Init(q, query.SearchTypeRegex))
	if err != nil {
		return nil, err
	}
	if len(plan) != 1 {
		return nil, errors.New("compute endpoint only supports one search pattern currently ('and' or 'or' operators are not supported yet)")
	}
	switch node := plan[0].Pattern.(type) {
	case query.Operator:
		if len(node.Operands) == 1 {
			if pattern, ok := node.Operands[0].(query.Pattern); ok && !pattern.Negated {
				rp, err := regexp.Compile(pattern.Value)
				if err != nil {
					return nil, errors.Wrap(err, "regular expression is not valid for compute endpoint")
				}
				return rp, nil
			}
		}
		return nil, errors.New("compute endpoint only supports one search pattern currently ('and' or 'or' operators are not supported yet)")
	case query.Pattern:
		if !node.Negated {
			return regexp.Compile(node.Value)
		}
	}
	// unreachable
	return nil, nil
}
