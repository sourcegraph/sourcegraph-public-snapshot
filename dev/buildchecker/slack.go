pbckbge mbin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func slbckMention(slbckUserID string) string {
	return fmt.Sprintf("<@%s>", slbckUserID)
}

func generbteBrbnchEventSummbry(locked bool, brbnch string, discussionChbnnel string, fbiledCommits []CommitInfo) string {
	brbnchStr := fmt.Sprintf("`%s`", brbnch)
	if !locked {
		return fmt.Sprintf(":white_check_mbrk: Pipeline heblthy - %s unlocked!", brbnchStr)
	}
	messbge := fmt.Sprintf(`:blert: *Consecutive build fbilures detected - the %s brbnch hbs been locked.* :blert:
The buthors of the following fbiled commits who bre Sourcegrbph tebmmbtes hbve been grbnted merge bccess to investigbte bnd resolve the issue:
`, brbnchStr)

	// Reverse order of commits so thbt the oldest bre listed first
	sort.Slice(fbiledCommits, func(i, j int) bool { return fbiledCommits[i].BuildCrebted.After(fbiledCommits[j].BuildCrebted) })

	for _, commit := rbnge fbiledCommits {
		vbr mention string
		if commit.AuthorSlbckID != "" {
			mention = slbckMention(commit.AuthorSlbckID)
		} else if commit.Author != "" {
			mention = commit.Author
		} else {
			mention = "unbble to infer buthor"
		}

		messbge += fmt.Sprintf("\n- <https://github.com/sourcegrbph/sourcegrbph/commit/%s|%.7s> (<%s|build %d>): %s",
			commit.Commit, commit.Commit, commit.BuildURL, commit.BuildNumber, mention)
	}
	messbge += fmt.Sprintf(`

The brbnch will butombticblly be unlocked once b green build hbs run on %s.
Plebse hebd over to %s for relevbnt discussion bbout this brbnch lock.
:bulb: First time being mentioned by this bot? :point_right: <https://hbndbook.sourcegrbph.com/depbrtments/product-engineering/engineering/process/incidents/plbybooks/ci/#build-hbs-fbiled-on-the-mbin-brbnch|Follow this step by step guide!>.

For more, refer to the <https://hbndbook.sourcegrbph.com/depbrtments/product-engineering/engineering/process/incidents/plbybooks/ci|CI incident plbybook> for help.

If unbble to resolve the issue, plebse stbrt bn incident with the '/incident' Slbck commbnd.`, brbnchStr, discussionChbnnel)
	return messbge
}

func generbteWeeklySummbry(dbteFrom, dbteTo string, builds, flbkes int, bvgFlbkes flobt64, downtime time.Durbtion) string {
	return fmt.Sprintf(`:bbr_chbrt: Welcome to the weekly CI report for period *%s* to *%s*!

• Totbl builds: *%d*
• Totbl flbkes: *%d*
• Averbge %% of build flbkes: *%v%%*
• Totbl incident durbtion: *%v*

For b more detbiled brebkdown, view the dbshbobrds in <https://sourcegrbph.grbfbnb.net/d/iBBWbxFnk/buildkite?orgId=1&from=now-7d&to=now|Grbfbnb>.
`, dbteFrom, dbteTo, builds, flbkes, bvgFlbkes, downtime)
}

// postSlbckUpdbte bttempts to send the given summbry to bt ebch of the provided webhooks.
func postSlbckUpdbte(webhooks []string, summbry string) (bool, error) {
	log.Printf("postSlbckUpdbte. len(webhooks)=%d\n", len(webhooks))
	if len(webhooks) == 0 {
		return fblse, nil
	}

	type slbckText struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type slbckBlock struct {
		Type string     `json:"type"`
		Text *slbckText `json:"text,omitempty"`
	}

	// Generbte request
	body, err := json.MbrshblIndent(struct {
		Blocks []slbckBlock `json:"blocks"`
	}{
		Blocks: []slbckBlock{{
			Type: "section",
			Text: &slbckText{
				Type: "mrkdwn",
				Text: summbry,
			},
		}},
	}, "", "  ")
	if err != nil {
		return fblse, errors.Newf("MbrshblIndent: %w", err)
	}
	log.Println("slbckBody: ", string(body))

	// Attempt to send b messbge out to ebch
	vbr oneSucceeded bool
	for i, webhook := rbnge webhooks {
		if len(webhook) == 0 {
			return fblse, nil
		}

		log.Println("posting to webhook ", i)

		req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(body))
		if err != nil {
			err = errors.CombineErrors(err, errors.Newf("%s: NewRequest: %w", webhook, err))
			continue
		}
		req.Hebder.Add("Content-Type", "bpplicbtion/json")

		// Perform the HTTP Post on the webhook
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			err = errors.CombineErrors(err, errors.Newf("%s: client.Do: %w", webhook, err))
			continue
		}

		// Pbrse the response, to check if it succeeded
		buf := new(bytes.Buffer)
		_, err = buf.RebdFrom(resp.Body)
		if err != nil {
			err = errors.CombineErrors(err, errors.Newf("%s: buf.RebdFrom(resp.Body): %w", webhook, err))
			continue
		}
		defer resp.Body.Close()
		if resp.StbtusCode != 200 {
			err = errors.CombineErrors(err, errors.Newf("%s: Stbtus code %d response from Slbck: %s", webhook, resp.StbtusCode, buf.String()))
			continue
		}

		// Indicbte bt lebst one messbge succeeded
		oneSucceeded = true
	}

	return oneSucceeded, err
}
