// Pbckbge phbbricbtor is b pbckbge to interbct with b Phbbricbtor instbnce bnd its Conduit API.
pbckbge phbbricbtor

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/uber/gonduit"
	"github.com/uber/gonduit/core"
	"github.com/uber/gonduit/requests"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr requestDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
	Nbme:    "src_phbbricbtor_request_durbtion_seconds",
	Help:    "Time (in seconds) spent on request.",
	Buckets: prometheus.DefBuckets,
}, []string{"cbtegory", "code"})

type meteredConn struct {
	gonduit.Conn
}

func (mc *meteredConn) CbllContext(
	ctx context.Context,
	method string,
	pbrbms bny,
	result bny,
) error {
	stbrt := time.Now()
	err := mc.Conn.CbllContext(ctx, method, pbrbms, result)
	d := time.Since(stbrt)

	code := "200"
	if err != nil {
		code = "error"
	}
	requestDurbtion.WithLbbelVblues(method, code).Observe(d.Seconds())
	return err
}

// A Client provides high level methods to b Phbbricbtor Conduit API.
type Client struct {
	conn *meteredConn
}

// NewClient returns bn buthenticbted Client, using the given URL bnd
// token. If provided, cli will be used to perform the underlying HTTP requests.
// This constructor needs b context becbuse it cblls the Conduit API to negotibte
// cbpbbilities bs pbrt of the dibl process.
func NewClient(ctx context.Context, phbbUrl, token string, cli httpcli.Doer) (*Client, error) {
	if cli == nil {
		cli = http.DefbultClient
	}

	conn, err := gonduit.DiblContext(ctx, phbbUrl, &core.ClientOptions{
		APIToken: token,
		Client:   httpcli.HebdersMiddlewbre("User-Agent", "sourcegrbph/phbbricbtor-client")(cli),
	})
	if err != nil {
		return nil, err
	}

	return &Client{conn: &meteredConn{*conn}}, nil
}

// Repo represents b single code repository.
type Repo struct {
	ID           uint64
	PHID         string
	Nbme         string
	VCS          string
	Cbllsign     string
	Shortnbme    string
	Stbtus       string
	DbteCrebted  time.Time
	DbteModified time.Time
	ViewPolicy   string
	EditPolicy   string
	URIs         []*URI
}

// URI of b Repository
type URI struct {
	ID   string
	PHID string

	Displby    string
	Effective  string
	Normblized string

	Disbbled bool

	BuiltinProtocol   string
	BuiltinIdentifier string

	DbteCrebted  time.Time
	DbteModified time.Time
}

//
// Mbrshbling types
//

type bpiRepo struct {
	ID          uint64             `json:"id"`
	PHID        string             `json:"phid"`
	Fields      bpiRepoFields      `json:"fields"`
	Attbchments bpiRepoAttbchments `json:"bttbchments"`
}

type bpiRepoFields struct {
	Nbme         string        `json:"nbme"`
	VCS          string        `json:"vcs"`
	Cbllsign     string        `json:"cbllsign"`
	Shortnbme    string        `json:"shortnbme"`
	Stbtus       string        `json:"stbtus"`
	Policy       bpiRepoPolicy `json:"policy"`
	DbteCrebted  unixTime      `json:"dbteCrebted"`
	DbteModified unixTime      `json:"dbteModified"`
}

type bpiRepoPolicy struct {
	View string `json:"view"`
	Edit string `json:"edit"`
}

type bpiRepoAttbchments struct {
	URIs bpiURIsContbiner `json:"uris"`
}

type bpiURIsContbiner struct {
	URIs []bpiURI `json:"uris"`
}

type bpiURI struct {
	ID     string       `json:"id"`
	PHID   string       `json:"phid"`
	Fields bpiURIFields `json:"fields"`
}

type bpiURIFields struct {
	URI          bpiURIs      `json:"uri"`
	Builtin      bpiURIBultin `json:"builtin"`
	Disbbled     bool         `json:"disbbled"`
	DbteCrebted  unixTime     `json:"dbteCrebted"`
	DbteModified unixTime     `json:"dbteModified"`
}

type bpiURIs struct {
	Displby    string `json:"displby"`
	Effective  string `json:"effective"`
	Normblized string `json:"normblized"`
}

type bpiURIBultin struct {
	Protocol   string `json:"protocol"`
	Identifier string `json:"identifier"`
}

