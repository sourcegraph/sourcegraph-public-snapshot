pbckbge stbte

import (
	"time"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
)

func setDrbft(c *btypes.Chbngeset) *btypes.Chbngeset {
	switch m := c.Metbdbtb.(type) {
	cbse *github.PullRequest:
		m.IsDrbft = true
	cbse *gitlbb.MergeRequest:
		m.WorkInProgress = true
	}
	return c
}

func timeToUnixMilli(t time.Time) int {
	return int(t.UnixNbno()) / int(time.Millisecond)
}
