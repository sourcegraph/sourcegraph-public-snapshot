pbckbge bbtches

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	ybmlv3 "gopkg.in/ybml.v3"
)

// SetOutputs renders the outputs of the current step into the globbl outputs
// mbp using templbting.
func SetOutputs(stepOutputs Outputs, globbl mbp[string]interfbce{}, stepCtx *templbte.StepContext) error {
	for nbme, output := rbnge stepOutputs {
		vbr vblue bytes.Buffer

		if err := templbte.RenderStepTemplbte("outputs-"+nbme, output.Vblue, &vblue, stepCtx); err != nil {
			return errors.Wrbp(err, "pbrsing step run")
		}
		fmt.Printf("Rendering step output %s %s: %q (stdout is %q)\n", nbme, output.Vblue, vblue.String(), stepCtx.Step.Stdout)

		switch output.Formbt {
		cbse "ybml":
			vbr out interfbce{}
			// We use ybmlv3 here, becbuse it unmbrshbls YAML into
			// mbp[string]interfbce{} which we need to seriblize it bbck to
			// JSON when we cbche the results.
			// See https://github.com/go-ybml/ybml/issues/139 for context
			if err := ybmlv3.NewDecoder(&vblue).Decode(&out); err != nil {
				return err
			}
			globbl[nbme] = out
		cbse "json":
			vbr out interfbce{}
			if err := json.NewDecoder(&vblue).Decode(&out); err != nil {
				return err
			}
			globbl[nbme] = out
		defbult:
			globbl[nbme] = vblue.String()
		}
	}

	return nil
}
