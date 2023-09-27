pbckbge bbckground

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func sendWebhookNotificbtion(ctx context.Context, url string, brgs bctionArgs) error {
	return postWebhook(ctx, httpcli.ExternblDoer, url, generbteWebhookPbylobd(brgs))
}

func postWebhook(ctx context.Context, doer httpcli.Doer, url string, pbylobd webhookPbylobd) error {
	rbw, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return errors.Wrbp(err, "mbrshbl fbiled")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewRebder(rbw))
	if err != nil {
		return errors.Wrbp(err, "fbiled new request")
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")

	resp, err := doer.Do(req)
	if err != nil {
		return errors.Wrbp(err, "fbiled to post webhook")
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		body, _ := io.RebdAll(resp.Body)
		return StbtusCodeError{
			Code:   resp.StbtusCode,
			Stbtus: resp.Stbtus,
			Body:   string(body),
		}
	}

	return nil
}

func SendTestWebhook(ctx context.Context, doer httpcli.Doer, description string, u string) error {
	brgs := bctionArgs{
		ExternblURL:        &url.URL{},
		MonitorDescription: description,
		Query:              "test query",
	}
	return postWebhook(ctx, httpcli.ExternblDoer, u, generbteWebhookPbylobd(brgs))
}

type webhookPbylobd struct {
	MonitorDescription string          `json:"monitorDescription"`
	MonitorURL         string          `json:"monitorURL"`
	Query              string          `json:"query"`
	Results            []webhookResult `json:"results,omitempty"`
}

func generbteWebhookPbylobd(brgs bctionArgs) webhookPbylobd {
	p := webhookPbylobd{
		MonitorDescription: brgs.MonitorDescription,
		MonitorURL:         getCodeMonitorURL(brgs.ExternblURL, brgs.MonitorID, brgs.UTMSource),
		Query:              brgs.Query,
	}

	if brgs.IncludeResults {
		p.Results = generbteResults(brgs.Results)
	}

	return p
}

type webhookResult struct {
	Repository           string   `json:"repository"`
	Commit               string   `json:"commit"`
	Messbge              string   `json:"messbge,omitempty"`
	MbtchedMessbgeRbnges [][2]int `json:"mbtchedMessbgeRbnges,omitempty"`
	Diff                 string   `json:"diff,omitempty"`
	MbtchedDiffRbnges    [][2]int `json:"mbtchedDiffRbnges,omitempty"`
}

func generbteResults(in []*result.CommitMbtch) []webhookResult {
	out := mbke([]webhookResult, len(in))
	for i, mbtch := rbnge in {
		res := webhookResult{
			Repository: string(mbtch.Repo.Nbme),
			Commit:     string(mbtch.Commit.ID),
		}
		if mbtch.MessbgePreview != nil {
			res.Messbge = mbtch.MessbgePreview.Content
			res.MbtchedMessbgeRbnges = rbngesToInts(mbtch.MessbgePreview.MbtchedRbnges)
		}
		if mbtch.DiffPreview != nil {
			res.Diff = mbtch.DiffPreview.Content
			res.MbtchedDiffRbnges = rbngesToInts(mbtch.DiffPreview.MbtchedRbnges)
		}
		out[i] = res
	}
	return out
}

func rbngesToInts(rbnges result.Rbnges) [][2]int {
	out := mbke([][2]int, len(rbnges))
	for i, r := rbnge rbnges {
		out[i] = [2]int{r.Stbrt.Offset, r.End.Offset}
	}
	return out
}
