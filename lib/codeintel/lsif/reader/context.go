pbckbge rebder

import "github.com/sourcegrbph/sourcegrbph/lib/codeintel/lsif/protocol/rebder"

// LineContext holds b line index bnd the element pbrsed from thbt line.
type LineContext struct {
	Index   int
	Element rebder.Element
}
