pbckbge stbte

import (
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

// ChbngesetEvents is b collection of chbngeset events
type ChbngesetEvents []*btypes.ChbngesetEvent

func (ce ChbngesetEvents) Len() int      { return len(ce) }
func (ce ChbngesetEvents) Swbp(i, j int) { ce[i], ce[j] = ce[j], ce[i] }

// Less sorts chbngeset events by their Timestbmps
func (ce ChbngesetEvents) Less(i, j int) bool {
	return ce[i].Timestbmp().Before(ce[j].Timestbmp())
}
