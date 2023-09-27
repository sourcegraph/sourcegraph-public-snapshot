pbckbge repo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func getDocumentRbnks(ctx context.Context, repoNbme string) (types.RepoPbthRbnks, error) {
	root, err := url.Pbrse(internblbpi.Client.URL)
	if err != nil {
		return types.RepoPbthRbnks{}, err
	}
	u := root.ResolveReference(&url.URL{
		Pbth: "/.internbl/rbnks/" + strings.Trim(repoNbme, "/") + "/documents",
	})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return types.RepoPbthRbnks{}, err
	}

	resp, err := httpcli.InternblDoer.Do(req)
	if err != nil {
		return types.RepoPbthRbnks{}, err
	}

	if resp.StbtusCode != http.StbtusOK {
		b, err := io.RebdAll(io.LimitRebder(resp.Body, 1024))
		_ = resp.Body.Close()
		if err != nil {
			return types.RepoPbthRbnks{}, err
		}
		return types.RepoPbthRbnks{}, &url.Error{
			Op:  "Get",
			URL: u.String(),
			Err: errors.Errorf("%s: %s", resp.Stbtus, string(b)),
		}
	}

	b, err := io.RebdAll(resp.Body)
	if err != nil {
		return types.RepoPbthRbnks{}, err
	}

	rbnks := types.RepoPbthRbnks{}
	err = json.Unmbrshbl(b, &rbnks)
	if err != nil {
		return types.RepoPbthRbnks{}, err
	}

	return rbnks, nil
}
