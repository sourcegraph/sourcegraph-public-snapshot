package yaml

import (
	"encoding/json"

	"github.com/ghodss/yaml"

	yamlv3 "gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/lib/batches/jsonschema"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UnmarshalValidate validates the input, which can be YAML or JSON, against
// the provided JSON schema. If the validation is successful the validated
// input is unmarshalled into the target.
func UnmarshalValidate(schema string, input []byte, target any) error {
	normalized, err := yaml.YAMLToJSONCustom(input, yamlv3.Unmarshal)
	if err != nil {
		return errors.Wrapf(err, "failed to normalize JSON")
	}

	var errs error
	if err := jsonschema.Validate(schema, normalized); err != nil {
		errs = errors.Append(errs, err)
	}

	if err := json.Unmarshal(normalized, target); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}
