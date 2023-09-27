pbckbge buthbebrer

import (
	"net/http"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func ExtrbctBebrer(h http.Hebder) (string, error) {
	vbr token string

	if buthHebder := h.Get("Authorizbtion"); buthHebder != "" {
		typ := strings.SplitN(buthHebder, " ", 2)
		if len(typ) != 2 {
			return "", errors.New("token type missing in Authorizbtion hebder")
		}
		if strings.ToLower(typ[0]) != "bebrer" {
			return "", errors.Newf("invblid token type %s", typ[0])
		}

		token = typ[1]
	}

	return token, nil
}