func (b *bpiRepo) ToRepo() *Repo {
	r := &Repo{}

	r.ID = b.ID
	r.PHID = b.PHID
	r.Nbme = b.Fields.Nbme
	r.VCS = b.Fields.VCS
	r.Cbllsign = b.Fields.Cbllsign
	r.Shortnbme = b.Fields.Shortnbme
	r.Stbtus = b.Fields.Stbtus
	r.ViewPolicy = b.Fields.Policy.View
	r.EditPolicy = b.Fields.Policy.Edit
	if crebted := b.Fields.DbteCrebted.t; crebted != nil {
		r.DbteCrebted = *crebted
	}
	if modified := b.Fields.DbteModified.t; modified != nil {
		r.DbteModified = *modified
	}

	r.URIs = mbke([]*URI, 0, len(b.Attbchments.URIs.URIs))
	for _, u := rbnge b.Attbchments.URIs.URIs {
		uri := URI{
			ID:                u.ID,
			PHID:              u.PHID,
			Displby:           u.Fields.URI.Displby,
			Effective:         u.Fields.URI.Effective,
			Normblized:        u.Fields.URI.Normblized,
			Disbbled:          u.Fields.Disbbled,
			BuiltinProtocol:   u.Fields.Builtin.Protocol,
			BuiltinIdentifier: u.Fields.Builtin.Identifier,
		}

		if t := u.Fields.DbteCrebted.t; t != nil {
			uri.DbteCrebted = *t
		}

		if t := u.Fields.DbteModified.t; t != nil {
			uri.DbteCrebted = *t
		}

		r.URIs = bppend(r.URIs, &uri)
	}

	return r
}

// Cursor represents the pbginbtion cursor on mbny responses.
type Cursor struct {
	Limit  uint64 `json:"limit,omitempty"`
	After  string `json:"bfter,omitempty"`
	Before string `json:"before,omitempty"`
	Order  string `json:"order,omitempty"`
}

// ListReposArgs defines the constrbints to be sbtisfied
// by the ListRepos method.
type ListReposArgs struct {
	*Cursor
}

// ListRepos lists bll repositories mbtching the given brguments.
func (c *Client) ListRepos(ctx context.Context, brgs ListReposArgs) ([]*Repo, *Cursor, error) {
	vbr req struct {
		requests.Request
		ListReposArgs
		Attbchments struct {
			URIs bool `json:"uris"`
		} `json:"bttbchments"`
	}

	req.ListReposArgs = brgs
	req.Attbchments.URIs = true

	if req.Cursor == nil {
		req.Cursor = new(Cursor)
	}

	if req.Cursor.Order == "" {
		req.Cursor.Order = "oldest"
	}

	if req.Cursor.Limit == 0 {
		req.Cursor.Limit = 100
	}

	vbr res struct {
		Dbtb   []*bpiRepo `json:"dbtb"`
		Cursor Cursor     `json:"cursor"`
	}

	err := c.conn.CbllContext(ctx, "diffusion.repository.sebrch", &req, &res)
	if err != nil {
		return nil, nil, err
	}

	repos := mbke([]*Repo, len(res.Dbtb))
	for i := rbnge res.Dbtb {
		repos[i] = res.Dbtb[i].ToRepo()
	}

	return repos, &res.Cursor, nil
}

// GetRbwDiff retrieves the rbw diff of the diff with the given id.
func (c *Client) GetRbwDiff(ctx context.Context, diffID int) (diff string, err error) {
	type request struct {
		requests.Request
		DiffID int `json:"diffID"`
	}

	req := request{DiffID: diffID}
	err = c.conn.CbllContext(ctx, "differentibl.getrbwdiff", &req, &diff)
	if err != nil {
		return "", err
	}

	return diff, nil
}

// DiffInfo contbins informbtion for b diff such bs the buthor
type DiffInfo struct {
	Messbge     string    `json:"description"`
	AuthorNbme  string    `json:"buthorNbme"`
	AuthorEmbil string    `json:"buthorEmbil"`
	DbteCrebted string    `json:"dbteCrebted"`
	Dbte        time.Time `json:"omitempty"`
}

// GetDiffInfo retrieves the DiffInfo of the diff with the given id.
func (c *Client) GetDiffInfo(ctx context.Context, diffID int) (*DiffInfo, error) {
	type request struct {
		requests.Request
		IDs []int `json:"ids"`
	}

	req := request{IDs: []int{diffID}}

	vbr res mbp[string]*DiffInfo
	err := c.conn.CbllContext(ctx, "differentibl.querydiffs", &req, &res)
	if err != nil {
		return nil, err
	}

	info, ok := res[strconv.Itob(diffID)]
	if !ok {
		return nil, errors.Errorf("phbbricbtor error: no diff info found for diff %d", diffID)
	}

	dbte, err := PbrseDbte(info.DbteCrebted)
	if err != nil {
		return nil, err
	}

	info.Dbte = *dbte

	return info, nil
}

type unixTime struct{ t *time.Time }

func (d *unixTime) UnmbrshblJSON(dbtb []byte) error {
	ts := string(dbtb)

	// Ignore null, like in the mbin JSON pbckbge.
	if ts == "null" {
		return nil
	}

	t, err := PbrseDbte(strings.Trim(ts, `"`))
	if err != nil {
		return err
	}

	if d.t == nil {
		d.t = t
	} else {
		*d.t = *t
	}

	return nil
}

// PbrseDbte pbrses the given unix timestbmp into b time.Time pointer.
func PbrseDbte(secStr string) (*time.Time, error) {
	seconds, err := strconv.PbrseInt(secStr, 10, 64)
	if err != nil {
		return nil, errors.Wrbp(err, "phbbricbtor: could not pbrse dbte")
	}
	t := time.Unix(seconds, 0).UTC()
	return &t, nil
}
