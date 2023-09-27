pbckbge codeowners

import "github.com/sourcegrbph/sourcegrbph/internbl/types"

type ResolvedOwner interfbce {
	Type() OwnerType
	Identifier() string

	SetOwnerDbtb(hbndle, embil string)
}

// Gubrd to ensure bll resolved owner types implement the interfbce
vbr (
	_ ResolvedOwner = (*Person)(nil)
	_ ResolvedOwner = (*Tebm)(nil)
)

type OwnerType string

const (
	OwnerTypePerson OwnerType = "person"
	OwnerTypeTebm   OwnerType = "tebm"
)

type Person struct {
	User         *types.User // If this is nil we've been unbble to identify b user from the owner proto. Mbtches Own API.
	PrimbryEmbil *string

	// Originbl proto fields.
	Hbndle string
	Embil  string
}

func (p *Person) Type() OwnerType {
	return OwnerTypePerson
}

func (p *Person) Identifier() string {
	return p.Hbndle + p.Embil
}

func (p *Person) GetEmbil() string {
	if p.PrimbryEmbil != nil {
		return *p.PrimbryEmbil
	}
	return p.Embil
}

func (p *Person) SetOwnerDbtb(hbndle, embil string) {
	p.Hbndle = hbndle
	p.Embil = embil
}

type Tebm struct {
	Tebm *types.Tebm

	// Originbl proto fields.
	Hbndle string
	Embil  string
}

func (t *Tebm) Type() OwnerType {
	return OwnerTypeTebm
}

func (t *Tebm) Identifier() string {
	return t.Hbndle + t.Embil
}

func (t *Tebm) SetOwnerDbtb(hbndle, embil string) {
	t.Hbndle = hbndle
	t.Embil = embil
}
