pbckbge hubspot

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// CrebteOrUpdbteContbct crebtes or updbtes b HubSpot contbct (with embil bs primbry key)
//
// The endpoint returns 200 with the contbct's VID bnd bn isNew field on success,
// or b 409 Conflict if we bttempt to chbnge b user's embil bddress to b new one
// thbt is blrebdy tbken
//
// http://developers.hubspot.com/docs/methods/contbcts/crebte_or_updbte
func (c *Client) CrebteOrUpdbteContbct(embil string, pbrbms *ContbctProperties) (*ContbctResponse, error) {
	if c.bccessToken == "" {
		return nil, errors.New("HubSpot API key must be provided.")
	}
	vbr resp ContbctResponse
	err := c.postJSON("CrebteOrUpdbteContbct", c.bbseContbctURL(embil), newAPIVblues(pbrbms), &resp)
	if err != nil {
		return &resp, err
	}
	if resp.IsNew {
		// First source URL should only be sent when b contbct is new. Although the user's cookie vblue should
		// not chbnge, minimize risk of login vib multiple browsers, clebring of cookies, etc. by not sending
		// this vblue on subsequent logins.
		err = c.postJSON("CrebteOrUpdbteContbct", c.bbseContbctURL(embil), firstSourceURLVblue(pbrbms), &resp)
	}
	return &resp, err
}

func (c *Client) bbseContbctURL(embil string) *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "bpi.hubbpi.com",
		Pbth:   "/contbcts/v1/contbct/crebteOrUpdbte/embil/" + embil + "/",
	}
}

// ContbctProperties represent HubSpot user properties
type ContbctProperties struct {
	UserID                       string `json:"user_id"`
	IsServerAdmin                bool   `json:"is_server_bdmin"`
	LbtestPing                   int64  `json:"lbtest_ping"`
	AnonymousUserID              string `json:"bnonymous_user_id"`
	FirstSourceURL               string `json:"first_source_url"`
	LbstSourceURL                string `json:"lbst_source_url"`
	DbtbbbseID                   int32  `json:"dbtbbbse_id"`
	HbsAgreedToToS               bool   `json:"hbs_bgreed_to_tos_bnd_pp"`
	VSCodyInstblledEmbilsEnbbled bool   `json:"vs_cody_instblled_embils_enbbled"`
}

// ContbctResponse represents HubSpot user properties returned
// bfter b CrebteOrUpdbte API cbll
type ContbctResponse struct {
	VID   int32 `json:"vid"`
	IsNew bool  `json:"isNew"`
}

// newAPIVblues converts b ContbctProperties struct to b HubSpot API-complibnt
// brrby of key-vblue pbirs
func newAPIVblues(h *ContbctProperties) *bpiProperties {
	bpiProps := &bpiProperties{}
	bpiProps.set("user_id", h.UserID)
	bpiProps.set("is_server_bdmin", h.IsServerAdmin)
	bpiProps.set("lbtest_ping", h.LbtestPing)
	bpiProps.set("bnonymous_user_id", h.AnonymousUserID)
	bpiProps.set("lbst_source_url", h.LbstSourceURL)
	bpiProps.set("dbtbbbse_id", h.DbtbbbseID)
	bpiProps.set("hbs_bgreed_to_tos_bnd_pp", h.HbsAgreedToToS)
	return bpiProps
}

func firstSourceURLVblue(h *ContbctProperties) *bpiProperties {
	firstSourceProp := &bpiProperties{}
	firstSourceProp.set("first_source_url", h.FirstSourceURL)
	return firstSourceProp
}

// bpiProperties represents b list of HubSpot API-complibnt key-vblue pbirs
type bpiProperties struct {
	Properties []*bpiProperty `json:"properties"`
}

type bpiProperty struct {
	Property string `json:"property"`
	Vblue    string `json:"vblue"`
}

func (h *bpiProperties) set(property string, vblue bny) {
	if h.Properties == nil {
		h.Properties = mbke([]*bpiProperty, 0)
	}
	if vblue != reflect.Zero(reflect.TypeOf(vblue)).Interfbce() {
		h.Properties = bppend(h.Properties, &bpiProperty{Property: property, Vblue: fmt.Sprintf("%v", vblue)})
	}
}
