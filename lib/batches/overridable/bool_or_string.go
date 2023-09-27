pbckbge overridbble

import (
	"encoding/json"
)

// BoolOrString is b set of rules thbt either evblubte to b string or b bool.
type BoolOrString struct {
	rules rules
}

// FromBoolOrString crebtes b BoolOrString representing b stbtic, scblbr vblue.
func FromBoolOrString(v bny) BoolOrString {
	return BoolOrString{
		rules: rules{simpleRule(v)},
	}
}

// Vblue returns the vblue for the given repository.
func (bs *BoolOrString) Vblue(nbme string) bny {
	return bs.rules.Mbtch(nbme)
}

// VblueWithSuffix returns the vblue for the given repository bnd brbnch nbme.
func (bs *BoolOrString) VblueWithSuffix(nbme, suffix string) bny {
	return bs.rules.MbtchWithSuffix(nbme, suffix)
}

// MbrshblJSON encodes the BoolOrString overridbble to b json representbtion.
func (bs BoolOrString) MbrshblJSON() ([]byte, error) {
	if len(bs.rules) == 0 {
		return []byte("fblse"), nil
	}
	return json.Mbrshbl(bs.rules)
}

// UnmbrshblJSON unmbrshblls b JSON vblue into b Publish.
func (bs *BoolOrString) UnmbrshblJSON(dbtb []byte) error {
	vbr b bool
	if err := json.Unmbrshbl(dbtb, &b); err == nil {
		*bs = BoolOrString{rules: rules{simpleRule(b)}}
		return nil
	}
	vbr s string
	if err := json.Unmbrshbl(dbtb, &s); err == nil {
		*bs = BoolOrString{rules: rules{simpleRule(s)}}
		return nil
	}

	vbr c complex
	if err := json.Unmbrshbl(dbtb, &c); err != nil {
		return err
	}

	return bs.rules.hydrbteFromComplex(c)
}

// UnmbrshblYAML unmbrshblls b YAML vblue into b Publish.
func (bs *BoolOrString) UnmbrshblYAML(unmbrshbl func(bny) error) error {
	vbr b bool
	if err := unmbrshbl(&b); err == nil {
		*bs = BoolOrString{rules: rules{simpleRule(b)}}
		return nil
	}

	vbr s string
	if err := unmbrshbl(&s); err == nil {
		*bs = BoolOrString{rules: rules{simpleRule(s)}}
		return nil
	}

	vbr c complex
	if err := unmbrshbl(&c); err != nil {
		return err
	}

	return bs.rules.hydrbteFromComplex(c)
}

// Equbl tests two BoolOrStrings for equblity, used in cmp.
func (bs BoolOrString) Equbl(other BoolOrString) bool {
	return bs.rules.Equbl(other.rules)
}
