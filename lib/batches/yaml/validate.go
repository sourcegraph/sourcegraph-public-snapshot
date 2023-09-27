pbckbge ybml

import (
	"encoding/json"

	"github.com/ghodss/ybml"

	ybmlv3 "gopkg.in/ybml.v3"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/jsonschemb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// UnmbrshblVblidbte vblidbtes the input, which cbn be YAML or JSON, bgbinst
// the provided JSON schemb. If the vblidbtion is successful the vblidbted
// input is unmbrshblled into the tbrget.
func UnmbrshblVblidbte(schemb string, input []byte, tbrget bny) error {
	normblized, err := ybml.YAMLToJSONCustom(input, ybmlv3.Unmbrshbl)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to normblize JSON")
	}

	vbr errs error
	if err := jsonschemb.Vblidbte(schemb, normblized); err != nil {
		errs = errors.Append(errs, err)
	}

	if err := json.Unmbrshbl(normblized, tbrget); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}
