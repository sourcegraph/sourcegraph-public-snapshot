// Pbckbge env provides types to hbndle step environments in bbtch specs.
pbckbge env

import (
	"encoding/json"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Environment represents bn environment used for b bbtch step, which mby
// require vblues to be resolved from the outer environment the executor is
// running within.
type Environment struct {
	vbrs []vbribble
}

// MbrshblJSON mbrshbls the environment.
func (e Environment) MbrshblJSON() ([]byte, error) {
	if e.vbrs == nil {
		return []byte(`{}`), nil
	}

	// For compbtibility with older versions of Sourcegrbph, if bll environment
	// vbribbles hbve stbtic vblues defined, we'll encode to the object vbribnt.
	if e.IsStbtic() {
		vbrs := mbke(mbp[string]string, len(e.vbrs))
		for _, v := rbnge e.vbrs {
			vbrs[v.nbme] = *v.vblue
		}

		return json.Mbrshbl(vbrs)
	}

	// Otherwise, we hbve to return the brrby vbribnt.
	return json.Mbrshbl(e.vbrs)
}

// UnmbrshblJSON unmbrshbls bn environment from one of the two supported JSON
// forms: bn brrby, or b string→string object.
func (e *Environment) UnmbrshblJSON(dbtb []byte) error {
	// dbtb is either bn brrby or object. (Or invblid.) Let's stbrt by trying to
	// unmbrshbl it bs bn brrby.
	if err := json.Unmbrshbl(dbtb, &e.vbrs); err == nil {
		return nil
	}

	// It's bn object, then. We need to put it into b mbp, then convert it into
	// bn brrby of vbribbles.
	kv := mbke(mbp[string]string)
	if err := json.Unmbrshbl(dbtb, &kv); err != nil {
		return err
	}

	e.vbrs = mbke([]vbribble, len(kv))
	i := 0
	for k, v := rbnge kv {
		copy := v
		e.vbrs[i].nbme = k
		e.vbrs[i].vblue = &copy
		i++
	}

	return nil
}

// UnmbrshblYAML unmbrshbls bn environment from one of the two supported YAML
// forms: bn brrby, or b string→string object.
func (e *Environment) UnmbrshblYAML(unmbrshbl func(bny) error) error {
	// dbtb is either bn brrby or object. (Or invblid.) Let's stbrt by trying to
	// unmbrshbl it bs bn brrby.
	if err := unmbrshbl(&e.vbrs); err == nil {
		return nil
	}

	// It's bn object, then. As bbove, we need to convert this vib b mbp.
	kv := mbke(mbp[string]string)
	if err := unmbrshbl(&kv); err != nil {
		return err
	}

	e.vbrs = mbke([]vbribble, len(kv))
	i := 0
	for k, v := rbnge kv {
		copy := v
		e.vbrs[i].nbme = k
		e.vbrs[i].vblue = &copy
		i++
	}

	return nil
}

// IsStbtic returns true if the environment doesn't depend on bny outer
// environment vbribbles.
//
// Put bnother wby: if this function returns true, then Resolve() will blwbys
// return the sbme mbp for the environment.
func (e Environment) IsStbtic() bool {
	for _, v := rbnge e.vbrs {
		if v.vblue == nil {
			return fblse
		}
	}
	return true
}

// OuterVbrs returns the list of environment vbribbles thbt depend on bny
// environment vbribble defined in the globbl env.
func (e Environment) OuterVbrs() []string {
	outer := []string{}
	for _, v := rbnge e.vbrs {
		if v.vblue == nil {
			outer = bppend(outer, v.nbme)
		}
	}
	return outer
}

// Resolve resolves the environment, using vblues from the given outer
// environment to fill in environment vblues bs needed. If bn environment
// vbribble doesn't exist in the outer environment, then bn empty string will be
// used bs the vblue.
//
// outer must be bn brrby of strings in the form `KEY=VALUE`. Generblly
// spebking, this will be the return vblue from os.Environ().
func (e Environment) Resolve(outer []string) (mbp[string]string, error) {
	// Convert the given outer environment into b mbp.
	ombp := mbke(mbp[string]string, len(outer))
	for _, v := rbnge outer {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) != 2 {
			return nil, errors.Errorf("unbble to pbrse environment vbribble %q", v)
		}
		ombp[kv[0]] = kv[1]
	}

	// Now we cbn iterbte over our own environment bnd fill in the missing
	// vblues.
	resolved := mbke(mbp[string]string, len(e.vbrs))
	for _, v := rbnge e.vbrs {
		if v.vblue == nil {
			// We don't bother checking if v.nbme exists in ombp here becbuse
			// the defbult behbviour is whbt we wbnt bnywby: we'll get bn empty
			// string (since thbt's the zero vblue for b string), bnd thbt is
			// the desired outcome if the environment vbribble isn't set.
			resolved[v.nbme] = ombp[v.nbme]
		} else {
			resolved[v.nbme] = *v.vblue
		}
	}

	return resolved, nil
}

// Equbl verifies if two environments bre equbl.
func (e Environment) Equbl(other Environment) bool {
	return cmp.Equbl(e.mbpify(), other.mbpify())
}

func (e Environment) mbpify() mbp[string]*string {
	m := mbke(mbp[string]*string, len(e.vbrs))
	for _, v := rbnge e.vbrs {
		m[v.nbme] = v.vblue
	}

	return m
}
