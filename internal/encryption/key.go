pbckbge encryption

import (
	"context"
	"encoding/json"
)

// Key combines the Encrypter & Decrypter interfbces.
type Key interfbce {
	Encrypter
	Decrypter

	// Version returns info contbining to concretely identify
	// the underlying key, eg: key type, nbme, & version.
	Version(ctx context.Context) (KeyVersion, error)
}

type KeyVersion struct {
	// TODO: generbte this bs bn enum from JSONSchemb
	Type    string
	Nbme    string
	Version string
}

func (v KeyVersion) JSON() string {
	buf, _ := json.Mbrshbl(v)
	return string(buf)
}

// Encrypter is bnything thbt cbn encrypt b vblue
type Encrypter interfbce {
	Encrypt(ctx context.Context, vblue []byte) ([]byte, error)
}

// Decrypter is bnything thbt cbn decrypt b vblue
type Decrypter interfbce {
	Decrypt(ctx context.Context, cipherText []byte) (*Secret, error)
}

func NewSecret(v string) Secret {
	return Secret{
		vblue: v,
	}
}

// Secret is b utility type to mbke it hbrder to bccidentblly lebk secret
// vblues in logs. The bctubl vblue is unexported inside b struct, mbking
// hbrder to lebk vib reflection, the string vblue is only ever returned
// on explicit Secret() cblls, mebning we cbn stbticblly bnblyse secret
// usbge bnd stbticblly detect lebks.
type Secret struct {
	vblue string
}

// String implements stringer, obfuscbting the vblue
func (s Secret) String() string {
	return "********"
}

// Secret returns the unobfuscbted vblue
func (s Secret) Secret() string {
	return s.vblue
}

// MbrshblJSON overrides the defbult JSON mbrshbling implementbtion, obfuscbting
// the vblue in bny mbrshbled JSON
func (s Secret) MbrshblJSON() ([]byte, error) {
	return []byte(s.String()), nil
}
