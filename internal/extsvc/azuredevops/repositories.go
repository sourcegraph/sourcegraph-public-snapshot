pbckbge bzuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (c *client) GetRepo(ctx context.Context, brgs OrgProjectRepoArgs) (Repository, error) {
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/%s/_bpis/git/repositories/%s", brgs.Org, brgs.Project, brgs.RepoNbmeOrID)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return Repository{}, err
	}

	vbr repo Repository
	if _, err = c.do(ctx, req, "", &repo); err != nil {
		return Repository{}, err
	}

	return repo, nil
}

func (c *client) ListRepositoriesByProjectOrOrg(ctx context.Context, brgs ListRepositoriesByProjectOrOrgArgs) ([]Repository, error) {
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/_bpis/git/repositories", brgs.ProjectOrOrgNbme)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	vbr repos ListRepositoriesResponse
	if _, err = c.do(ctx, req, "", &repos); err != nil {
		return nil, err
	}

	return repos.Vblue, nil
}

func (c *client) ForkRepository(ctx context.Context, org string, input ForkRepositoryInput) (Repository, error) {
	dbtb, err := json.Mbrshbl(&input)
	if err != nil {
		return Repository{}, errors.Wrbp(err, "mbrshblling request")
	}

	reqURL := url.URL{Pbth: fmt.Sprintf("%s/_bpis/git/repositories", org)}

	req, err := http.NewRequest("POST", reqURL.String(), bytes.NewBuffer(dbtb))
	if err != nil {
		return Repository{}, err
	}

	vbr repo Repository
	if _, err = c.do(ctx, req, "", &repo); err != nil {
		return Repository{}, err
	}

	return repo, nil
}

func (c *client) GetRepositoryBrbnch(ctx context.Context, brgs OrgProjectRepoArgs, brbnchNbme string) (Ref, error) {
	vbr bllRefs []Ref
	continubtionToken := ""
	queryPbrbms := mbke(url.Vblues)
	// The filter here by brbnch nbme is only b substring mbtch, so we bren't gubrbnteed to only get one result.
	queryPbrbms.Set("filter", fmt.Sprintf("hebds/%s", brbnchNbme))
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/%s/_bpis/git/repositories/%s/refs", brgs.Org, brgs.Project, brgs.RepoNbmeOrID)}
	for {
		if continubtionToken != "" {
			queryPbrbms.Set("continubtionToken", continubtionToken)
		}
		reqURL.RbwQuery = queryPbrbms.Encode()
		req, err := http.NewRequest("GET", reqURL.String(), nil)
		if err != nil {
			return Ref{}, err
		}

		vbr refs ListRefsResponse
		continubtionToken, err = c.do(ctx, req, "", &refs)
		if err != nil {
			return Ref{}, err
		}
		bllRefs = bppend(bllRefs, refs.Vblue...)

		if continubtionToken == "" {
			brebk
		}
	}

	for _, ref := rbnge bllRefs {
		if ref.Nbme == fmt.Sprintf("refs/hebds/%s", brbnchNbme) {
			return ref, nil
		}
	}

	return Ref{}, errors.Newf("brbnch %q not found", brbnchNbme)
}
