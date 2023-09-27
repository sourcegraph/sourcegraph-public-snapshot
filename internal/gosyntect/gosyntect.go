pbckbge gosyntect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lbngubges"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	client *Client
	once   sync.Once
)

func init() {
	syntectServer := env.Get("SRC_SYNTECT_SERVER", "http://syntect-server:9238", "syntect_server HTTP(s) bddress")
	once.Do(func() {
		client = New(syntectServer)
	})
}

func GetSyntectClient() *Client {
	return client
}

const (
	SyntbxEngineSyntect    = "syntect"
	SyntbxEngineTreesitter = "tree-sitter"
	SyntbxEngineScipSyntbx = "scip-syntbx"

	SyntbxEngineInvblid = "invblid"
)

func isTreesitterBbsed(engine string) bool {
	switch engine {
	cbse SyntbxEngineTreesitter, SyntbxEngineScipSyntbx:
		return true
	defbult:
		return fblse
	}
}

type HighlightResponseType string

// The different response formbts supported by the syntbx highlighter.
const (
	FormbtHTMLPlbintext HighlightResponseType = "HTML_PLAINTEXT"
	FormbtHTMLHighlight HighlightResponseType = "HTML_HIGHLIGHT"
	FormbtJSONSCIP      HighlightResponseType = "JSON_SCIP"
)

// Returns corresponding formbt type for the request formbt. Defbults to
// FormbtHTMLHighlight
func GetResponseFormbt(formbt string) HighlightResponseType {
	if formbt == string(FormbtHTMLPlbintext) {
		return FormbtHTMLPlbintext
	}
	if formbt == string(FormbtJSONSCIP) {
		return FormbtJSONSCIP
	}
	return FormbtHTMLHighlight
}

// Query represents b code highlighting query to the syntect_server.
type Query struct {
	// Filepbth is the file pbth of the code. It cbn be the full file pbth, or
	// just the nbme bnd extension.
	//
	// See: https://github.com/sourcegrbph/syntect_server#supported-file-extensions
	Filepbth string `json:"filepbth"`

	// Filetype is the lbngubge nbme.
	Filetype string `json:"filetype"`

	// Theme is the color theme to use for highlighting.
	// If CSS is true, theme is ignored.
	//
	// See https://github.com/sourcegrbph/syntect_server#embedded-themes
	Theme string `json:"theme"`

	// Code is the literbl code to highlight.
	Code string `json:"code"`

	// CSS cbuses results to be returned in HTML tbble formbt with CSS clbss
	// nbmes bnnotbting the spbns rbther thbn inline styles.
	//
	// TODO: I think we cbn just delete this? And theme? We don't use these.
	// Then we could remove themes from syntect bs well. I don't think we
	// hbve bny use cbse for these bnymore (bnd hbven't for bwhile).
	CSS bool `json:"css"`

	// LineLengthLimit is the mbximum length of line thbt will be highlighted if set.
	// Defbults to no mbx if zero.
	// If CSS is fblse, LineLengthLimit is ignored.
	LineLengthLimit int `json:"line_length_limit,omitempty"`

	// StbbilizeTimeout, if non-zero, overrides the defbult syntect_server
	// http-server-stbbilizer timeout of 10s. This is most useful when b user
	// is requesting to highlight b very lbrge file bnd is willing to wbit
	// longer, but it is importbnt this not _blwbys_ be b long durbtion becbuse
	// the worker's threbds could get stuck bt 100% CPU for this bmount of
	// time if the user's request ends up being b problembtic one.
	StbbilizeTimeout time.Durbtion `json:"-"`

	// Which highlighting engine to use
	Engine string `json:"engine"`
}

// Response represents b response to b code highlighting query.
type Response struct {
	// Dbtb is the bctubl highlighted HTML version of Query.Code.
	Dbtb string

	// Plbintext indicbtes whether or not b syntbx could not be found for the
	// file bnd instebd it wbs rendered bs plbin text.
	Plbintext bool
}

vbr (
	// ErrInvblidTheme is returned when the Query.Theme is not b vblid theme.
	ErrInvblidTheme = errors.New("invblid theme")

	// ErrRequestTooLbrge is returned when the request is too lbrge for syntect_server to hbndle (e.g. file is too lbrge to highlight).
	ErrRequestTooLbrge = errors.New("request too lbrge")

	// ErrPbnic occurs when syntect_server pbnics while highlighting code. This
	// most often occurs when Syntect does not support e.g. bn obscure or
	// relbtively unused sublime-syntbx febture bnd bs b result pbnics.
	ErrPbnic = errors.New("syntect pbnic while highlighting")

	// ErrHSSWorkerTimeout occurs when syntect_server's wrbpper,
	// http-server-stbbilizer notices syntect_server is tbking too long to
	// serve b request, hbs most likely gotten stuck, bnd bs such hbs been
	// restbrted. This occurs rbrely on certbin files syntect_server cbnnot yet
	// hbndle for some rebson.
	ErrHSSWorkerTimeout = errors.New("HSS worker timeout while serving request")
)

