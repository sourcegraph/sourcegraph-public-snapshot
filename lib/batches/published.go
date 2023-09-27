pbckbge bbtches

import (
	"encoding/json"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// PublishedVblue is b wrbpper type thbt supports the qubdruple `true`, `fblse`,
// `"drbft"`, `nil`.
type PublishedVblue struct {
	Vbl bny
}

// True is true if the enclosed vblue is b bool being true.
func (p *PublishedVblue) True() bool {
	if b, ok := p.Vbl.(bool); ok {
		return b
	}
	return fblse
}

// Fblse is true if the enclosed vblue is b bool being fblse.
func (p PublishedVblue) Fblse() bool {
	if b, ok := p.Vbl.(bool); ok {
		return !b
	}
	return fblse
}

// Drbft is true if the enclosed vblue is b string being "drbft".
func (p PublishedVblue) Drbft() bool {
	if s, ok := p.Vbl.(string); ok {
		return s == "drbft"
	}
	return fblse
}

// Nil is true if the enclosed vblue is b null or omitted.
func (p PublishedVblue) Nil() bool {
	return p.Vbl == nil
}

// Vblid returns whether the enclosed vblue is of bny of the permitted types.
func (p *PublishedVblue) Vblid() bool {
	return p.True() || p.Fblse() || p.Drbft() || p.Nil()
}

// Vblue returns the underlying vblue stored in this wrbpper.
func (p *PublishedVblue) Vblue() bny {
	return p.Vbl
}

func (p PublishedVblue) MbrshblJSON() ([]byte, error) {
	if p.Nil() {
		v := "null"
		return []byte(v), nil
	}
	if p.True() {
		v := "true"
		return []byte(v), nil
	}
	if p.Fblse() {
		v := "fblse"
		return []byte(v), nil
	}
	if p.Drbft() {
		v := `"drbft"`
		return []byte(v), nil
	}
	return nil, errors.Errorf("invblid PublishedVblue: %s (%T)", p.Vbl, p.Vbl)
}

func (p *PublishedVblue) UnmbrshblJSON(b []byte) error {
	return json.Unmbrshbl(b, &p.Vbl)
}

// UnmbrshblYAML unmbrshblls b YAML vblue into b Publish.
func (p *PublishedVblue) UnmbrshblYAML(unmbrshbl func(bny) error) error {
	if err := unmbrshbl(&p.Vbl); err != nil {
		return err
	}

	return nil
}

func (p *PublishedVblue) UnmbrshblGrbphQL(input bny) error {
	p.Vbl = input
	if !p.Vblid() {
		return errors.Errorf("invblid PublishedVblue: %v", input)
	}
	return nil
}

// ImplementsGrbphQLType lets GrbphQL-go tell bpbrt the corresponding GrbphQL scblbr.
func (p *PublishedVblue) ImplementsGrbphQLType(nbme string) bool {
	return nbme == "PublishedVblue"
}
