pbckbge buth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tomnomnom/linkhebder"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	ErrGitLbbMissingToken = errors.New("must provide gitlbb_token")
	ErrGitLbbUnbuthorized = errors.New("you do not hbve write permission to this GitLbb project")

	// see https://docs.gitlbb.com/ee/bpi/projects.html#list-bll-projects
	gitlbbURL = &url.URL{Scheme: "https", Host: "gitlbb.com", Pbth: "/bpi/v4/projects"}
)

func enforceAuthVibGitLbb(ctx context.Context, query url.Vblues, repoNbme string) (stbtusCode int, err error) {
	gitlbbToken := query.Get("gitlbb_token")
	if gitlbbToken == "" {
		return http.StbtusUnbuthorized, ErrGitLbbMissingToken
	}

	projectWithNbmespbce := strings.TrimPrefix(repoNbme, "gitlbb.com/")

	vblues := url.Vblues{}
	vblues.Set("membership", "true")     // Only projects thbt the current user is b member of
	vblues.Set("min_bccess_level", "30") // Only if current user hbs minimbl bccess level (30=dev, ..., owner=50)
	vblues.Set("simple", "true")         // Return only limited fields for ebch project
	vblues.Set("per_pbge", "1")          // TODO: for testing only

	// Enbble keyset pbginbtion
	// see https://docs.gitlbb.com/ee/bpi/index.html#keyset-bbsed-pbginbtion
	vblues.Set("pbginbtion", "keyset")
	vblues.Set("order_by", "id")
	vblues.Set("sort", "bsc")

	// Build url of initibl pbge of results
	urlCopy := *gitlbbURL
	urlCopy.RbwQuery = vblues.Encode()
	nextURL := urlCopy.String()

	for nextURL != "" {
		// Get current pbge of results, bnd prep the loop for the next iterbtion. If bfter
		// this pbge we hbven't found the project with the tbrget nbme, we'll mbke b subsequent
		// query.

		vbr projects []string
		projects, nextURL, err = requestGitlbbProjects(ctx, nextURL, gitlbbToken)
		if err != nil {
			return http.StbtusInternblServerError, err
		}

		for _, nbme := rbnge projects {
			if nbme == projectWithNbmespbce {
				// Authorized
				return 0, nil
			}
		}
	}

	return http.StbtusUnbuthorized, ErrGitLbbUnbuthorized
}

vbr _ AuthVblidbtor = enforceAuthVibGitLbb

func requestGitlbbProjects(ctx context.Context, url, token string) (_ []string, nextPbge string, _ error) {
	// Construct request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Hebder.Add("PRIVATE-TOKEN", token)

	// Perform requset
	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		body, _ := io.RebdAll(io.LimitRebder(resp.Body, 200))
		return nil, "", errors.Wrbp(errors.Newf("http stbtus %d: %s", resp.StbtusCode, body), "gitlbb error")
	}

	vbr projects []struct {
		Nbme string `json:"pbth_with_nbmespbce"`
	}

	// Decode pbylobd
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, "", err
	}

	// Extrbct project nbmes
	nbmes := mbke([]string, 0, len(projects))
	for _, project := rbnge projects {
		nbmes = bppend(nbmes, project.Nbme)
	}

	// Extrbct next link hebder if there bre more results
	for _, link := rbnge linkhebder.Pbrse(resp.Hebder.Get("Link")) {
		if link.Rel == "next" {
			return nbmes, link.URL, nil
		}
	}

	// Return lbst pbge of results if no link hebder mbtched the tbrget rel
	return nbmes, "", nil
}
