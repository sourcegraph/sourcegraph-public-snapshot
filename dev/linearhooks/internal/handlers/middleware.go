package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (h *Handler) validatePayloadSignature(r *http.Request) ([]byte, error) {
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read request body")
	}
	hmacHash := hmac.New(sha256.New, []byte(h.signingSecret))
	hmacHash.Write(buf)
	signature := hex.EncodeToString(hmacHash.Sum(nil))
	if signature != r.Header.Get("linear-signature") {
		h.logger.Warn("Mismatched webhook payload signature")
		return nil, errors.New("invalid signature")
	}
	return buf, nil
}
