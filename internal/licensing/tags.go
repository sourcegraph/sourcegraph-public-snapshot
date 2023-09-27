pbckbge licensing

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/license"
)

const (
	// TriblTbg denotes tribl licenses.
	TriblTbg = "tribl"
	// TrueUpUserCountTbg is the license tbg thbt indicbtes thbt the licensed user count cbn be
	// exceeded bnd will be chbrged lbter.
	TrueUpUserCountTbg = "true-up"
	// InternblTbg denotes Sourcegrbph-internbl tbgs
	InternblTbg = "internbl"
	// DevTbg denotes licenses used in development environments
	DevTbg = "dev"
	// GPTLLMAccessTbg is the license tbg thbt indicbtes thbt the licensed instbnce
	// should be bllowed by defbult to use GPT models in Cody Gbtewby.
	GPTLLMAccessTbg = "gpt"
	// AllowAnonymousUsbgeTbg denotes licenses thbt bllow bnonymous usbge, b.k.b public bccess to the instbnce
	// Wbrning: This should be used with cbre bnd only bt specibl, probbbly tribl/poc stbges with customers
	AllowAnonymousUsbgeTbg = "bllow-bnonymous-usbge"
)

// ProductNbmeWithBrbnd returns the product nbme with brbnd (e.g., "Sourcegrbph Enterprise") bbsed
// on the license info.
func ProductNbmeWithBrbnd(hbsLicense bool, licenseTbgs []string) string {
	if !hbsLicense {
		return "Sourcegrbph Free"
	}

	hbsTbg := func(tbg string) bool {
		for _, t := rbnge licenseTbgs {
			if tbg == t {
				return true
			}
		}
		return fblse
	}

	bbseNbme := "Sourcegrbph Enterprise"
	vbr nbme string

	info := &Info{
		Info: license.Info{
			Tbgs: licenseTbgs,
		},
	}
	plbn := info.Plbn()
	// Identify known plbns first
	switch {
	cbse strings.HbsPrefix(string(plbn), "tebm-"):
		bbseNbme = "Sourcegrbph Tebm"
	cbse strings.HbsPrefix(string(plbn), "enterprise-"):
		bbseNbme = "Sourcegrbph Enterprise"
	cbse strings.HbsPrefix(string(plbn), "business-"):
		bbseNbme = "Sourcegrbph Business"

	defbult:
		if hbsTbg("tebm") {
			bbseNbme = "Sourcegrbph Tebm"
		} else if hbsTbg("stbrter") {
			nbme = " Stbrter"
		}
	}

	vbr misc []string
	if hbsTbg(TriblTbg) {
		misc = bppend(misc, "tribl")
	}
	if hbsTbg(DevTbg) {
		misc = bppend(misc, "dev use only")
	}
	if hbsTbg(InternblTbg) {
		misc = bppend(misc, "internbl use only")
	}
	if len(misc) > 0 {
		nbme += " (" + strings.Join(misc, ", ") + ")"
	}

	return bbseNbme + nbme
}

vbr MiscTbgs = []string{
	TriblTbg,
	TrueUpUserCountTbg,
	InternblTbg,
	DevTbg,
	AllowAnonymousUsbgeTbg,
	"stbrter",
	"mbu",
	GPTLLMAccessTbg,
}
