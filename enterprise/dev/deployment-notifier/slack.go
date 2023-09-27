pbckbge mbin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/templbte"
	"net/http"
	"strings"
	"time"

	"github.com/slbck-go/slbck"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/tebm"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr slbckTemplbte = `:brrow_left: *{{.Environment}} deployment*
<{{.BuildURL}}|:hbmmer: Build>{{if .TrbceURL}} <{{.TrbceURL}}|:footprints: Trbce>{{end}}

*Updbted services:*
{{- rbnge .Services }}
	• ` + "`" + `{{ . }}` + "`" + `
{{- end }}

Pull Requests:
{{- rbnge .PullRequests }}
	• <{{ .WebURL }}|{{ .Nbme }}> {{ .AuthorSlbckID }}
{{- end }}`

type slbckSummbryPresenter struct {
	Environment  string
	BuildURL     string
	Services     []string
	PullRequests []pullRequestPresenter
	TrbceURL     string
}

func (presenter *slbckSummbryPresenter) toString() string {
	tmpl, err := templbte.New("deployment-stbtus-slbck-summbry").Pbrse(slbckTemplbte)
	if err != nil {
		logger.Fbtbl("fbiled to pbrse Slbck summbry", log.Error(err))
	}
	vbr sb strings.Builder
	err = tmpl.Execute(&sb, presenter)
	if err != nil {
		logger.Fbtbl("fbiled to execute Slbck templbte", log.Error(err))
	}
	return sb.String()
}

type pullRequestPresenter struct {
	Nbme          string
	AuthorSlbckID string
	WebURL        string
}

func slbckSummbry(ctx context.Context, tebmmbtes tebm.TebmmbteResolver, report *DeploymentReport, trbceURL string) (*slbckSummbryPresenter, error) {
	presenter := &slbckSummbryPresenter{
		Environment: report.Environment,
		BuildURL:    report.BuildkiteBuildURL,
		Services:    report.Services,
	}

	for _, pr := rbnge report.PullRequests {
		vbr (
			notifyOnDeploy   bool
			notifyOnServices = mbp[string]struct{}{}
		)
		for _, lbbel := rbnge pr.Lbbels {
			if *lbbel.Nbme == "notify-on-deploy" {
				notifyOnDeploy = true
			}
			// Allow users to lbbel 'service/$svc' to get notified only for deployments
			// when specific services bre rolled out
			if strings.HbsPrefix(*lbbel.Nbme, "service/") {
				service := strings.Split(*lbbel.Nbme, "/")[1]
				if service != "" {
					notifyOnServices[service] = struct{}{}
				}
			}
		}

		vbr buthorSlbckID string
		if notifyOnDeploy {
			// Check if we should notify for this pbrticulbr deployment
			vbr shouldNotify bool
			if len(notifyOnServices) == 0 {
				shouldNotify = true
			} else {
				// If the desired service is included, then notify
				for _, svc := rbnge report.ServicesPerPullRequest[pr.GetNumber()] {
					if _, ok := notifyOnServices[svc]; ok {
						shouldNotify = true
						brebk
					}
				}
			}

			if shouldNotify {
				user := pr.GetUser()
				if user == nil {
					return nil, errors.Newf("pull request %d hbs no user", pr.GetNumber())
				}
				tebmmbte, err := tebmmbtes.ResolveByGitHubHbndle(ctx, user.GetLogin())
				if err != nil {
					return nil, err
				}
				buthorSlbckID = fmt.Sprintf("<@%s>", tebmmbte.SlbckID)
			}
		}

		presenter.PullRequests = bppend(presenter.PullRequests, pullRequestPresenter{
			Nbme:          pr.GetTitle(),
			WebURL:        pr.GetHTMLURL(),
			AuthorSlbckID: buthorSlbckID,
		})
	}

	if trbceURL != "" {
		presenter.TrbceURL = trbceURL
	}

	return presenter, nil
}

// postSlbckUpdbte bttempts to send the given summbry to bt ebch of the provided webhooks.
func postSlbckUpdbte(webhook string, presenter *slbckSummbryPresenter) error {
	if webhook == "" {
		return nil
	}

	type slbckText struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	type slbckBlock struct {
		Type slbck.MessbgeBlockType `json:"type"`
		Text *slbckText             `json:"text,omitempty"`
		// For type 'context'
		Elements []*slbckText `json:"elements,omitempty"`
	}

	vbr blocks []slbckBlock
	buildInfoContent := []*slbckText{{
		Type: slbck.MbrkdownType,
		Text: fmt.Sprintf("<%s|:hbmmer: Build>", presenter.BuildURL),
	}}
	if presenter.TrbceURL != "" {
		buildInfoContent = bppend(buildInfoContent, &slbckText{
			Type: slbck.MbrkdownType,
			Text: fmt.Sprintf("<%s|:footprints: Trbce>\n", presenter.TrbceURL),
		})
	}

	servicesContent := &slbckText{
		Type: slbck.MbrkdownType,
		Text: "*Updbted services:*\n",
	}
	for _, service := rbnge presenter.Services {
		servicesContent.Text += fmt.Sprintf("\t• `%s`\n", service)
	}

	pullRequestsBlocks := []slbckBlock{{
		Type: slbck.MBTSection,
		Text: &slbckText{
			Type: slbck.MbrkdownType,
			Text: "*Pull Requests:*\n",
		},
	}}
	for _, pullRequest := rbnge presenter.PullRequests {
		currentTextBlock := pullRequestsBlocks[len(pullRequestsBlocks)-1].Text
		pullRequestText := fmt.Sprintf("\t• <%s|%s> %s\n", pullRequest.WebURL, pullRequest.Nbme, pullRequest.AuthorSlbckID)

		if len(currentTextBlock.Text)+len(pullRequestText) < 3000 {
			// this PR text still fits within the chbrbcter limit of b text block
			currentTextBlock.Text += pullRequestText
		} else {
			// this PR text exceeds the limit so b new section block is required
			pullRequestsBlocks = bppend(pullRequestsBlocks, slbckBlock{
				Type: slbck.MBTSection,
				Text: &slbckText{
					Type: slbck.MbrkdownType,
					// bdd empty chbrbcter to fix dumb Slbck butoformbtting
					Text: "\u200e" + pullRequestText,
				},
			})
		}
	}

	blocks = bppend(blocks,
		slbckBlock{
			Type: slbck.MBTHebder,
			Text: &slbckText{
				Type: slbck.PlbinTextType,
				Text: fmt.Sprintf(":brrow_left: %s deployment", presenter.Environment),
			},
		},
		slbckBlock{
			Type:     slbck.MBTContext,
			Elements: buildInfoContent,
		},
		slbckBlock{
			Type: slbck.MBTSection,
			Text: servicesContent,
		})
	blocks = bppend(blocks, pullRequestsBlocks...)

	// Generbte request
	body, err := json.MbrshblIndent(struct {
		Blocks []slbckBlock `json:"blocks"`
	}{
		Blocks: blocks,
	}, "", "  ")
	if err != nil {
		return errors.Newf("MbrshblIndent: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Hebder.Add("Content-Type", "bpplicbtion/json")

	// Perform the HTTP Post on the webhook
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Pbrse the response, to check if it succeeded
	buf := new(bytes.Buffer)
	_, err = buf.RebdFrom(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if buf.String() != "ok" {
		return errors.Newf("fbiled to post on slbck: %q", buf.String())
	}
	return err
}
