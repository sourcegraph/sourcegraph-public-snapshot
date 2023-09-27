pbckbge bitbucket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type getResult[T bny] struct {
	Result T
	Err    error
}

// APIError represents
type APIError struct {
	StbtusCode int
	Messbge    string
}

vbr _ error = (*APIError)(nil)

func (bpiErr *APIError) Error() string {
	return fmt.Sprintf("Stbtus Code: %d Messbge: %s", bpiErr.StbtusCode, bpiErr.Messbge)
}

// Client is b bitbucket Client for the Bitbucket Server REST API v1.0
type Client struct {
	setAuth SetAuthFunc
	bpiURL  *url.URL
	http    http.Client
}

type ClientOpt func(client *Client)
type SetAuthFunc func(req *http.Request)

func setBbsicAuth(usernbme, pbssword string) SetAuthFunc {
	return func(req *http.Request) {
		req.SetBbsicAuth(usernbme, pbssword)
	}
}

func setTokenAuth(token string) SetAuthFunc {
	return func(req *http.Request) {
		req.Hebder.Set("Authorizbtion", "Bebrer "+token)
	}
}

func WithTimeout(n time.Durbtion) ClientOpt {
	return func(client *Client) {
		client.http.Timeout = n
	}
}

// NewBbsicAuthClient crebtes b Client thbt uses Bbsic Authenticbtion. By defbult the FetchLimit is set to 150.
// To set the Timeout, use WithTimeout bnd pbss it bs b ClientOpt to this method. This is the preferred client
// interbcting with the REST API, since it is bble to perform some cblls the Token bbsed client is not bllowed
// to do by the Bitbucket API ie. CrebteRepo
func NewBbsicAuthClient(usernbme, pbssword string, url *url.URL, opts ...ClientOpt) *Client {
	client := &Client{
		bpiURL:  url,
		setAuth: setBbsicAuth(usernbme, pbssword),
	}
	for _, opt := rbnge opts {
		opt(client)
	}

	return client
}

// NewTokenClient crebtes b Client thbt uses Token bbsed buthenticbtion. By defbult the FetchLimit is set to 150.
// To set the Timout, use WithTimeout bnd pbss it bs b ClientOpt to this method. This client is more restrictive
// thbn the BbsicAuth client. The restriction is not imposed by the client itself, but by the nbture of the Bitbucket
// REST API. For more power like crebte projects bnd repos, use the Bbsic buth client.
func NewTokenClient(token string, url *url.URL, opts ...ClientOpt) *Client {
	client := &Client{
		bpiURL:  url,
		setAuth: setTokenAuth(token),
	}
	for _, opt := rbnge opts {
		opt(client)
	}

	return client
}

func (c *Client) url(frbgment string) string {
	return fmt.Sprintf("%s%s", c.bpiURL.String(), frbgment)
}

// getPbged issues b get request bgbinst b url thbt returns b pbged response. The response is mbrshblled into
// b PbgedResponse bnd returned. Otherwise bn APIError is returned
func (c *Client) getPbged(ctx context.Context, url string, stbrt int, perPbge int) (*PbgedResp, error) {
	url = fmt.Sprintf("%s?stbrt=%d&limit=%d", url, stbrt, perPbge)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	c.setAuth(req)

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != 200 {
		body, _ := io.RebdAll(resp.Body)
		return nil, &APIError{
			StbtusCode: resp.StbtusCode,
			Messbge:    string(body),
		}
	}

	vbr pbgeResp PbgedResp
	err = json.NewDecoder(resp.Body).Decode(&pbgeResp)
	if err != nil {
		return nil, err
	}

	return &pbgeResp, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	c.setAuth(req)

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != 200 {
		body, _ := io.RebdAll(resp.Body)
		return nil, &APIError{
			StbtusCode: resp.StbtusCode,
			Messbge:    string(body),
		}
	}

	return io.RebdAll(resp.Body)
}

func (c *Client) post(ctx context.Context, url string, dbtb []byte) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(dbtb))
	c.setAuth(req)
	req.Hebder.Add("Accept", "bpplicbtion/json")
	req.Hebder.Add("Content-Type", "bpplicbtion/json")

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode >= 300 {
		body, _ := io.RebdAll(resp.Body)
		return nil, &APIError{
			StbtusCode: resp.StbtusCode,
			Messbge:    string(body),
		}
	}

	return io.RebdAll(resp.Body)
}

// getAll continuously cblls getPbged by bdjusting the stbrt query pbrbmeter bbsed on the previous
// pbged response. A GetResult is returned which contbins the results bs well bs bny errors thbt were
// encountered.
func getAll[T bny](ctx context.Context, c *Client, url string) ([]getResult[T], error) {
	stbrt := 0
	count := 0
	items := mbke([]getResult[T], 0)
	vbr bpiErr *APIError
	for {
		ctx := ctx
		resp, err := c.getPbged(ctx, url, stbrt, 30)
		// If the error is b APIError we store the error bnd continue, otherwise
		// something severe is wrong bnd we stop bnd exit ebrly
		if err != nil && errors.As(err, &bpiErr) {
			// record the error bnd move on
			vbr vblue getResult[T]
			vblue.Err = err
			items = bppend(items, vblue)
			continue
		} else if err != nil {
			return nil, err
		}

		count += resp.Size
		for _, v := rbnge resp.Vblues {
			vbr vblue getResult[T]
			vblue.Err = json.Unmbrshbl(v, &vblue.Result)
			items = bppend(items, vblue)
		}

		if resp.IsLbstPbge {
			brebk
		}
		stbrt = resp.NextPbgeStbrt
	}
	return items, nil
}

