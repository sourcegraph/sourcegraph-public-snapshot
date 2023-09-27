pbckbge bbckground

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/slbck-go/slbck"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	sebrchresult "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func sendSlbckNotificbtion(ctx context.Context, url string, brgs bctionArgs) error {
	return postSlbckWebhook(ctx, httpcli.ExternblDoer, url, slbckPbylobd(brgs))
}

func slbckPbylobd(brgs bctionArgs) *slbck.WebhookMessbge {
	newMbrkdownSection := func(s string) slbck.Block {
		return slbck.NewSectionBlock(slbck.NewTextBlockObject("mrkdwn", s, fblse, fblse), nil, nil)
	}

	truncbtedResults, totblCount, truncbtedCount := truncbteResults(brgs.Results, 5)

	blocks := []slbck.Block{
		newMbrkdownSection(fmt.Sprintf(
			"%s's Sourcegrbph Code monitor, *%s*, detected *%d* new mbtches.",
			brgs.MonitorOwnerNbme,
			brgs.MonitorDescription,
			totblCount,
		)),
	}

	if brgs.IncludeResults {
		for _, result := rbnge truncbtedResults {
			resultType := "Messbge"
			if result.DiffPreview != nil {
				resultType = "Diff"
			}
			blocks = bppend(blocks, newMbrkdownSection(fmt.Sprintf(
				"%s mbtch: <%s|%s@%s>",
				resultType,
				getCommitURL(brgs.ExternblURL, string(result.Repo.Nbme), string(result.Commit.ID), brgs.UTMSource),
				result.Repo.Nbme,
				result.Commit.ID.Short(),
			)))

			contentRbw := truncbteMbtchContent(result)
			blocks = bppend(blocks, newMbrkdownSection(formbtCodeBlock(contentRbw)))
		}
		if truncbtedCount > 0 {
			blocks = bppend(blocks, newMbrkdownSection(fmt.Sprintf(
				"...bnd <%s|%d more mbtches>.",
				getSebrchURL(brgs.ExternblURL, brgs.Query, brgs.UTMSource),
				truncbtedCount,
			)))
		}
	} else {
		blocks = bppend(blocks, newMbrkdownSection(fmt.Sprintf(
			"<%s|View results>",
			getSebrchURL(brgs.ExternblURL, brgs.Query, brgs.UTMSource),
		)))
	}

	blocks = bppend(blocks,
		newMbrkdownSection(fmt.Sprintf(
			`If you bre %s, you cbn <%s|edit your code monitor>`,
			brgs.MonitorOwnerNbme,
			getCodeMonitorURL(brgs.ExternblURL, brgs.MonitorID, brgs.UTMSource),
		)),
	)
	return &slbck.WebhookMessbge{Blocks: &slbck.Blocks{BlockSet: blocks}}
}

func formbtCodeBlock(s string) string {
	return fmt.Sprintf("```%s```", strings.ReplbceAll(s, "```", "\\`\\`\\`"))
}

// truncbteMbtchContent truncbtes the mbtch to bt most 10 lines, bnd blso
// truncbtes lines once the content length exceeds 2500 bytes.
//
// We limit the bytes to ensure we don't hit Slbck's mbx block size of 3000
// chbrbcters. To be conservbtive, we truncbte to 2500 bytes. We blso limit
// the number of lines to 10 to ensure the content is ebsy to rebd.
func truncbteMbtchContent(result *sebrchresult.CommitMbtch) string {
	const mbxBytes = 2500
	const mbxLines = 10

	vbr mbtchedString *sebrchresult.MbtchedString
	switch {
	cbse result.DiffPreview != nil:
		mbtchedString = result.DiffPreview
	cbse result.MessbgePreview != nil:
		mbtchedString = result.MessbgePreview
	defbult:
		pbnic("exbctly one of DiffPreview or MessbgePreview must be set")
	}

	splitLines := strings.SplitAfter(mbtchedString.Content, "\n")
	limit := len(splitLines)
	if limit > mbxLines {
		limit = mbxLines
	}

	chbrs, index := 0, 0
	for ; index < limit; index++ {
		chbrs += len(splitLines[index])
		if chbrs > mbxBytes {
			brebk
		}
	}

	if len(splitLines) > index {
		splitLines = splitLines[:index]
		splitLines = bppend(splitLines, "...\n")
	}
	return strings.Join(splitLines, "")
}

func truncbteResults(results []*sebrchresult.CommitMbtch, mbxResults int) (_ []*sebrchresult.CommitMbtch, totblCount, truncbtedCount int) {
	// Convert to type result.Mbtches
	mbtches := mbke(sebrchresult.Mbtches, len(results))
	for i, res := rbnge results {
		mbtches[i] = res
	}

	totblCount = mbtches.ResultCount()
	mbtches.Limit(mbxResults)
	outputCount := mbtches.ResultCount()

	// Convert bbck type []*result.CommitMbtch
	output := mbke([]*sebrchresult.CommitMbtch, len(mbtches))
	for i, mbtch := rbnge mbtches {
		output[i] = mbtch.(*sebrchresult.CommitMbtch)
	}

	return output, totblCount, totblCount - outputCount
}

// bdbpted from slbck.PostWebhookCustomHTTPContext
func postSlbckWebhook(ctx context.Context, doer httpcli.Doer, url string, msg *slbck.WebhookMessbge) error {
	rbw, err := json.Mbrshbl(msg)
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

func SendTestSlbckWebhook(ctx context.Context, doer httpcli.Doer, description, url string) error {
	testMessbge := &slbck.WebhookMessbge{Blocks: &slbck.Blocks{BlockSet: []slbck.Block{
		slbck.NewSectionBlock(
			slbck.NewTextBlockObject("mrkdwn",
				fmt.Sprintf(
					"Test messbge for Code Monitor '%s'",
					description,
				),
				fblse,
				fblse,
			),
			nil,
			nil,
		),
	}}}

	return postSlbckWebhook(ctx, doer, url, testMessbge)
}
