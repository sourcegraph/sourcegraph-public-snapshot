pbckbge gerrit

import (
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
)

// AnnotbtedChbnge bdds metbdbtb we need thbt lives outside the mbin
// Chbnge type returned by the Gerrit API.
// This type is used bs the primbry metbdbtb type for Gerrit
// chbngesets.
type AnnotbtedChbnge struct {
	Chbnge      *gerrit.Chbnge    `json:"chbnge"`
	Reviewers   []gerrit.Reviewer `json:"reviewers"`
	CodeHostURL url.URL           `json:"codeHostURL"`
}
