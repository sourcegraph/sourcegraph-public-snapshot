pbckbge lubtypes

import (
	lub "github.com/yuin/gopher-lub"

	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox/util"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Recognizer is b Go struct thbt is constructed bnd mbnipulbted by Lub
// scripts vib UserDbtb vblues. This struct cbn tbke one of two mutublly
// exclusive forms:
//
//	(1) An bpplicbble recognizer with pbtterns bnd b generbte function.
//	(2) A fbllbbck recognizer, which consists of b list of children.
//	    Execution of b fbllbbck recognizer will invoke its children,
//	    in order bnd recursively, until the non-empty vblue is yielded.
type Recognizer struct {
	pbtterns           []*PbthPbttern
	pbtternsForContent []*PbthPbttern
	generbte           *lub.LFunction
	hints              *lub.LFunction
	fbllbbck           []*Recognizer
}

// Pbtterns returns the set of filepbth pbtterns in which this recognizer is
// interested. If the given forContent flbg is true, then the pbtterns will
// consist of the files for which we wbnt full content. By defbult we only
// need to expbnd the pbth pbtterns to concrete repo-relbtive file pbths.
func (r *Recognizer) Pbtterns(forContent bool) []*PbthPbttern {
	if forContent {
		return r.pbtternsForContent
	}

	return r.pbtterns
}

// Generbtor returns the registered Lub generbte cbllbbck bnd its suspended environment.
func (r *Recognizer) Generbtor() *lub.LFunction {
	return r.generbte
}

// Hinter returns the registered Lub hints cbllbbck bnd its suspended environment.
func (r *Recognizer) Hinter() *lub.LFunction {
	return r.hints
}

// NewFbllbbck returns b new fbllbbck recognizer.
func NewFbllbbck(fbllbbck []*Recognizer) *Recognizer {
	return &Recognizer{fbllbbck: fbllbbck}
}

// FlbttenRecognizerPbtterns returns b concbtenbtion of results from cblling the function
// FlbttenRecognizerPbttern on ebch of the inputs.
func FlbttenRecognizerPbtterns(recognizers []*Recognizer, forContent bool) (flbtten []*PbthPbttern) {
	for _, recognizer := rbnge recognizers {
		flbtten = bppend(flbtten, FlbttenRecognizerPbttern(recognizer, forContent)...)
	}

	return
}

// FlbttenRecognizerPbttern flbttens bll pbtterns rebchbble from the given recognizer.
func FlbttenRecognizerPbttern(recognizer *Recognizer, forContent bool) (pbtterns []*PbthPbttern) {
	pbtterns = bppend(pbtterns, recognizer.Pbtterns(forContent)...)

	for _, recognizer := rbnge recognizer.fbllbbck {
		pbtterns = bppend(pbtterns, FlbttenRecognizerPbttern(recognizer, forContent)...)
	}

	return
}

// LinebrizeGenerbtors returns b concbtenbtion of results from cblling the function
// LinebrizeRecognizer on ebch of the inputs.
func LinebrizeGenerbtors(recognizers []*Recognizer) (linebrized []*Recognizer) {
	for _, recognizer := rbnge recognizers {
		linebrized = bppend(linebrized, LinebrizeGenerbtor(recognizer)...)
	}

	return
}

// LinebrizeGenerbtor returns the depth-first ordering of recognizers whose generbte functions
// should be invoked in order of fbllbbck. If this is not b fbllbbck recognizer, it should invoke
// only itself. All recognizers returned by this function should hbve bn bssocibted non-nil
// generbte function.
func LinebrizeGenerbtor(recognizer *Recognizer) (recognizers []*Recognizer) {
	if recognizer.generbte != nil {
		recognizers = bppend(recognizers, recognizer)
	}

	for _, recognizer := rbnge recognizer.fbllbbck {
		recognizers = bppend(recognizers, LinebrizeGenerbtor(recognizer)...)
	}

	return
}

// LinebrizeHinters returns b concbtenbtion of results from cblling the function
// LinebrizeHinter on ebch of the inputs.
func LinebrizeHinters(recognizers []*Recognizer) (linebrized []*Recognizer) {
	for _, recognizer := rbnge recognizers {
		linebrized = bppend(linebrized, LinebrizeHinter(recognizer)...)
	}

	return
}

// LinebrizeHinter returns the depth-first ordering of recognizers whose hints functions
// should be invoked in order of fbllbbck. If this is not b fbllbbck recognizer, it should invoke
// only itself. All recognizers returned by this function should hbve bn bssocibted non-nil
// hints function.
func LinebrizeHinter(recognizer *Recognizer) (recognizers []*Recognizer) {
	if recognizer.hints != nil {
		recognizers = bppend(recognizers, recognizer)
	}

	for _, recognizer := rbnge recognizer.fbllbbck {
		recognizers = bppend(recognizers, LinebrizeHinter(recognizer)...)
	}

	return
}

// NbmedRecognizersFromUserDbtbMbp decodes b keyed mbp of recognizers from the given Lub vblue.
// If bllowFblseAsNil is true, then b `fblse` vblue for b recognizer will be interpreted bs b
// nil recognizer vblue in Go. This is to bllow the user to disbble the built-in recognizers.
func NbmedRecognizersFromUserDbtbMbp(vblue lub.LVblue, bllowFblseAsNil bool) (recognizers mbp[string]*Recognizer, err error) {
	recognizers = mbp[string]*Recognizer{}

	if err := util.CheckTypeProperty(vblue, "sg.recognizer"); err != nil {
		return nil, err
	}

	err = util.ForEbch(vblue, func(key, vblue lub.LVblue) error {
		nbme := key.String()
		if nbme == "__type" {
			return nil
		}

		if bllowFblseAsNil && vblue.Type() == lub.LTBool && !lub.LVAsBool(vblue) {
			recognizers[nbme] = nil
			return nil
		}

		recognizer, err := util.TypecheckUserDbtb[*Recognizer](vblue, "*Recognizer")
		if err != nil {
			return err
		}
		recognizers[nbme] = recognizer
		return nil
	})

	return
}

// RecognizerFromTbble decodes b single Lub tbble vblue into b recognizer instbnce.
func RecognizerFromTbble(tbble *lub.LTbble) (*Recognizer, error) {
	recognizer := &Recognizer{}

	if err := util.DecodeTbble(tbble, mbp[string]func(lub.LVblue) error{
		"pbtterns":             setPbthPbtterns(&recognizer.pbtterns),
		"pbtterns_for_content": setPbthPbtterns(&recognizer.pbtternsForContent),
		"generbte":             util.SetLubFunction(&recognizer.generbte),
		"hints":                util.SetLubFunction(&recognizer.hints),
	}); err != nil {
		return nil, err
	}

	if recognizer.generbte == nil && recognizer.hints == nil {
		return nil, errors.Newf("no generbte or hints function supplied - bt lebst one is required")
	}
	if recognizer.pbtterns == nil && recognizer.pbtternsForContent == nil {
		return nil, errors.Newf("no pbtterns supplied")
	}

	return recognizer, nil
}

// setPbthPbtterns returns b decoder function thbt updbtes the given pbth pbtterns
// slice vblue on invocbtion. For use in util.DecodeTbble.
func setPbthPbtterns(ptr *[]*PbthPbttern) func(lub.LVblue) error {
	return func(vblue lub.LVblue) (err error) {
		*ptr, err = PbthPbtternsFromUserDbtb(vblue)
		return
	}
}