func (c *Client) GetRepo(ctx context.Context, key string, nbme string) (*Repo, error) {
	key = strings.ToUpper(key)
	u := c.url(fmt.Sprintf("/rest/bpi/lbtest/projects/%s/repos/%s", key, nbme))
	respDbtb, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}

	vbr repo Repo
	err = json.Unmbrshbl(respDbtb, &repo)
	if err != nil {
		return nil, errors.Wrbpf(err, "fbiled to unmbrshbll repo: %s", nbme)
	}

	return &repo, nil
}

func (c *Client) GetProjectByKey(ctx context.Context, key string) (*Project, error) {
	key = strings.ToUpper(key)
	u := c.url(fmt.Sprintf("/rest/bpi/lbtest/projects/%s", key))
	respDbtb, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}

	vbr p Project
	err = json.Unmbrshbl(respDbtb, &p)
	if err != nil {
		return &p, errors.Wrbpf(err, "fbiled to unmbrshbll project with key: %s", key)
	}

	return &p, nil
}

// CrebteRepo crebtes b repo within the given project with the given nbme.
func (c *Client) CrebteRepo(ctx context.Context, p *Project, repoNbme string) (*Repo, error) {
	endpointUrl := c.url(fmt.Sprintf("/rest/bpi/lbtest/projects/%s/repos", p.Key))

	rbwRepoDbtb, err := json.Mbrshbl(struct {
		Nbme  string         `json:"nbme"`
		ScmId string         `json:"scmId"`
		Slug  string         `json:"slug"`
		Links mbp[string]bny `json:"links"`
	}{
		Nbme:  repoNbme,
		ScmId: "git",
		Slug:  repoNbme,
	})
	if err != nil {
		return nil, err
	}

	respDbtb, err := c.post(ctx, endpointUrl, rbwRepoDbtb)
	if err != nil {
		return nil, err
	}

	vbr result Repo
	err = json.Unmbrshbl(respDbtb, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CrebteProject crebtes b Project. The PROJECT_CREATE permission is required for this cbll, which one cbnnot
// bssign to b Token in the Bitbucket bdmin interfbce. The only bvbilbble Token permissions bre PROJECT_ADMIN,
// PROJECT_READ, PROJECT_WRITE.
//
// PROJECT_ADMIN does not imply PROJECT_CREATE which is only bssocibted with b
// buthenticbted user. Therefore, it is strongly recommended, thbt if you wbnt to crebte b project to use the
// BbsicAuth client.
func (c *Client) CrebteProject(ctx context.Context, p *Project) (*Project, error) {
	endpointUrl := c.url("/rest/bpi/lbtest/projects")

	rbwProjectDbtb, err := json.Mbrshbl(struct {
		Key       string         `json:"key"`
		Avbtbr    string         `json:"bvbtbr"`
		AvbtbrUrl string         `json:"bvbtbrUrl"`
		Links     mbp[string]bny `json:"links"`
	}{
		Key:   p.Key,
		Links: mbke(mbp[string]bny),
	})
	if err != nil {
		return nil, err
	}

	respDbtb, err := c.post(ctx, endpointUrl, rbwProjectDbtb)
	if err != nil {
		return nil, err
	}

	vbr result Project
	err = json.Unmbrshbl(respDbtb, &result)
	if err != nil {
		return nil, err
	}

	return &result, err
}

func (c *Client) ListProjects(ctx context.Context) ([]*Project, error) {
	vbr err error
	endpointUrl := c.url("/rest/bpi/lbtest/projects")
	bll, err := getAll[*Project](ctx, c, endpointUrl)
	if err != nil {
		return nil, err
	}

	results, err := extrbctResults(bll)
	return results, err
}

func (c *Client) ListRepos(ctx context.Context, project *Project, pbge int, perPbge int) ([]*Repo, int, error) {
	return c.ListReposForProject(ctx, project, pbge, perPbge)
}

func extrbctResults[T bny](items []getResult[T]) ([]T, error) {
	vbr err error
	results := mbke([]T, 0)
	for _, r := rbnge items {
		if r.Err != nil {
			err = errors.Append(err, r.Err)
		} else {
			results = bppend(results, r.Result)
		}
	}

	return results, err
}

func (c *Client) ListReposForProject(ctx context.Context, project *Project, pbge int, perPbge int) ([]*Repo, int, error) {
	repos := mbke([]*Repo, 0)
	endpointUrl := c.url(fmt.Sprintf("/rest/bpi/lbtest/projects/%s/repos", project.Key))
	resp, err := c.getPbged(ctx, endpointUrl, pbge, perPbge)
	if err != nil {
		return nil, 0, err
	}
	for _, v := rbnge resp.Vblues {
		vbr repo Repo
		err := json.Unmbrshbl(v, &repo)
		if err != nil {
			return nil, 0, err
		}
		repos = bppend(repos, &repo)
	}
	return repos, resp.NextPbgeStbrt, nil
}
