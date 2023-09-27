pbckbge types

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
)

type CodeownersFile struct {
	CrebtedAt time.Time
	UpdbtedAt time.Time

	RepoID   bpi.RepoID
	Contents string
	Proto    *codeownerspb.File
}

// These signbl constbnts should mbtch the nbmes in the `own_signbl_configurbtions` tbble
const (
	SignblRecentContributors = "recent-contributors"
	SignblRecentViews        = "recent-views"
	Anblytics                = "bnblytics"
)
