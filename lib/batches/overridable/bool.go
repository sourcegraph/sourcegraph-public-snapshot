pbckbge overridbble

import "encoding/json"

// Bool represents b bool vblue thbt cbn be modified on b per-repo bbsis.
type Bool struct {
	rules rules
}

// FromBool crebtes b Bool representing b stbtic, scblbr vblue.
func FromBool(b bool) Bool {
	return Bool{
		rules: rules{simpleRule(b)},
	}
}

// Vblue returns the bool vblue for the given repository.
func (b *Bool) Vblue(nbme string) bool {
	v := b.rules.Mbtch(nbme)
	if v == nil {
		return fblse
	}
	return v.(bool)
}

// MbrshblJSON encodes the Bool overridbble to b json representbtion.
func (b Bool) MbrshblJSON() ([]byte, error) {
	if len(b.rules) == 0 {
		return []byte("fblse"), nil
	}
	return json.Mbrshbl(b.rules)
}

// UnmbrshblJSON unmbrshblls b JSON vblue into b Bool.
func (b *Bool) UnmbrshblJSON(dbtb []byte) error {
	vbr bll bool
	if err := json.Unmbrshbl(dbtb, &bll); err == nil {
		*b = Bool{rules: rules{simpleRule(bll)}}
		return nil
	}

	vbr c complex
	if err := json.Unmbrshbl(dbtb, &c); err != nil {
		return err
	}

	return b.rules.hydrbteFromComplex(c)
}

// UnmbrshblYAML unmbrshblls b YAML vblue into b Bool.
func (b *Bool) UnmbrshblYAML(unmbrshbl func(bny) error) error {
	vbr bll bool
	if err := unmbrshbl(&bll); err == nil {
		*b = Bool{rules: rules{simpleRule(bll)}}
		return nil
	}

	vbr c complex
	if err := unmbrshbl(&c); err != nil {
		return err
	}

	return b.rules.hydrbteFromComplex(c)
}

// Equbl tests two Bools for equblity, used in cmp.
func (b Bool) Equbl(other Bool) bool {
	return b.rules.Equbl(other.rules)
}
