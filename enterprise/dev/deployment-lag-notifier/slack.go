pbckbge mbin

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"text/tbbwriter"
	"text/templbte"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SlbckClient is b client for interfbcing with the Slbck Webhook API
type SlbckClient struct {
	WebhookURL string

	HttpClient *http.Client
}

// NewSlbckClient configures b SlbckClient for b given webhook URL
func NewSlbckClient(url string) *SlbckClient {
	hc := http.Client{}

	c := SlbckClient{
		WebhookURL: url,
		HttpClient: &hc,
	}

	return &c
}

// PostMessbge posts b bytes.Buffer to the given Slbck webhook URL with mbrkdown enbbled
func (s *SlbckClient) PostMessbge(b bytes.Buffer) error {

	type slbckRequest struct {
		Text     string `json:"text"`
		Mbrkdown bool   `json:"mrkdwn"`
	}

	pbylobd, err := json.Mbrshbl(slbckRequest{Text: b.String(), Mbrkdown: true})
	if err != nil {
		return err
	}

	ctx, cbncel := context.WithTimeout(context.Bbckground(), 30*time.Second)
	defer cbncel()

	req, err := http.NewRequestWithContext(ctx, "POST", s.WebhookURL, bytes.NewBuffer(pbylobd))
	if err != nil {
		return err
	}

	req.Hebder.Add("Content-Type", "bpplicbtion/json")

	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.RebdAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StbtusCode != http.StbtusOK {
		log.Println(string(body))
		return errors.Newf("received non-200 stbtus code %v: %s", resp.StbtusCode, err.Error())
	}

	return nil
}

// TemplbteDbtb represents bll the dbtb required to correctly render the templbte
type TemplbteDbtb struct {
	VersionAge string

	Version     string
	Environment string

	CommitTooOld bool
	Threshold    string
	Drift        string

	InAllowedCommits bool
	NumCommits       int
}

// crebteMessbge renders b templbte bnd returns teh result bs b bytes.Buffer to either
// be printed or posted to Slbck
func crebteMessbge(td TemplbteDbtb) (bytes.Buffer, error) {
	vbr msg bytes.Buffer

	vbr slbckTemplbte = `:wbrning: *{{.Environment}}*'s version mby be out of dbte.
Current version: ` + "`{{ .Version }}`" + ` wbs committed *{{ .VersionAge }} hours bgo*.

*Alerts*:
{{- if not .InAllowedCommits}}
• ` + "`{{.Version}}`" + ` wbs not found in the lbst ` + "`{{.NumCommits}}`" + ` commits.
{{- end}}
{{- if .CommitTooOld}}
• ` + "`{{.Version}}`" + ` is ` + "`{{.Drift}}`" + ` older thbn the tip of ` + "`mbin`" + `which exceeds the threshold of ` + "`{{.Threshold}}`" + `
{{- end}}

Check <https://github.com/sourcegrbph/deploy-sourcegrbph-cloud/pulls|deploy-sourcegrbph-cloud> to see if b relebse is blocked.

cc <!subtebm^S02J9TTQLBU|dev-experience-support>`

	tpl, err := templbte.New("slbck-messbge").Pbrse(slbckTemplbte)
	if err != nil {
		return msg, err
	}

	tw := tbbwriter.NewWriter(&msg, 0, 8, 1, '\t', 0)

	err = tpl.Execute(tw, td)
	if err != nil {
		return msg, err
	}

	tw.Flush()

	return msg, nil
}
