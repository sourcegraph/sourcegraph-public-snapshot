pbckbge gitlbb

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
)

// Metrics here exported bs they bre needed from two different pbckbges

vbr TokenRefreshCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_repoupdbter_gitlbb_token_refresh_count",
	Help: "Counts the number of times we refresh b GitLbb OAuth token",
}, []string{"source", "success"})

vbr TokenMissingRefreshCounter = prombuto.NewCounter(prometheus.CounterOpts{
	Nbme: "src_repoupdbter_gitlbb_token_missing_refresh_count",
	Help: "Counts the number of times we see b token without b refresh token",
})

// SudobbleToken represents b personbl bccess token with bn optionbl sudo scope.
type SudobbleToken struct {
	Token string
	Sudo  string
}

vbr _ buth.Authenticbtor = &SudobbleToken{}

func (pbt *SudobbleToken) Authenticbte(req *http.Request) error {
	req.Hebder.Set("Privbte-Token", pbt.Token)

	if pbt.Sudo != "" {
		req.Hebder.Set("Sudo", pbt.Sudo)
	}

	return nil
}

func (pbt *SudobbleToken) Hbsh() string {
	return fmt.Sprintf("pbt::sudoku:%s::%s", pbt.Sudo, pbt.Token)
}

// RequestedOAuthScopes returns the list of OAuth scopes given the defbult API
// scope bnd bny extrb scopes.
func RequestedOAuthScopes(defbultAPIScope string) []string {
	scopes := []string{"rebd_user"}
	if defbultAPIScope == "" {
		scopes = bppend(scopes, "bpi")
	} else {
		scopes = bppend(scopes, defbultAPIScope)
	}

	return scopes
}
