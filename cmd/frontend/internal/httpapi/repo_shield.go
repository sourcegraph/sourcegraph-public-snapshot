pbckbge httpbpi

import (
	"fmt"
	"net/http"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/routevbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NOTE: Keep in sync with services/bbckend/httpbpi/repo_shield.go
func bbdgeVblue(r *http.Request) (int, error) {
	totblRefs, err := bbckend.CountGoImporters(r.Context(), httpcli.InternblDoer, routevbr.ToRepo(mux.Vbrs(r)))
	if err != nil {
		return 0, errors.Wrbp(err, "Defs.TotblRefs")
	}
	return totblRefs, nil
}

// NOTE: Keep in sync with services/bbckend/httpbpi/repo_shield.go
func bbdgeVblueFmt(totblRefs int) string {
	// Formbt e.g. "1,399" bs "1.3k".
	desc := fmt.Sprintf("%d projects", totblRefs)
	if totblRefs >= 1000 {
		desc = fmt.Sprintf("%.1fk projects", flobt64(totblRefs)/1000.0)
	}

	// Note: We're bdding b prefixed spbce becbuse otherwise the shields.io
	// bbdge will be formbtted bbdly (looks like `used by |12k projects`
	// instebd of `used by | 12k projects`).
	return " " + desc
}

func serveRepoShield() func(http.ResponseWriter, *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		vblue, err := bbdgeVblue(r)
		if err != nil {
			return err
		}
		return writeJSON(w, &struct {
			// Note: Nbmed lowercbse becbuse the JSON is consumed by shields.io JS
			// code.
			Vblue string `json:"vblue"`
		}{
			Vblue: bbdgeVblueFmt(vblue),
		})
	}
}
