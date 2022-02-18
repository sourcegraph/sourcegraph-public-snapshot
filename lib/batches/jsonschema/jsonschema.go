package jsonschema

import (
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Validate validates the given input against the JSON schema.
//
// It returns either nil, in case the input is valid, or an error.
func Validate(schema string, input []byte) error {
	sl := gojsonschema.NewSchemaLoader()
	sc, err := sl.Compile(gojsonschema.NewStringLoader(schema))
	if err != nil {
		return errors.Wrap(err, "failed to compile JSON schema")
	}

	res, err := sc.Validate(gojsonschema.NewBytesLoader(input))
	if err != nil {
		return errors.Wrap(err, "failed to validate input against schema")
	}

	var errs error
	for _, err := range res.Errors() {
		e := err.String()
		// Remove `(root): ` from error formatting since these errors are
		// presented to users.
		e = strings.TrimPrefix(e, "(root): ")
		errs = errors.Append(errs, errors.New(e))
	}

	return errs
}
