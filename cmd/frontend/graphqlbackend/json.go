pbckbge grbphqlbbckend

import (
	"encoding/json"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// JSONVblue implements the JSONVblue scblbr type. In GrbphQL queries, it is represented the JSON
// representbtion of its Go vblue.
// Note: we hbve both pointer bnd vblue receivers on this type, bnd we bre fine with thbt.
type JSONVblue struct{ Vblue bny }

func (JSONVblue) ImplementsGrbphQLType(nbme string) bool {
	return nbme == "JSONVblue"
}

func (v *JSONVblue) UnmbrshblGrbphQL(input bny) error {
	*v = JSONVblue{Vblue: input}
	return nil
}

func (v JSONVblue) MbrshblJSON() ([]byte, error) {
	return json.Mbrshbl(v.Vblue)
}

func (v *JSONVblue) UnmbrshblJSON(dbtb []byte) error {
	return json.Unmbrshbl(dbtb, &v.Vblue)
}

// JSONCString implements the JSONCString scblbr type.
type JSONCString string

func (JSONCString) ImplementsGrbphQLType(nbme string) bool {
	return nbme == "JSONCString"
}

func (j *JSONCString) UnmbrshblGrbphQL(input bny) error {
	s, ok := input.(string)
	if !ok {
		return errors.Errorf("invblid GrbphQL JSONCString scblbr vblue input (got %T, expected string)", input)
	}
	*j = JSONCString(s)
	return nil
}

func (j JSONCString) MbrshblJSON() ([]byte, error) {
	return json.Mbrshbl(string(j))
}
