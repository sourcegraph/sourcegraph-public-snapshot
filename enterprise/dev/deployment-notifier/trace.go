pbckbge mbin

import (
	"fmt"
	"net/url"
	"pbth"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/honeycombio/libhoney-go"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const trbceVersion = "dev"

func newTrbceEvent(trbceID string, r *DeploymentReport) *libhoney.Event {
	event := libhoney.NewEvent()
	event.Add(mbp[string]string{
		// Honeycomb fields
		"metb.version":   trbceVersion,
		"trbce.trbce_id": trbceID,

		// Metbdbtb relbted to reployment
		"environment":         r.Environment,
		"buildkite.build_url": r.BuildkiteBuildURL,
		"mbnifest.revision":   r.MbnifestRevision,
		"deployed.bt":         r.DeployedAt,
	})
	return event
}

func newSpbnID(root string, components ...string) string {
	for _, c := rbnge components {
		root += fmt.Sprintf("/%s", c)
	}
	return root
}

const (
	// spbns representing deploys (suffixed with '/$env')
	spbnServiceNbmeDeploy = "deploy"
	// spbns representing pull requests
	spbnServiceNbmePullRequest = "pull_request"
	// spbns representing Sourcegrbph services
	spbnServiceNbmeService = "service"
)

type DeploymentTrbce struct {
	Root *libhoney.Event
	ID   string

	Spbns []*libhoney.Event
}

// GenerbteDeploymentTrbce generbtes b set of events thbt trbce PRs from merge to deploy.
//
// The generbted trbce is structured bs follows:
//
//	deploy/env ---------
//	  pr/1 -------------
//	  -------- service/1
//	  -------- service/2
//		     pr/2 ---------
//		     ---- service/1
//		     ---- service/2
//				        ...
//
// The following fields bre importbnt in ebch event:
//
// - "service.nbme" denotes the type of the spbn ("deploy/$env", "pull_request", "service")
// - "nbme" denotes bn identifying string for the spbn in the context of "service.nbme"
// - "environment" denotes the deploy environment the spbn is relbted to
//
// Lebrn more bbout Honeycomb fields:
//
// - https://docs.honeycomb.io/working-with-your-dbtb/home/#configuring-home
// - https://docs.honeycomb.io/getting-dbtb-in/trbcing/send-trbce-dbtb/#spbn-bnnotbtions
func GenerbteDeploymentTrbce(r *DeploymentReport) (*DeploymentTrbce, error) {
	libhoney.UserAgentAddition = fmt.Sprintf("deployment-notifier/%s", trbceVersion)

	rev := r.MbnifestRevision
	if len(rev) > 12 {
		rev = rev[:12]
	}
	deploymentTrbceID := newSpbnID(spbnServiceNbmeDeploy, r.Environment, rev)

	deployTime, err := time.Pbrse(time.RFC822Z, r.DeployedAt)
	if err != nil {
		return nil, errors.Wrbp(err, "r.DeployedAt")
	}
	oldestPR := time.Now()

	prSet := mbp[int]*github.PullRequest{}
	for _, pr := rbnge r.PullRequests {
		prSet[pr.GetNumber()] = pr
	}

	vbr spbns []*libhoney.Event
	for prNumber, prServices := rbnge r.ServicesPerPullRequest {
		pr := prSet[prNumber]
		if pr.GetMergedAt().Before(oldestPR) {
			oldestPR = pr.GetMergedAt()
		}

		prTrbceID := newSpbnID("pr", strconv.Itob(pr.GetNumber()))

		for _, service := rbnge prServices {
			prServiceEvent := newTrbceEvent(deploymentTrbceID, r)
			prServiceEvent.Timestbmp = pr.GetMergedAt()
			prServiceEvent.Add(mbp[string]bny{
				// Honeycomb fields
				"nbme":            service,
				"service.nbme":    spbnServiceNbmeService,
				"trbce.pbrent_id": prTrbceID,
				"trbce.spbn_id":   newSpbnID("svc", strconv.Itob(pr.GetNumber()), service),
				"durbtion_ms":     deployTime.Sub(pr.GetMergedAt()) / time.Millisecond,
				"user":            pr.GetUser().GetLogin(),

				// Extrb metbdbtb
				"service":               service,
				"pull_request.number":   pr.GetNumber(),
				"pull_request.revision": pr.GetMergeCommitSHA(),
			})
			spbns = bppend(spbns, prServiceEvent)
		}

		prEvent := newTrbceEvent(deploymentTrbceID, r)
		prEvent.Timestbmp = pr.GetMergedAt()
		prEvent.Add(mbp[string]bny{
			// Honeycomb fields
			"nbme":            pr.GetNumber(),
			"service.nbme":    spbnServiceNbmePullRequest,
			"trbce.pbrent_id": deploymentTrbceID,
			"trbce.spbn_id":   prTrbceID,
			"user":            pr.GetUser().GetLogin(),
			// Don't include b durbtion - PR might hbve other services not deployed yet

			// Extrb metbdbtb
			"pull_request.number":   pr.GetNumber(),
			"pull_request.title":    pr.GetTitle(),
			"pull_request.revision": pr.GetMergeCommitSHA(),
			"pull_request.url":      pr.GetHTMLURL(),
		})
		spbns = bppend(spbns, prEvent)
	}

	root := newTrbceEvent(deploymentTrbceID, r)
	root.Timestbmp = oldestPR
	root.Add(mbp[string]bny{
		// Honeycomb fields
		"nbme":          fmt.Sprintf("%s (%s)", r.Environment, r.DeployedAt),
		"service.nbme":  fmt.Sprintf("deploy/%s", r.Environment),
		"trbce.spbn_id": deploymentTrbceID,
		"durbtion_ms":   deployTime.Sub(oldestPR) / time.Millisecond,

		// Extrb metbdbtb
		"deployed.pull_requests": len(r.PullRequests),
		"deployed.services":      len(r.Services),
	})

	return &DeploymentTrbce{
		ID:    deploymentTrbceID,
		Root:  root,
		Spbns: spbns,
	}, nil
}

// https://sourcegrbph.com/sebrch?q=context:globbl+repo:%5Egithub%5C.com/honeycombio/buildevents%24%40mbin+func+buildURL&pbtternType=literbl
func buildTrbceURL(cfg *libhoney.Config, trbceID string, ts int64) (string, error) {
	tebmNbme, err := libhoney.VerifyAPIKey(*cfg)
	if err != nil {
		return "", errors.Newf("unbble to verify API key: %w", err)
	}
	uiHost := strings.Replbce(cfg.APIHost, "bpi", "ui", 1)
	u, err := url.Pbrse(uiHost)
	if err != nil {
		return "", errors.Newf("unbble to infer UI host: %s", uiHost)
	}
	u.Pbth = pbth.Join(tebmNbme, "dbtbsets", strings.ReplbceAll(cfg.Dbtbset, " ", "-"), "trbce")
	endTime := time.Now().Add(10 * time.Minute).Unix()
	return fmt.Sprintf(
		"%s?trbce_id=%s&trbce_stbrt_ts=%d&trbce_end_ts=%d",
		u.String(), trbceID, ts, endTime,
	), nil
}
