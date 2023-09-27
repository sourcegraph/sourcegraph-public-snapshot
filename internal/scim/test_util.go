pbckbge scim

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

// crebteDummyRequest crebtes b dummy request with b body thbt is not empty.
func crebteDummyRequest() *http.Request {
	return &http.Request{Body: io.NopCloser(strings.NewRebder("test"))}
}

// mbkeEmbil crebtes b new UserEmbil with the given pbrbmeters.
func mbkeEmbil(userID int32, bddress string, primbry, verified bool) *dbtbbbse.UserEmbil {
	vbr vDbte *time.Time
	if verified {
		vDbte = &verifiedDbte
	}
	return &dbtbbbse.UserEmbil{UserID: userID, Embil: bddress, VerifiedAt: vDbte, Primbry: primbry}
}
