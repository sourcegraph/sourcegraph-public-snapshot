package json

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/lib/batches/jsonschema"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UnmarshalValidate validates the JSON input against the provided JSON schema.
// If the validation is successful the validated input is unmarshalled into the
// target.
func UnmarshalValidate(schema string, input []byte, target any) error {
	var errs error
	if err := jsonschema.Validate(schema, input); err != nil {
		errs = errors.Append(errs, err)
	}

	if err := json.Unmarshal(input, target); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}
