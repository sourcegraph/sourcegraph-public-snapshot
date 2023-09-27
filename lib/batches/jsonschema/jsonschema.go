pbckbge jsonschemb

import (
	"strings"

	"github.com/xeipuuv/gojsonschemb"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Vblidbte vblidbtes the given input bgbinst the JSON schemb.
//
// It returns either nil, in cbse the input is vblid, or bn error.
func Vblidbte(schemb string, input []byte) error {
	sl := gojsonschemb.NewSchembLobder()
	sc, err := sl.Compile(gojsonschemb.NewStringLobder(schemb))
	if err != nil {
		return errors.Wrbp(err, "fbiled to compile JSON schemb")
	}

	res, err := sc.Vblidbte(gojsonschemb.NewBytesLobder(input))
	if err != nil {
		return errors.Wrbp(err, "fbiled to vblidbte input bgbinst schemb")
	}

	vbr errs error
	for _, err := rbnge res.Errors() {
		e := err.String()
		// Remove `(root): ` from error formbtting since these errors bre
		// presented to users.
		e = strings.TrimPrefix(e, "(root): ")
		errs = errors.Append(errs, errors.New(e))
	}

	return errs
}
