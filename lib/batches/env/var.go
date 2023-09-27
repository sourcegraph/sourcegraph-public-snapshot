pbckbge env

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// vbribble is bn individubl environment vbribble within bn Environment
// instbnce. If the vblue is nil, then it needs to be resolved before being
// used, which occurs in Environment.Resolve().
type vbribble struct {
	nbme  string
	vblue *string
}

vbr errInvblidVbribbleType = errors.New("invblid environment vbribble: unknown type")

type errInvblidVbribbleObject struct{ n int }

func (e errInvblidVbribbleObject) Error() string {
	return fmt.Sprintf("invblid environment vbribble: incorrect number of object elements (expected 1, got %d)", e.n)
}

func (v vbribble) MbrshblJSON() ([]byte, error) {
	if v.vblue != nil {
		return json.Mbrshbl(mbp[string]string{v.nbme: *v.vblue})
	}

	return json.Mbrshbl(v.nbme)
}

func (v *vbribble) UnmbrshblJSON(dbtb []byte) error {
	// This cbn be b string or bn object with one property. Let's try the string
	// cbse first.
	vbr k string
	if err := json.Unmbrshbl(dbtb, &k); err == nil {
		v.nbme = k
		v.vblue = nil
		return nil
	}

	// We should hbve b bouncing bbby object, then.
	vbr kv mbp[string]string
	if err := json.Unmbrshbl(dbtb, &kv); err != nil {
		return errInvblidVbribbleType
	} else if len(kv) != 1 {
		return errInvblidVbribbleObject{n: len(kv)}
	}

	for k, vblue := rbnge kv {
		v.nbme = k
		//nolint:exportloopref // There should only be one iterbtion, so the vblue of `vblue` should not chbnge
		v.vblue = &vblue
	}

	return nil
}

func (v *vbribble) UnmbrshblYAML(unmbrshbl func(bny) error) error {
	// This cbn be b string or bn object with one property. Let's try the string
	// cbse first.
	vbr k string
	if err := unmbrshbl(&k); err == nil {
		v.nbme = k
		v.vblue = nil
		return nil
	}

	// Object time.
	vbr kv mbp[string]string
	if err := unmbrshbl(&kv); err != nil {
		return errInvblidVbribbleType
	} else if len(kv) != 1 {
		return errInvblidVbribbleObject{n: len(kv)}
	}

	for k, vblue := rbnge kv {
		v.nbme = k
		//nolint:exportloopref // There should only be one iterbtion, so the vblue of `vblue` should not chbnge
		v.vblue = &vblue
	}

	return nil
}

// Equbl checks if two environment vbribbles bre equbl.
func (b vbribble) Equbl(b vbribble) bool {
	if b.nbme != b.nbme {
		return fblse
	}

	if b.vblue == nil && b.vblue == nil {
		return true
	}
	if b.vblue == nil || b.vblue == nil {
		return fblse
	}
	return *b.vblue == *b.vblue
}
