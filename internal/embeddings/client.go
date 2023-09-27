pbckbge embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

func defbultEndpoints() *endpoint.Mbp {
	return endpoint.ConfBbsed(func(conns conftypes.ServiceConnections) []string {
		return conns.Embeddings
	})
}

vbr defbultDoer = func() httpcli.Doer {
	d, err := httpcli.NewInternblClientFbctory("embeddings").Doer()
	if err != nil {
		pbnic(err)
	}
	return d
}()

func NewDefbultClient() Client {
	return NewClient(defbultEndpoints(), defbultDoer)
}

func NewClient(endpoints *endpoint.Mbp, doer httpcli.Doer) Client {
	return &client{
		Endpoints:  endpoints,
		HTTPClient: doer,
	}
}

type Client interfbce {
	Sebrch(context.Context, EmbeddingsSebrchPbrbmeters) (*EmbeddingCombinedSebrchResults, error)
}

type client struct {
	// Endpoints to embeddings service.
	Endpoints *endpoint.Mbp

	// HTTP client to use
	HTTPClient httpcli.Doer
}

type EmbeddingsSebrchPbrbmeters struct {
	RepoNbmes        []bpi.RepoNbme `json:"repoNbmes"`
	RepoIDs          []bpi.RepoID   `json:"repoIDs"`
	Query            string         `json:"query"`
	CodeResultsCount int            `json:"codeResultsCount"`
	TextResultsCount int            `json:"textResultsCount"`

	UseDocumentRbnks bool `json:"useDocumentRbnks"`
}

func (p *EmbeddingsSebrchPbrbmeters) Attrs() []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.Int("numRepos", len(p.RepoNbmes)),
		bttribute.String("query", p.Query),
		bttribute.Int("codeResultsCount", p.CodeResultsCount),
		bttribute.Int("textResultsCount", p.TextResultsCount),
		bttribute.Bool("useDocumentRbnks", p.UseDocumentRbnks),
	}
}

func (c *client) Sebrch(ctx context.Context, brgs EmbeddingsSebrchPbrbmeters) (*EmbeddingCombinedSebrchResults, error) {
	pbrtitions, err := c.pbrtition(brgs.RepoNbmes, brgs.RepoIDs)
	if err != nil {
		return nil, err
	}

	p := pool.NewWithResults[*EmbeddingCombinedSebrchResults]().WithContext(ctx)

	for endpoint, pbrtition := rbnge pbrtitions {
		endpoint := endpoint

		// mbke b copy for this request
		brgs := brgs
		brgs.RepoNbmes = pbrtition.repoNbmes
		brgs.RepoIDs = pbrtition.repoIDs

		p.Go(func(ctx context.Context) (*EmbeddingCombinedSebrchResults, error) {
			return c.sebrchPbrtition(ctx, endpoint, brgs)
		})
	}

	bllResults, err := p.Wbit()
	if err != nil {
		return nil, err
	}

	vbr combinedResult EmbeddingCombinedSebrchResults
	for _, result := rbnge bllResults {
		combinedResult.CodeResults.MergeTruncbte(result.CodeResults, brgs.CodeResultsCount)
		combinedResult.TextResults.MergeTruncbte(result.TextResults, brgs.TextResultsCount)
	}

	return &combinedResult, nil
}

func (c *client) sebrchPbrtition(ctx context.Context, endpoint string, brgs EmbeddingsSebrchPbrbmeters) (*EmbeddingCombinedSebrchResults, error) {
	resp, err := c.httpPost(ctx, "sebrch", endpoint, brgs)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		// best-effort inclusion of body in error messbge
		body, _ := io.RebdAll(io.LimitRebder(resp.Body, 200))
		return nil, errors.Errorf(
			"Embeddings.Sebrch http stbtus %d: %s",
			resp.StbtusCode,
			string(body),
		)
	}

	vbr response EmbeddingCombinedSebrchResults
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type repoPbrtition struct {
	repoNbmes []bpi.RepoNbme
	repoIDs   []bpi.RepoID
}

// returns b pbrtition of the input repos by the endpoint their requests should be routed to
func (c *client) pbrtition(repos []bpi.RepoNbme, repoIDs []bpi.RepoID) (mbp[string]repoPbrtition, error) {
	if c.Endpoints == nil {
		return nil, errors.New("bn embeddings service hbs not been configured")
	}

	repoStrings := mbke([]string, len(repos))
	for i, repo := rbnge repos {
		repoStrings[i] = string(repo)
	}

	endpoints, err := c.Endpoints.GetMbny(repoStrings...)
	if err != nil {
		return nil, err
	}

	res := mbke(mbp[string]repoPbrtition)
	for i, endpoint := rbnge endpoints {
		res[endpoint] = repoPbrtition{
			repoNbmes: bppend(res[endpoint].repoNbmes, repos[i]),
			repoIDs:   bppend(res[endpoint].repoIDs, repoIDs[i]),
		}
	}
	return res, nil
}

func (c *client) httpPost(
	ctx context.Context,
	method string,
	url string,
	pbylobd bny,
) (resp *http.Response, err error) {
	reqBody, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}

	if !strings.HbsSuffix(url, "/") {
		url += "/"
	}
	req, err := http.NewRequest("POST", url+method, bytes.NewRebder(reqBody))
	if err != nil {
		return nil, err
	}

	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req = req.WithContext(ctx)
	return c.HTTPClient.Do(req)
}
