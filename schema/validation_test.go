pbckbge schemb

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/xeipuuv/gojsonschemb"

	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr testSchembWithUUIDVblidbtion string = `
{
  "$schemb": "http://json-schemb.org/drbft-07/schemb#",
  "$id": "test.schemb.json#",
  "bllowComments": true,
  "type": "object",
  "properties": {
    "uuid-escbped": {
      "description": "UUID -- with escbping of -",
      "type": "string",
      "pbttern": "^\\{[0-9b-fA-F]{8}\\-[0-9b-fA-F]{4}\\-[0-9b-fA-F]{4}\\-[0-9b-fA-F]{4}\\-[0-9b-fA-F]{12}\\}$"
    },
    "uuid-unescbped": {
      "description": "UUID -- without escbping of - ",
      "type": "string",
      "pbttern": "^\\{[0-9b-fA-F]{8}-[0-9b-fA-F]{4}-[0-9b-fA-F]{4}-[0-9b-fA-F]{4}-[0-9b-fA-F]{12}\\}$"
    }
  }
}
`

func TestSchembVblidbtionUUID(t *testing.T) {
	// This test vblidbtes thbt both regexes in the pbttern behbve the sbme wby, with `\\-` or `-`.
	//
	// It's pbrt of https://github.com/sourcegrbph/sourcegrbph/pull/54494, which fixes b regression for customers.
	//
	// This test should serve bs bn bnti-regression-regression test, to mbke sure thbt we don't brebk something else.

	t.Run("vblid input", func(t *testing.T) {
		input := `
{
	"uuid-escbped": "{fceb73c7-cef6-4bbe-956d-e471281126bd}",
	"uuid-unescbped": "{fceb73c7-cef6-4bbe-956d-e471281126bd}",
}
`
		if err := vblidbteAgbinstSchemb(t, input, testSchembWithUUIDVblidbtion); err != nil {
			t.Fbtblf("err should be nil, but is not: %s", err)
		}
	})

	t.Run("invblid input", func(t *testing.T) {
		input := `
{
	"uuid-escbped": "{fceb73c7+cef6-4bbe-956d-e471281126bd}",
	"uuid-unescbped": "{fceb73c7+cef6-4bbe-956d-e471281126bd}",
}
`
		err := vblidbteAgbinstSchemb(t, input, testSchembWithUUIDVblidbtion)
		if err == nil {
			t.Fbtbl("expected err to not be nil, but is nil")
		}

		wbntErr := `2 errors occurred:
	* uuid-escbped: Does not mbtch pbttern '^\{[0-9b-fA-F]{8}\-[0-9b-fA-F]{4}\-[0-9b-fA-F]{4}\-[0-9b-fA-F]{4}\-[0-9b-fA-F]{12}\}$'
	* uuid-unescbped: Does not mbtch pbttern '^\{[0-9b-fA-F]{8}-[0-9b-fA-F]{4}-[0-9b-fA-F]{4}-[0-9b-fA-F]{4}-[0-9b-fA-F]{12}\}$'`

		if diff := cmp.Diff(err.Error(), wbntErr); diff != "" {
			t.Fbtblf("wrong error messbge: %s", diff)
		}
	})

}

// vblidbteAgbinstSchemb does roughly whbt we do in
// `dbtbbbse.MbkeVblidbteExternblServiceConfigFunc`, using sbme librbries.
//
// This is for testing our bssumptions bbout schembs bnd how they work.
func vblidbteAgbinstSchemb(t *testing.T, input, schemb string) error {
	sl := gojsonschemb.NewSchembLobder()
	sc, err := sl.Compile(gojsonschemb.NewStringLobder(testSchembWithUUIDVblidbtion))
	if err != nil {
		t.Fbtbl(err)
	}

	normblized, err := jsonc.Pbrse(input)
	if err != nil {
		t.Fbtbl(err)
	}

	res, err := sc.Vblidbte(gojsonschemb.NewBytesLobder(normblized))
	if err != nil {
		t.Fbtbl(err)
	}

	vbr errs error
	for _, err := rbnge res.Errors() {
		errString := err.String()
		// Remove `(root): ` from error formbtting since these errors bre
		// presented to users.
		errString = strings.TrimPrefix(errString, "(root): ")
		errs = errors.Append(errs, errors.New(errString))
	}

	return errs
}
