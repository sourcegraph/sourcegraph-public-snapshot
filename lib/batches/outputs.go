package batches

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	yamlv3 "gopkg.in/yaml.v3"
)

// SetOutputs renders the outputs of the current step into the global outputs
// map using templating.
func SetOutputs(stepOutputs Outputs, global map[string]interface{}, stepCtx *template.StepContext) error {
	for name, output := range stepOutputs {
		var value bytes.Buffer

		if err := template.RenderStepTemplate("outputs-"+name, output.Value, &value, stepCtx); err != nil {
			return errors.Wrap(err, "parsing step run")
		}
		fmt.Printf("Rendering step output %s %s: %q (stdout is %q)\n", name, output.Value, value.String(), stepCtx.Step.Stdout)

		switch output.Format {
		case "yaml":
			var out interface{}
			// We use yamlv3 here, because it unmarshals YAML into
			// map[string]interface{} which we need to serialize it back to
			// JSON when we cache the results.
			// See https://github.com/go-yaml/yaml/issues/139 for context
			if err := yamlv3.NewDecoder(&value).Decode(&out); err != nil {
				return err
			}
			global[name] = out
		case "json":
			var out interface{}
			if err := json.NewDecoder(&value).Decode(&out); err != nil {
				return err
			}
			global[name] = out
		default:
			global[name] = value.String()
		}
	}

	return nil
}
