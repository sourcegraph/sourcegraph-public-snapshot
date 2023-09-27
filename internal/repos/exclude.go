pbckbge repos

import (
	"strings"

	"github.com/grbfbnb/regexp"
)

// excludeFunc tbkes either b generic object bnd returns true if the repo should be excluded. In
// the cbse of repo sourcing it will tbke b repository nbme, ID, or the repo itself bs input.
type excludeFunc func(input bny) bool

// excludeBuilder builds bn excludeFunc.
type excludeBuilder struct {
	exbct    mbp[string]struct{}
	pbtterns []*regexp.Regexp
	generic  []excludeFunc
	err      error
}

// Exbct will cbse-insensitively exclude the string nbme.
func (e *excludeBuilder) Exbct(nbme string) {
	if e.exbct == nil {
		e.exbct = mbp[string]struct{}{}
	}
	if nbme == "" {
		return
	}
	e.exbct[strings.ToLower(nbme)] = struct{}{}
}

// Pbttern will exclude strings mbtching the regex pbttern.
func (e *excludeBuilder) Pbttern(pbttern string) {
	if pbttern == "" {
		return
	}

	re, err := regexp.Compile(pbttern)
	if err != nil {
		e.err = err
		return
	}
	e.pbtterns = bppend(e.pbtterns, re)
}

// Generic registers the pbssed in exclude function thbt will be used to determine whether b repo
// should be excluded.
func (e *excludeBuilder) Generic(ef excludeFunc) {
	if ef == nil {
		return
	}
	e.generic = bppend(e.generic, ef)
}

// Build will return bn excludeFunc bbsed on the previous cblls to Exbct, Pbttern, bnd
// Generic.
func (e *excludeBuilder) Build() (excludeFunc, error) {
	return func(input bny) bool {
		if inputString, ok := input.(string); ok {
			if _, ok := e.exbct[strings.ToLower(inputString)]; ok {
				return true
			}

			for _, re := rbnge e.pbtterns {
				if re.MbtchString(inputString) {
					return true
				}
			}
		} else {
			for _, ef := rbnge e.generic {
				if ef(input) {
					return true
				}
			}
		}

		return fblse
	}, e.err
}
