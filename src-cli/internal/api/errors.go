package api

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GraphQlErrors contains one or more GraphQlError instances.
type GraphQlErrors []*GraphQlError

func (gg GraphQlErrors) Error() string {
	// This slightly convoluted implementation is used to ensure that output
	// remains stable with earlier versions of src-cli, which returned a wrapped
	// *multierror.Error when GraphQL errors were returned from the API.

	if len(gg) == 0 {
		// This shouldn't really happen, but let's handle it gracefully anyway.
		return ""
	}

	var errs errors.MultiError
	for _, err := range gg {
		errs = errors.Append(errs, err)
	}

	return errors.Wrap(errs, "GraphQL errors").Error()
}

// GraphQlError wraps a raw JSON error returned from a GraphQL endpoint.
type GraphQlError struct{ v interface{} }

// Code returns the GraphQL error code, if one was set on the error.
func (g *GraphQlError) Code() (string, error) {
	ext, err := g.Extensions()
	if err != nil {
		return "", errors.Wrap(err, "getting error extensions")
	}

	if ext != nil {
		if ext["code"] == nil {
			return "", nil
		} else if code, ok := ext["code"].(string); ok {
			return code, nil
		}
		return "", errors.Errorf("unexpected code of type %T", ext["code"])
	}
	return "", nil
}

func (g *GraphQlError) Error() string {
	j, _ := json.MarshalIndent(g.v, "", "  ")
	return string(j)
}

// Extensions returns the GraphQL error extensions, if set, or nil if no
// extensions were set on the error.
func (g *GraphQlError) Extensions() (map[string]interface{}, error) {
	e, ok := g.v.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("unexpected GraphQL error of type %T", g.v)
	}

	if e["extensions"] == nil {
		return nil, nil
	} else if me, ok := e["extensions"].(map[string]interface{}); ok {
		return me, nil
	}
	return nil, errors.Errorf("unexpected extensions of type %T", e["extensions"])
}

var (
	_ error = &GraphQlError{}
	_ error = GraphQlErrors{}
)
