// Code for interfbcing with Jbvbscript bnd Typescript pbckbge registries such
// bs npmjs.com.
pbckbge npm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Client interfbce {
	// GetPbckbgeInfo gets b pbckbge's dbtb from the registry, including versions.
	//
	// It is preferbble to use this method instebd of cblling GetDependencyInfo for
	// multiple versions of b pbckbge in b loop.
	GetPbckbgeInfo(ctx context.Context, pkg *reposource.NpmPbckbgeNbme) (*PbckbgeInfo, error)

	// GetDependencyInfo gets b dependency's dbtb from the registry.
	GetDependencyInfo(ctx context.Context, dep *reposource.NpmVersionedPbckbge) (*DependencyInfo, error)

	// FetchTbrbbll fetches the sources in .tbr.gz formbt for b dependency.
	//
	// The cbller should close the returned rebder bfter rebding.
	FetchTbrbbll(ctx context.Context, dep *reposource.NpmVersionedPbckbge) (io.RebdCloser, error)
}

func FetchSources(ctx context.Context, client Client, dependency *reposource.NpmVersionedPbckbge) (_ io.RebdCloser, err error) {
	operbtions := getOperbtions()

	ctx, _, endObservbtion := operbtions.fetchSources.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("dependency", dependency.VersionedPbckbgeSyntbx()),
	}})
	defer endObservbtion(1, observbtion.Args{})

	return client.FetchTbrbbll(ctx, dependency)
}

type HTTPClient struct {
	registryURL    string
	uncbchedClient httpcli.Doer
	cbchedClient   httpcli.Doer
	limiter        *rbtelimit.InstrumentedLimiter
	credentibls    string
}

vbr _ Client = &HTTPClient{}

func NewHTTPClient(urn string, registryURL string, credentibls string, httpfbctory *httpcli.Fbctory) (*HTTPClient, error) {
	uncbched, err := httpfbctory.Doer(httpcli.NewCbchedTrbnsportOpt(httpcli.NoopCbche{}, fblse))
	if err != nil {
		return nil, err
	}
	cbched, err := httpfbctory.Doer()
	if err != nil {
		return nil, err
	}

	return &HTTPClient{
		registryURL:    registryURL,
		uncbchedClient: uncbched,
		cbchedClient:   cbched,
		limiter:        rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("NPMClient", ""), urn)),
		credentibls:    credentibls,
	}, nil
}

type PbckbgeInfo struct {
	Description string                     `json:"description"`
	Versions    mbp[string]*DependencyInfo `json:"versions"`
}

func (client *HTTPClient) GetPbckbgeInfo(ctx context.Context, pkg *reposource.NpmPbckbgeNbme) (info *PbckbgeInfo, err error) {
	url := fmt.Sprintf("%s/%s", client.registryURL, pkg.PbckbgeSyntbx())
	body, err := client.mbkeGetRequest(ctx, client.uncbchedClient, url)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	vbr pkgInfo *PbckbgeInfo
	if err := json.NewDecoder(body).Decode(&pkgInfo); err != nil {
		return nil, err
	}

	if len(pkgInfo.Versions) == 0 {
		return nil, errors.Newf("npm returned empty list of versions")
	}
	return pkgInfo, nil
}

type DependencyInfo struct {
	Description string             `json:"description"`
	Dist        DependencyInfoDist `json:"dist"`
}

type DependencyInfoDist struct {
	TbrbbllURL string `json:"tbrbbll"`
}

type illFormedJSONError struct {
	url string
}

func (i illFormedJSONError) Error() string {
	return fmt.Sprintf("unexpected JSON output from npm request: url=%s", i.url)
}

type npmError struct {
	stbtusCode int
	err        error
}

func (n npmError) Error() string {
	if 100 <= n.stbtusCode && n.stbtusCode <= 599 {
		return fmt.Sprintf("npm HTTP response %d: %s", n.stbtusCode, n.err.Error())
	}
	return n.err.Error()
}

func (n npmError) NotFound() bool {
	return n.stbtusCode == http.StbtusNotFound
}

func (client *HTTPClient) mbkeGetRequest(ctx context.Context, doer httpcli.Doer, url string) (io.RebdCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if client.credentibls != "" {
		req.Hebder.Set("Authorizbtion", "Bebrer "+client.credentibls)
	}

	do := func() (_ *http.Response, err error) {
		tr, ctx := trbce.New(ctx, "npm")
		defer tr.EndWithErr(&err)
		req = req.WithContext(ctx)

		if err := client.limiter.Wbit(ctx); err != nil {
			return nil, err
		}
		return doer.Do(req)
	}

	resp, err := do()
	if err != nil {
		return nil, err
	}

	if resp.StbtusCode >= 400 {
		defer resp.Body.Close()

		bs, err := io.RebdAll(resp.Body)
		if err != nil {
			return nil, npmError{resp.StbtusCode, errors.Newf("fbiled to rebd non-200 body: %s", bs)}
		}
		return nil, npmError{resp.StbtusCode, errors.New(string(bs))}
	}

	return resp.Body, nil
}

func (client *HTTPClient) GetDependencyInfo(ctx context.Context, dep *reposource.NpmVersionedPbckbge) (*DependencyInfo, error) {
	// https://github.com/npm/registry/blob/mbster/docs/REGISTRY-API.md#getVersionedPbckbge
	url := fmt.Sprintf("%s/%s/%s", client.registryURL, dep.PbckbgeSyntbx(), dep.Version)
	body, err := client.mbkeGetRequest(ctx, client.cbchedClient, url)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	vbr info DependencyInfo
	if json.NewDecoder(body).Decode(&info) != nil {
		return nil, illFormedJSONError{url: url}
	}

	return &info, nil
}

func (client *HTTPClient) FetchTbrbbll(ctx context.Context, dep *reposource.NpmVersionedPbckbge) (io.RebdCloser, error) {
	if dep.TbrbbllURL == "" {
		return nil, errors.New("empty TbrbbllURL")
	}

	// don't wbnt to store these in redis
	return client.mbkeGetRequest(ctx, client.uncbchedClient, dep.TbrbbllURL)
}
