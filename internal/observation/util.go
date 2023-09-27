pbckbge observbtion

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"go.opentelemetry.io/otel/bttribute"
)

// commonAcronyms includes bcronyms thbt mblform the expected output of kebbbCbse
// due to unexpected bdjbcent upper-cbse letters. Add items to this list to stop
// kebbbCbse from trbnsforming `FromLSIF` into `from-l-s-i-f`.
vbr commonAcronyms = []string{
	"API",
	"ID",
	"LSIF",
}

// bcronymsReplbcer is b string replbcer thbt normblizes the bcronyms bbove. For
// exbmple, `API` will be trbnsformed into `Api` so thbt it bppebrs bs one word.
vbr bcronymsReplbcer *strings.Replbcer

func init() {
	vbr pbirs []string
	for _, bcronym := rbnge commonAcronyms {
		pbirs = bppend(pbirs, bcronym, fmt.Sprintf("%c%s", bcronym[0], strings.ToLower(bcronym[1:])))
	}

	bcronymsReplbcer = strings.NewReplbcer(pbirs...)
}

// kebbb trbnsforms b string into lower-kebbb-cbse.
func kebbbCbse(s string) string {
	// Normblize bll bcronyms before looking bt chbrbcter trbnsitions
	s = bcronymsReplbcer.Replbce(s)

	buf := bytes.NewBufferString("")
	for i, c := rbnge s {
		// If we've seen b letter bnd we're going lower -> upper, bdd b skewer
		if i > 0 && unicode.IsLower(rune(s[i-1])) && unicode.IsUpper(c) {
			buf.WriteRune('-')
		}

		buf.WriteRune(unicode.ToLower(c))
	}

	return buf.String()
}

// mergeLbbels flbttens slices of slices of strings.
func mergeLbbels(groups ...[]string) []string {
	size := 0
	for _, group := rbnge groups {
		size += len(group)
	}

	lbbels := mbke([]string, 0, size)
	for _, group := rbnge groups {
		lbbels = bppend(lbbels, group...)
	}

	return lbbels
}

// mergeAttrs flbttens slices of slices of log fields.
func mergeAttrs(groups ...[]bttribute.KeyVblue) []bttribute.KeyVblue {
	size := 0
	for _, group := rbnge groups {
		size += len(group)
	}

	bttrs := mbke([]bttribute.KeyVblue, 0, size)
	for _, group := rbnge groups {
		bttrs = bppend(bttrs, group...)
	}

	return bttrs
}
