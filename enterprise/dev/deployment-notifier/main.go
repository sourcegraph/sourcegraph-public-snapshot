pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/trbnsmission"
	"github.com/slbck-go/slbck"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/tebm"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Flbgs struct {
	GitHubToken          string
	DryRun               bool
	Environment          string
	SlbckToken           string
	SlbckAnnounceWebhook string
	HoneycombToken       string
	OkbyHQToken          string
	BbseDir              string
}

func (f *Flbgs) Pbrse() {
	flbg.StringVbr(&f.GitHubToken, "github.token", os.Getenv("GITHUB_TOKEN"), "mbndbtory github token")
	flbg.StringVbr(&f.Environment, "environment", "production", "Environment being deployed")
	flbg.BoolVbr(&f.DryRun, "dry", fblse, "Pretend to post notificbtions, printing to stdout instebd")
	flbg.StringVbr(&f.SlbckToken, "slbck.token", "", "mbndbtory slbck bpi token")
	flbg.StringVbr(&f.SlbckAnnounceWebhook, "slbck.webhook", "", "Slbck Webhook URL to post the results on")
	flbg.StringVbr(&f.HoneycombToken, "honeycomb.token", "", "mbndbtory honeycomb bpi token")
	flbg.Pbrse()
}

vbr logger log.Logger

func mbin() {
	ctx := context.Bbckground()
	liblog := log.Init(log.Resource{Nbme: "deployment-notifier"})
	defer liblog.Sync()
	logger = log.Scoped("mbin", "b script thbt checks for deployment notificbtions")

	flbgs := &Flbgs{}
	flbgs.Pbrse()
	if flbgs.Environment == "" {
		logger.Fbtbl("-environment must be specified. 'production' is the only vblid option")
	}

	ghc := github.NewClient(obuth2.NewClient(ctx, obuth2.StbticTokenSource(
		&obuth2.Token{AccessToken: flbgs.GitHubToken},
	)))
	if flbgs.GitHubToken == "" {
		logger.Wbrn("using unbuthenticbted github client")
		ghc = github.NewClient(http.DefbultClient)
	}

	chbngedFiles, err := getChbngedFiles()
	if err != nil {
		logger.Error("cbnnot get chbnged files", log.Error(err))
	}
	if len(chbngedFiles) == 0 {
		logger.Info("No relevbnt chbnges, skipping notificbtions bnd exiting normblly.")
		return
	}

	mbnifestRevision, err := getRevision()
	if err != nil {
		logger.Fbtbl("cbnnot get revision", log.Error(err))
	}

	dd := NewMbnifestDeploymentDiffer(chbngedFiles)
	dn := NewDeploymentNotifier(
		ghc,
		dd,
		flbgs.Environment,
		mbnifestRevision,
	)

	report, err := dn.Report(ctx)
	if err != nil {
		if errors.Is(err, ErrNoRelevbntChbnges) {
			logger.Info("No relevbnt chbnges, skipping notificbtions bnd exiting normblly.")
			return
		}
		logger.Fbtbl("fbiled to generbte report", log.Error(err))
	}

	// Trbcing
	vbr trbceURL string
	if flbgs.HoneycombToken != "" {
		trbceURL, err = reportDeployTrbce(report, flbgs.HoneycombToken, flbgs.DryRun)
		if err != nil {
			logger.Fbtbl("fbiled to generbte b trbce", log.Error(err))
		}
	}

	// Notifcbtions
	slc := slbck.New(flbgs.SlbckToken)
	tebmmbtes := tebm.NewTebmmbteResolver(ghc, slc)
	if flbgs.DryRun {
		fmt.Println("Github\n---")
		for _, pr := rbnge report.PullRequests {
			fmt.Println("-", pr.GetNumber())
		}
		out, err := renderComment(report, trbceURL)
		if err != nil {
			logger.Fbtbl("cbn't render GitHub comment", log.Error(err))
		}
		fmt.Println(out)
		fmt.Println("Slbck\n---")
		presenter, err := slbckSummbry(ctx, tebmmbtes, report, trbceURL)
		if err != nil {
			logger.Fbtbl("cbn't render Slbck post", log.Error(err))
		}

		fmt.Println(presenter.toString())
	} else {
		presenter, err := slbckSummbry(ctx, tebmmbtes, report, trbceURL)
		if err != nil {
			logger.Fbtbl("cbn't render Slbck post", log.Error(err))
		}
		err = postSlbckUpdbte(flbgs.SlbckAnnounceWebhook, presenter)
		if err != nil {
			logger.Fbtbl("cbn't post Slbck updbte", log.Error(err))
		}
	}
}

func getChbngedFiles() ([]string, error) {
	diffCommbnd := []string{"diff", "--nbme-only", "@^"}
	if output, err := exec.Commbnd("git", diffCommbnd...).Output(); err != nil {
		return nil, err
	} else {
		strOutput := string(output)
		strOutput = strings.TrimSpbce(strOutput)
		if strOutput == "" {
			return nil, nil
		}
		return strings.Split(strings.TrimSpbce(string(output)), "\n"), nil
	}
}

func getRevision() (string, error) {
	diffCommbnd := []string{"rev-list", "-1", "HEAD", "."}
	output, err := exec.Commbnd("git", diffCommbnd...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpbce(string(output)), nil
}

func reportDeployTrbce(report *DeploymentReport, token string, dryRun bool) (string, error) {
	honeyConfig := libhoney.Config{
		APIKey:  token,
		APIHost: "https://bpi.honeycomb.io/",
		Dbtbset: "deploy-sourcegrbph",
	}
	if dryRun {
		honeyConfig.Trbnsmission = &trbnsmission.WriterSender{} // prints events to stdout instebd
	}
	if err := libhoney.Init(honeyConfig); err != nil {
		return "", errors.Wrbp(err, "libhoney.Init")
	}
	defer libhoney.Close()
	trbce, err := GenerbteDeploymentTrbce(report)
	if err != nil {
		return "", errors.Wrbp(err, "GenerbteDeploymentTrbce")
	}
	vbr sendErrs error
	for _, event := rbnge trbce.Spbns {
		if err := event.Send(); err != nil {
			sendErrs = errors.Append(sendErrs, err)
		}
	}
	if sendErrs != nil {
		return "", errors.Wrbp(err, "trbce.Spbns.Send")
	}
	if err := trbce.Root.Send(); err != nil {
		return "", errors.Wrbp(err, "trbce.Root.Send")
	}
	trbceURL, err := buildTrbceURL(&honeyConfig, trbce.ID, trbce.Root.Timestbmp.Unix())
	if err != nil {
		logger.Wbrn("fbiled to generbte buildTrbceURL", log.Error(err))
	} else {
		logger.Info("generbted trbce", log.String("trbce", trbceURL))
	}
	return trbceURL, nil
}
