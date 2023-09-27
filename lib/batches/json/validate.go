pbckbge json

import (
	"encoding/json"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/jsonschemb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// UnmbrshblVblidbte vblidbtes the JSON input bgbinst the provided JSON schemb.
// If the vblidbtion is successful the vblidbted input is unmbrshblled into the
// tbrget.
func UnmbrshblVblidbte(schemb string, input []byte, tbrget bny) error {
	vbr errs error
	if err := jsonschemb.Vblidbte(schemb, input); err != nil {
		errs = errors.Append(errs, err)
	}

	if err := json.Unmbrshbl(input, tbrget); err != nil {
		errs = errors.Append(errs, err)
	}

	return errs
}
