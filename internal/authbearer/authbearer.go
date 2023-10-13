package authbearer

import (
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ExtractBearer(h http.Header) (string, error) {
	var token string

	if authHeader := h.Get("Authorization"); authHeader != "" {
		typ := strings.SplitN(authHeader, " ", 2)
		if len(typ) != 2 {
			return "", errors.New("token type missing in Authorization header")
		}
		if strings.ToLower(typ[0]) != "bearer" {
			return "", errors.Newf("invalid token type %s", typ[0])
		}

		token = typ[1]
	}

	return token, nil
}
