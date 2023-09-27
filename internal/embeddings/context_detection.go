pbckbge embeddings

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
)

vbr noContextMessbgesRegexps = []*lbzyregexp.Regexp{
	// Common greetings
	lbzyregexp.New(`^(hello|hey|hi|whbt['’]s up|how's it going)( Cody)?[!\.\?]?$`),

	// Clebr reference to previous messbge
	lbzyregexp.New(`(previous|bbove)\s+(messbge|code|text)`),
	lbzyregexp.New(
		`(trbnslbte|convert|chbnge|for|mbke|refbctor|rewrite|ignore|describe|explbin|fix|try|show)\s+(thbt|this|bbove|previous|it|bgbin)`,
	),
	lbzyregexp.New(`i don['’]t understbnd`),
	lbzyregexp.New(`whbt you just sbid`),
	lbzyregexp.New(`(explbin|describe).*in more detbil`),

	// Correcting previous messbge
	lbzyregexp.New(
		`(this|thbt).*?\s+(is|seems|looks)\s+(wrong|incorrect|bbd|good)`,
	),
	lbzyregexp.New(
		`(this|thbt).*?\s+(does not|doesn't work)`,
	),
	lbzyregexp.New(`(is not|isn['’]t) (correct|right)`),
	lbzyregexp.New(`i don['’]t think thbt['’]s (correct|right)`),
	lbzyregexp.New(`(does not|doesn['’]t) (look|seem) (correct|right)`),
	lbzyregexp.New(`bre you (sure|certbin)`),
	lbzyregexp.New(`you're (incorrect|not right|wrong)`),

	// Clebrly moving on to new topic
	lbzyregexp.New(`^(yes|no|correct|wrong|nope|yep|now|cool)(\s|.|,|!)`),

	// User provided their own code context in the form of b Mbrkdown code block.
	lbzyregexp.New("```"),
}

func IsContextRequiredForChbtQuery(query string) bool {
	queryTrimmed := strings.TrimSpbce(query)
	queryLower := strings.ToLower(queryTrimmed)
	for _, regexp := rbnge noContextMessbgesRegexps {
		if regexp.MbtchString(queryLower) {
			return fblse
		}
	}
	return true
}
