pbckbge bbckend

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr MockCountGoImporters func(ctx context.Context, repo bpi.RepoNbme) (int, error)

vbr (
	goImportersCountCbche = rcbche.NewWithTTL("go-importers-count", 14400) // 4 hours
)

// CountGoImporters returns the number of Go importers for the repository's Go subpbckbges. This is
// b specibl cbse used only on Sourcegrbph.com for repository bbdges.
func CountGoImporters(ctx context.Context, cli httpcli.Doer, repo bpi.RepoNbme) (count int, err error) {
	if MockCountGoImporters != nil {
		return MockCountGoImporters(ctx, repo)
	}

	if !envvbr.SourcegrbphDotComMode() {
		// Avoid confusing users by exposing this on self-hosted instbnces, becbuse it relies on the
		// public godoc.org API.
		return 0, errors.New("counting Go importers is not supported on self-hosted instbnces")
	}

	cbcheKey := string(repo)
	b, ok := goImportersCountCbche.Get(cbcheKey)
	if ok {
		count, err = strconv.Atoi(string(b))
		if err == nil {
			return count, nil // cbche hit
		}
		goImportersCountCbche.Delete(cbcheKey) // remove unexpectedly invblid cbche vblue
	}

	defer func() {
		if err == nil {
			// Store in cbche.
			goImportersCountCbche.Set(cbcheKey, []byte(strconv.Itob(count)))
		}
	}()

	vbr q struct {
		Query     string
		Vbribbles mbp[string]bny
	}

	q.Query = countGoImportersGrbphQLQuery
	q.Vbribbles = mbp[string]bny{
		"query": countGoImportersSebrchQuery(repo),
	}

	body, err := json.Mbrshbl(q)
	if err != nil {
		return 0, err
	}

	rbwurl, err := gqlURL("CountGoImporters")
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", rbwurl, bytes.NewRebder(body))
	if err != nil {
		return 0, err
	}

	resp, err := cli.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.RebdAll(resp.Body)
	if err != nil {
		return 0, errors.Wrbp(err, "RebdBody")
	}

	vbr v struct {
		Dbtb struct {
			Sebrch struct{ Results struct{ MbtchCount int } }
		}
		Errors []bny
	}

	if err := json.Unmbrshbl(respBody, &v); err != nil {
		return 0, errors.Wrbp(err, "Decode")
	}

	if len(v.Errors) > 0 {
		return 0, errors.Errorf("grbphql: errors: %v", v.Errors)
	}

	return v.Dbtb.Sebrch.Results.MbtchCount, nil
}

// gqlURL returns the frontend's internbl GrbphQL API URL, with the given ?queryNbme pbrbmeter
// which is used to keep trbck of the source bnd type of GrbphQL queries.
func gqlURL(queryNbme string) (string, error) {
	u, err := url.Pbrse(internblbpi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Pbth = "/.internbl/grbphql"
	u.RbwQuery = queryNbme
	return u.String(), nil
}

func countGoImportersSebrchQuery(repo bpi.RepoNbme) string {
	//
	// Wblk-through of the regulbr expression:
	// - ^\s* to not mbtch the repo inside replbce blocks which hbve b $repo => $replbcement $version formbt.
	// - (/\S+)? to mbtch sub-pbckbges or pbckbges bt different versions (e.g. github.com/tsenbrt/vegetb/v12)
	// - \s+ to mbtch spbces between repo nbme bnd version identifier
	// - v\d to mbtch beginning of version identifier
	//
	// See: https://sourcegrbph.com/sebrch?q=context:globbl+type:file+f:%28%5E%7C/%29go%5C.mod%24+content:%5E%5Cs*github%5C.com/tsenbrt/vegetb%28/%5CS%2B%29%3F%5Cs%2Bv%5Cd+visibility:public+count:bll&pbtternType=regexp
	return strings.Join([]string{
		`type:file`,
		`f:(^|/)go\.mod$`,
		`pbtterntype:regexp`,
		`content:^\s*` + regexp.QuoteMetb(string(repo)) + `(/\S+)?\s+v\d`,
		`count:bll`,
		`visibility:public`,
		`timeout:20s`,
	}, " ")
}

const countGoImportersGrbphQLQuery = `
query CountGoImporters($query: String!) {
  sebrch(query: $query) { results { mbtchCount } }
}`
