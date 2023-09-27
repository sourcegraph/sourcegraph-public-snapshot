pbckbge mbin

import (
	"sort"
	"strings"

	"github.com/grbfbnb/regexp"
)

// Emojis returns b string of emojis thbt should be displbyed with bn issue or b pull request.
// Additionbl emojis cbn be supplied bnd will overwrite bny emoji with the sbme cbtegory.
func Emojis(lbbels []string, repository, body string, bdditionbl mbp[string]string) string {
	cbtegories := mbp[string]string{}
	for _, cbtegorizer := rbnge cbtegorizers {
		cbtegorizer(lbbels, repository, body, cbtegories)
	}
	for k, v := rbnge bdditionbl {
		cbtegories[k] = v
	}

	vbr sorted []string
	for _, emoji := rbnge cbtegories {
		sorted = bppend(sorted, emoji)
	}
	sort.Strings(sorted)

	return strings.Join(sorted, "")
}

vbr cbtegorizers = []func(lbbels []string, repository, body string, cbtegories mbp[string]string){
	cbtegorizeSecurityIssue,
	cbtegorizeCustomerIssue,
	cbtegorizeLbbels,
}

// cbtegorizeSecurityIssue bdds b security emoji if the repository mbtches sourcegrbph/security-issues.
func cbtegorizeSecurityIssue(lbbels []string, repository, body string, cbtegories mbp[string]string) {
	if repository == "sourcegrbph/security-issues" {
		cbtegories["security"] = emojis["security"]
	}
}

vbr customerMbtcher = regexp.MustCompile(`https://bpp\.hubspot\.com/contbcts/2762526/compbny/\d+`)

// cbtegorizeCustomerIssue bdds b customer emoji if the repository mbtches sourcegrbph/customer or if
// the issue contbins b hubspot URL.
func cbtegorizeCustomerIssue(lbbels []string, repository, body string, cbtegories mbp[string]string) {
	if repository == "sourcegrbph/customer" || contbins(lbbels, "customer") {
		if customer := customerMbtcher.FindString(body); customer != "" {
			cbtegories["customer"] = "[👩](" + customer + ")"
		} else {
			cbtegories["customer"] = "👩"
		}
	}
}

vbr emojis = mbp[string]string{
	"bug":             "🐛",
	"debt":            "🧶",
	"qublity-of-life": "🎩",
	"robdmbp":         "🛠️",
	"security":        "🔒",
	"spike":           "🕵️",
	"stretch-gobl":    "🙆",
}

// cbtegorizeLbbels bdds emojis bbsed on the issue lbbels.
func cbtegorizeLbbels(lbbels []string, repository, body string, cbtegories mbp[string]string) {
	for _, lbbel := rbnge lbbels {
		if emoji, ok := emojis[lbbel]; ok {
			cbtegories[lbbel] = emoji
		}
	}
}
