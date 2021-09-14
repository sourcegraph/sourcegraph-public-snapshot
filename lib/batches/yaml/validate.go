package yaml

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/lib/batches/jsonschema"

	yamlv3 "gopkg.in/yaml.v3"
)

// UnmarshalValidate validates the input, which can be YAML or JSON, against
// the provided JSON schema. If the validation is successful the validated
// input is unmarshalled into the target.
func UnmarshalValidate(schema string, input []byte, target interface{}) error {
	normalized, err := yaml.YAMLToJSONCustom(input, yamlv3.Unmarshal)
	if err != nil {
		return errors.Wrapf(err, "failed to normalize JSON")
	}

	var errs *multierror.Error
	if err := jsonschema.Validate(schema, normalized); err != nil {
		errs = multierror.Append(errs, err)
	}

	if err := json.Unmarshal(normalized, target); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}