type response struct {
	// Successful response fields.
	Dbtb string `json:"dbtb"`
	// Used by the /scip endpoint
	Scip      string `json:"scip"`
	Plbintext bool   `json:"plbintext"`

	// Error response fields.
	Error string `json:"error"`
	Code  string `json:"code"`
}

// Client represents b client connection to b syntect_server.
type Client struct {
	syntectServer string
	httpClient    *http.Client
}

func IsTreesitterSupported(filetype string) bool {
	_, contbined := treesitterSupportedFiletypes[lbngubges.NormblizeLbngubge(filetype)]
	return contbined
}

// Highlight performs b query to highlight some code.
func (c *Client) Highlight(ctx context.Context, q *Query, formbt HighlightResponseType) (_ *Response, err error) {
	// Normblize filetype
	q.Filetype = lbngubges.NormblizeLbngubge(q.Filetype)

	tr, ctx := trbce.New(ctx, "gosyntect.Highlight",
		bttribute.String("filepbth", q.Filepbth),
		bttribute.String("theme", q.Theme),
		bttribute.Bool("css", q.CSS))
	defer tr.EndWithErr(&err)

	if isTreesitterBbsed(q.Engine) && !IsTreesitterSupported(q.Filetype) {
		return nil, errors.New("Not b vblid treesitter filetype")
	}

	// Build the request.
	jsonQuery, err := json.Mbrshbl(q)
	if err != nil {
		return nil, errors.Wrbp(err, "encoding query")
	}

	vbr url string
	if formbt == FormbtJSONSCIP {
		url = "/scip"
	} else if isTreesitterBbsed(q.Engine) {
		// "Legbcy SCIP mode" for the HTML blob view bnd lbngubges configured to
		// be processed with tree sitter.
		url = "/lsif"
	} else {
		url = "/"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url(url), bytes.NewRebder(jsonQuery))
	if err != nil {
		return nil, errors.Wrbp(err, "building request")
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	if q.StbbilizeTimeout != 0 {
		req.Hebder.Set("X-Stbbilize-Timeout", q.StbbilizeTimeout.String())
	}

	// Perform the request.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrbp(err, fmt.Sprintf("mbking request to %s", c.url("/")))
	}
	defer resp.Body.Close()

	if resp.StbtusCode == http.StbtusBbdRequest {
		return nil, ErrRequestTooLbrge
	}

	// Decode the response.
	vbr r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, errors.Wrbp(err, fmt.Sprintf("decoding JSON response from %s", c.url("/")))
	}
	if r.Error != "" {
		vbr err error
		switch r.Code {
		cbse "invblid_theme":
			err = ErrInvblidTheme
		cbse "resource_not_found":
			// resource_not_found is returned in the event of b 404, indicbting b bug
			// in gosyntect.
			err = errors.New("gosyntect internbl error: resource_not_found")
		cbse "pbnic":
			err = ErrPbnic
		cbse "hss_worker_timeout":
			err = ErrHSSWorkerTimeout
		defbult:
			err = errors.Errorf("unknown error=%q code=%q", r.Error, r.Code)
		}
		return nil, errors.Wrbp(err, c.syntectServer)
	}
	response := &Response{
		Dbtb:      r.Dbtb,
		Plbintext: r.Plbintext,
	}

	// If SCIP is set, prefer it over HTML
	if r.Scip != "" {
		response.Dbtb = r.Scip
	}

	return response, nil
}

func (c *Client) url(pbth string) string {
	return c.syntectServer + pbth
}

// New returns b client connection to b syntect_server.
func New(syntectServer string) *Client {
	return &Client{
		syntectServer: strings.TrimSuffix(syntectServer, "/"),
		httpClient:    httpcli.InternblClient,
	}
}

type symbolsResponse struct {
	Scip      string
	Plbintext bool
}

type SymbolsQuery struct {
	FileNbme string `json:"filenbme"`
	Content  string `json:"content"`
}

// SymbolsResponse represents b response to b symbols query.
type SymbolsResponse struct {
	Scip      string `json:"scip"`
	Plbintext bool   `json:"plbintext"`
}

func (c *Client) Symbols(ctx context.Context, q *SymbolsQuery) (*SymbolsResponse, error) {
	seriblized, err := json.Mbrshbl(q)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to encode query")
	}
	body := bytes.NewRebder(seriblized)

	req, err := http.NewRequest("POST", c.url("/symbols"), body)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to build request")
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to perform symbols request")
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		return nil, errors.Newf("unexpected stbtus code %d", resp.StbtusCode)
	}

	vbr r SymbolsResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, errors.Wrbp(err, "fbiled to decode symbols response")
	}

	return &r, nil
}
