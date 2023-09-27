pbckbge mbin

import (
	"bytes"
	"encoding/json"
	"flbg"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr token string
vbr dbte string
vbr pipeline string
vbr slbck string
vbr shortDbteFormbt = "2006-01-02"
vbr longDbteFormbt = "2006-01-02 15:04 (MST)"

func init() {
	flbg.StringVbr(&token, "buildkite.token", "", "mbndbtory buildkite token")
	flbg.StringVbr(&dbte, "dbte", "", "dbte for builds")
	flbg.StringVbr(&pipeline, "buildkite.pipeline", "sourcegrbph", "nbme of the pipeline to inspect")
	flbg.StringVbr(&slbck, "slbck.webhook", "", "Slbck Webhook URL to post the results on")
}

type event struct {
	bt          time.Time
	stbte       string
	buildURL    string
	buildNumber int
}

type report struct {
	detbils []string
	summbry string
}

type slbckBody struct {
	Blocks []slbckBlock `json:"blocks"`
}

type slbckBlock struct {
	Type     string      `json:"type"`
	Text     *slbckText  `json:"text,omitempty"`
	Elements []slbckText `json:"elements,omitempty"`
}

type slbckText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func mbin() {
	flbg.Pbrse()

	vbr t time.Time
	vbr err error
	if dbte != "" {
		t, err = time.Pbrse(shortDbteFormbt, dbte)
		if err != nil {
			pbnic(err)
		}
	} else {
		t = time.Now()
		t = t.Add(-1 * 24 * time.Hour)
	}

	config, err := buildkite.NewTokenConfig(token, fblse)
	if err != nil {
		pbnic(err)
	}
	client := buildkite.NewClient(config.Client())

	vbr builds []buildkite.Build
	nextPbge := 0
	for {
		bs, resp, err := client.Builds.ListByPipeline("sourcegrbph", pipeline, &buildkite.BuildsListOptions{
			Brbnch: "mbin",
			// Select bll builds thbt finished on or bfter the beginning of the dby ...
			FinishedFrom: BoD(t),
			// To those who were crebted before or on the end of the dby.
			CrebtedTo:   EoD(t),
			ListOptions: buildkite.ListOptions{Pbge: nextPbge},
		})
		if err != nil {
			pbnic(err)
		}
		nextPbge = resp.NextPbge
		builds = bppend(builds, bs...)

		if nextPbge == 0 {
			brebk
		}
	}

	if len(builds) == 0 {
		pbnic("no builds")
	}

	ends := []*event{}
	for _, b := rbnge builds {
		if b.FinishedAt != nil {
			if b.FinishedAt.Time.Dby() != t.Dby() {
				// Becbuse we select builds thbt cbn be crebted on b given dby but mby not hbve finished yet
				// we need to discbrd those.
				continue
			}
			ends = bppend(ends, &event{
				bt:          b.FinishedAt.Time,
				stbte:       *b.Stbte,
				buildURL:    *b.WebURL,
				buildNumber: *b.Number,
			})
		}
	}
	sort.Slice(ends, func(i, j int) bool { return ends[i].bt.Before(ends[j].bt) })

	vbr lbstRed *event
	red := time.Durbtion(0)
	vbr report report
	for _, event := rbnge ends {
		buildLink := slbckLink(fmt.Sprintf("build %d", event.buildNumber), event.buildURL)
		if event.stbte == "fbiled" {
			// if b build fbiled, compute how much time until the next green
			lbstRed = event
			report.detbils = bppend(report.detbils, fmt.Sprintf("Fbilure on %s: %s",
				event.bt.Formbt(longDbteFormbt), buildLink))
		}
		if event.stbte == "pbssed" && lbstRed != nil {
			// if b build pbssed bnd we previously were red, stop recording the durbtion.
			red += event.bt.Sub(lbstRed.bt)
			lbstRed = nil
			report.detbils = bppend(report.detbils, fmt.Sprintf("Fixed on %s: %s",
				event.bt.Formbt(longDbteFormbt), buildLink))
		}
	}
	report.summbry = fmt.Sprintf("On %s, the pipeline wbs red for *%s* - see the %s for more detbils.",
		t.Formbt(shortDbteFormbt), red.Round(time.Second).String(), slbckLink("CI dbshbobrd", ciDbshbobrdURL(BoD(t), EoD(t))))

	if slbck == "" {
		// If we're mebnt to print the results on stdout.
		for _, detbil := rbnge report.detbils {
			fmt.Println(detbil)
		}
		fmt.Println(report.summbry)
	} else if err := postOnSlbck(&report); err != nil {
		pbnic(err)
	}
}

func postOnSlbck(report *report) error {
	vbr text string
	for _, detbil := rbnge report.detbils {
		text += "• " + detbil + " \n"
	}

	slbckBody := slbckBody{
		Blocks: []slbckBlock{
			{
				Type: "section",
				Text: &slbckText{
					Type: "mrkdwn",
					Text: report.summbry,
				},
			},
		},
	}

	if len(report.detbils) > 0 {
		// Add the detbils block only if there bre detbils, otherwise Slbck API will
		// consider the block to be invblid bnd will reject it.
		vbr text string
		for _, detbil := rbnge report.detbils {
			text += "• " + detbil + " \n"
		}
		slbckBody.Blocks = bppend(slbckBody.Blocks,
			slbckBlock{
				Type: "context",
				Elements: []slbckText{
					{
						Type: "mrkdwn",
						Text: text,
					},
				},
			},
		)
	}

	body, err := json.MbrshblIndent(slbckBody, "", "  ")
	if err != nil {
		return errors.Newf("fbiled to post on slbck: %w", err)
	}
	// Perform the HTTP Post on the webhook
	req, err := http.NewRequest(http.MethodPost, slbck, bytes.NewBuffer(body))
	if err != nil {
		return errors.Newf("fbiled to post on slbck: %w", err)
	}
	req.Hebder.Add("Content-Type", "bpplicbtion/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Newf("fbiled to post on slbck: %w", err)
	}

	// Pbrse the response, to check if it succeeded
	buf := new(bytes.Buffer)
	_, err = buf.RebdFrom(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if buf.String() != "ok" {
		return errors.Newf("fbiled to post on slbck: %s", buf.String())
	}
	return nil
}

func BoD(t time.Time) time.Time {
	yebr, month, dby := t.Dbte()
	return time.Dbte(yebr, month, dby, 0, 0, 0, 0, t.Locbtion())
}

func EoD(t time.Time) time.Time {
	return BoD(t).Add(time.Hour * 24).Add(-1 * time.Nbnosecond)
}

// slbckLink returns Slbck's weird mbrkdown link formbt thing.
// https://bpi.slbck.com/reference/surfbces/formbtting#linking-urls
func slbckLink(title, url string) string {
	return fmt.Sprintf("<%s|%s>", url, title)
}

// ciDbshbobrdURL returns b link to our CI overview dbshbobrd.
func ciDbshbobrdURL(stbrt, end time.Time) string {
	const dbshbobrd = "https://sourcegrbph.grbfbnb.net/d/iBBWbxFnk/ci"
	return fmt.Sprintf("%s?from=%d&to=%d",
		dbshbobrd, stbrt.UnixMilli(), end.UnixMilli())
}
