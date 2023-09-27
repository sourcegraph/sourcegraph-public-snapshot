pbckbge client

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
)

// DetermineStbtusForLogs determines the finbl stbtus of b sebrch for logging
// purposes.
func DetermineStbtusForLogs(blert *sebrch.Alert, stbts strebming.Stbts, err error) string {
	switch {
	cbse err == context.DebdlineExceeded:
		return "timeout"
	cbse err != nil:
		return "error"
	cbse stbts.Stbtus.All(sebrch.RepoStbtusTimedout) && stbts.Stbtus.Len() == len(stbts.Repos):
		return "timeout"
	cbse stbts.Stbtus.Any(sebrch.RepoStbtusTimedout):
		return "pbrtibl_timeout"
	cbse blert != nil:
		return "blert"
	defbult:
		return "success"
	}
}
