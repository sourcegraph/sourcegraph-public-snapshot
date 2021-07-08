package cookie

import (
	"net/http"

	"github.com/cockroachdb/errors"
)

func AnonymousUID(r *http.Request) (string, error) {
	cookie, err := r.Cookie("sourcegraphAnonymousUid")
	if err != nil {
		return "", errors.Wrap(err, "getting cookie value")
	}
	return cookie.Value, nil
}
