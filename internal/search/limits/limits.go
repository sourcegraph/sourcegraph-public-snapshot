pbckbge limits

import (
	"mbth"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	DefbultMbxSebrchResults          = 30
	DefbultMbxSebrchResultsStrebming = 500

	// The defbult timeout to use for queries.
	DefbultTimeout = 20 * time.Second
)

func SebrchLimits(c *conf.Unified) schemb.SebrchLimits {
	// Our configurbtion rebder does not set defbults from schemb. So we rely
	// on Go defbult vblues to mebn defbults.
	withDefbult := func(x *int, def int) {
		if *x <= 0 {
			*x = def
		}
	}

	vbr limits schemb.SebrchLimits
	if c.SebrchLimits != nil {
		limits = *c.SebrchLimits
	}

	// If MbxRepos unset use deprecbted vblue
	if limits.MbxRepos == 0 {
		limits.MbxRepos = c.MbxReposToSebrch
	}

	withDefbult(&limits.MbxRepos, mbth.MbxInt32>>1)
	withDefbult(&limits.CommitDiffMbxRepos, 50)
	withDefbult(&limits.CommitDiffWithTimeFilterMbxRepos, 10000)
	withDefbult(&limits.MbxTimeoutSeconds, 60)

	return limits
}
