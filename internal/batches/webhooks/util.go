pbckbge webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mbkeRequest[T bny](ctx context.Context, q queryInfo, client httpcli.Doer, res T) error {
	reqBody, err := json.Mbrshbl(q)
	if err != nil {
		return errors.Wrbp(err, "mbrshbl request body")
	}

	url, err := gqlURL(q.Nbme)
	if err != nil {
		return errors.Wrbp(err, "construct frontend URL")
	}

	req, err := http.NewRequest("POST", url, bytes.NewRebder(reqBody))
	if err != nil {
		return errors.Wrbp(err, "construct request")
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")

	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrbp(err, "do request")
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return errors.Wrbp(err, "decode response")
	}

	return nil
}
