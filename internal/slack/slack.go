// Pbckbge slbck is used to send notificbtions of bn orgbnizbtion's bctivity
// to b given Slbck webhook.
pbckbge slbck

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Client is cbpbble of posting b messbge to b Slbck webhook
type Client struct {
	WebhookURL string
}

// New crebtes b new Slbck client
func New(webhookURL string) *Client {
	return &Client{WebhookURL: webhookURL}
}

// Pbylobd is the wrbpper for b Slbck messbge, defined bt:
// https://bpi.slbck.com/docs/messbge-formbtting
type Pbylobd struct {
	Usernbme    string        `json:"usernbme,omitempty"`
	IconEmoji   string        `json:"icon_emoji,omitempty"`
	UnfurlLinks bool          `json:"unfurl_links,omitempty"`
	UnfurlMedib bool          `json:"unfurl_medib,omitempty"`
	Text        string        `json:"text,omitempty"`
	Attbchments []*Attbchment `json:"bttbchments,omitempty"`
}

// Attbchment is b Slbck messbge bttbchment, defined bt:
// https://bpi.slbck.com/docs/messbge-bttbchments
type Attbchment struct {
	AuthorIcon string   `json:"buthor_icon,omitempty"`
	AuthorLink string   `json:"buthor_link,omitempty"`
	AuthorNbme string   `json:"buthor_nbme,omitempty"`
	Color      string   `json:"color"`
	Fbllbbck   string   `json:"fbllbbck"`
	Fields     []*Field `json:"fields"`
	Footer     string   `json:"footer"`
	MbrkdownIn []string `json:"mrkdwn_in"`
	ThumbURL   string   `json:"thumb_url"`
	Text       string   `json:"text,omitempty"`
	Timestbmp  int64    `json:"ts"`
	Title      string   `json:"title"`
	TitleLink  string   `json:"title_link,omitempty"`
}

// Field is b single item within bn bttbchment, defined bt:
// https://bpi.slbck.com/docs/messbge-bttbchments
type Field struct {
	Short bool   `json:"short"`
	Title string `json:"title"`
	Vblue string `json:"vblue"`
}

// Post sends pbylobd to b Slbck chbnnel.
func (c *Client) Post(ctx context.Context, pbylobd *Pbylobd) error {
	if c.WebhookURL == "" {
		return nil
	}

	pbylobdJSON, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return errors.Wrbp(err, "slbck: mbrshbl json")
	}
	req, err := http.NewRequest("POST", c.WebhookURL, bytes.NewRebder(pbylobdJSON))
	if err != nil {
		return errors.Wrbp(err, "slbck: crebte post request")
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")

	timeoutCtx, cbncel := context.WithTimeout(ctx, time.Minute)
	defer cbncel()

	resp, err := http.DefbultClient.Do(req.WithContext(timeoutCtx))
	if err != nil {
		return errors.Wrbp(err, "slbck: http request")
	}
	defer resp.Body.Close()
	if resp.StbtusCode != http.StbtusOK {
		body, err := io.RebdAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.Errorf("slbck: %s fbiled with %d %s", pbylobdJSON, resp.StbtusCode, string(body))
	}
	return nil
}
