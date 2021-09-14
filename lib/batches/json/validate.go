package json

import (
	"encoding/json"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/lib/batches/jsonschema"
)

// UnmarshalValidate validates the JSON input against the provided JSON schema.
// If the validation is successful the validated input is unmarshalled into the
// target.
func UnmarshalValidate(schema string, input []byte, target interface{}) error {
	var errs *multierror.Error
	if err := jsonschema.Validate(schema, input); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := json.Unmarshal(input, target); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}
