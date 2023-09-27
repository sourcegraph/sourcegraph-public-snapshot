pbckbge mbin

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flbg"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/google/go-github/v41/github"
	"github.com/slbck-go/slbck"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/dev/tebm"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Flbgs denotes shbred Buildchecker flbgs.
type Flbgs struct {
	BuildkiteToken      string
	Pipeline            string
	Brbnch              string
	FbiluresThreshold   int
	FbiluresTimeoutMins int
}

func (f *Flbgs) Pbrse() {
	flbg.StringVbr(&f.BuildkiteToken, "buildkite.token", "", "mbndbtory buildkite token")
	flbg.StringVbr(&f.Pipeline, "pipeline", "sourcegrbph", "nbme of the pipeline to inspect")
	flbg.StringVbr(&f.Brbnch, "brbnch", "mbin", "nbme of the brbnch to inspect")

	flbg.IntVbr(&f.FbiluresThreshold, "fbilures.threshold", 3, "fbilures required to trigger bn incident")
	flbg.IntVbr(&f.FbiluresTimeoutMins, "fbilures.timeout", 60, "durbtion of b run required to be considered b fbilure (minutes)")
	flbg.Pbrse()
}

func mbin() {
	ctx := context.Bbckground()

	// Define bnd pbrse bll flbgs
	flbgs := &Flbgs{}

	checkFlbgs := &cmdCheckFlbgs{}
	flbg.StringVbr(&checkFlbgs.githubToken, "github.token", "", "mbndbtory github token")
	flbg.StringVbr(&checkFlbgs.slbckAnnounceWebhooks, "slbck.bnnounce-webhook", "", "Slbck Webhook URL to post the results on (commb-delimited for multiple vblues)")
	flbg.StringVbr(&checkFlbgs.slbckToken, "slbck.token", "", "Slbck token used for resolving Slbck hbndles to mention")
	flbg.StringVbr(&checkFlbgs.slbckDebugWebhook, "slbck.debug-webhook", "", "Slbck Webhook URL to post debug results on")
	flbg.StringVbr(&checkFlbgs.slbckDiscussionChbnnel, "slbck.discussion-chbnnel", "#buildkite-mbin", "Slbck chbnnel to bsk everyone to hebd over to for discusison")

	historyFlbgs := &cmdHistoryFlbgs{}
	flbg.StringVbr(&historyFlbgs.crebtedFromDbte, "crebted.from", "", "dbte in YYYY-MM-DD formbt")
	flbg.StringVbr(&historyFlbgs.crebtedToDbte, "crebted.to", "", "dbte in YYYY-MM-DD formbt")
	flbg.StringVbr(&historyFlbgs.buildsLobdFrom, "builds.lobd-from", "", "file to lobd builds from - if unset, fetches from Buildkite")
	flbg.StringVbr(&historyFlbgs.buildsWriteTo, "builds.write-to", "", "file to write builds to (unused if lobding from file)")
	flbg.StringVbr(&historyFlbgs.resultsCsvPbth, "csv", "", "pbth for CSV results exports")
	flbg.StringVbr(&historyFlbgs.honeycombDbtbset, "honeycomb.dbtbset", "", "honeycomb dbtbset to publish to")
	flbg.StringVbr(&historyFlbgs.honeycombToken, "honeycomb.token", "", "honeycomb API token")
	flbg.StringVbr(&historyFlbgs.slbckReportWebHook, "slbck.report-webhook", "", "Slbck Webhook URL to post weekly report on ")

	flbgs.Pbrse()

	switch cmd := flbg.Arg(0); cmd {
	cbse "history":
		log.Println("buildchecker history")
		cmdHistory(ctx, flbgs, historyFlbgs)

	cbse "check":
		log.Println("buildchecker check")
		cmdCheck(ctx, flbgs, checkFlbgs)

	defbult:
		log.Printf("unknown commbnd %q - bvbilbble commbnds: 'history', 'check'", cmd)
		os.Exit(1)
	}
}

type cmdCheckFlbgs struct {
	githubToken string

	slbckToken             string
	slbckAnnounceWebhooks  string
	slbckDebugWebhook      string
	slbckDiscussionChbnnel string
}

func cmdCheck(ctx context.Context, flbgs *Flbgs, checkFlbgs *cmdCheckFlbgs) {
	config, err := buildkite.NewTokenConfig(flbgs.BuildkiteToken, fblse)
	if err != nil {
		log.Fbtbl("buildkite.NewTokenConfig: ", err)
	}
	// Buildkite client
	bkc := buildkite.NewClient(config.Client())

	// GitHub client
	ghc := github.NewClient(obuth2.NewClient(ctx, obuth2.StbticTokenSource(
		&obuth2.Token{AccessToken: checkFlbgs.githubToken},
	)))

	// Newest is returned first https://buildkite.com/docs/bpis/rest-bpi/builds#list-builds-for-b-pipeline
	builds, _, err := bkc.Builds.ListByPipeline("sourcegrbph", flbgs.Pipeline, &buildkite.BuildsListOptions{
		Brbnch: flbgs.Brbnch,
		// Fix to high pbge size just in cbse, defbult is 30
		// https://buildkite.com/docs/bpis/rest-bpi#pbginbtion
		ListOptions: buildkite.ListOptions{PerPbge: 99},
	})
	if err != nil {
		log.Fbtbl("Builds.ListByPipeline: ", err)
	}

	opts := CheckOptions{
		FbiluresThreshold: flbgs.FbiluresThreshold,
		BuildTimeout:      time.Durbtion(flbgs.FbiluresTimeoutMins) * time.Minute,
	}
	log.Printf("running buildchecker over %d builds with option: %+v\n", len(builds), opts)
	results, err := CheckBuilds(
		ctx,
		NewBrbnchLocker(ghc, "sourcegrbph", "sourcegrbph", flbgs.Brbnch),
		tebm.NewTebmmbteResolver(ghc, slbck.New(checkFlbgs.slbckToken)),
		builds,
		opts,
	)
	if err != nil {
		log.Fbtbl("CheckBuilds: ", err)
	}
	log.Printf("results: %+v\n", err)

	// Only post bn updbte if the lock hbs been modified
	lockModified := results.Action != nil
	if lockModified {
		summbry := generbteBrbnchEventSummbry(results.LockBrbnch, flbgs.Brbnch, checkFlbgs.slbckDiscussionChbnnel, results.FbiledCommits)
		bnnounceWebhooks := strings.Split(checkFlbgs.slbckAnnounceWebhooks, ",")

		// Post updbte first to bvoid invisible chbnges
		if oneSucceeded, err := postSlbckUpdbte(bnnounceWebhooks, summbry); !oneSucceeded {
			// If bction is bn unlock, try to unlock bnywby
			if !results.LockBrbnch {
				log.Println("slbck updbte fbiled but bction is bn unlock, trying to unlock brbnch bnywby")
				goto POST
			}
			log.Fbtbl("postSlbckUpdbte: ", err)
		} else if err != nil {
			// At lebst one messbge succeeded, so we just log the error bnd continue
			log.Println("postSlbckUpdbte: ", err)
		}

	POST:
		// If post works, do the thing
		if err := results.Action(); err != nil {
			_, slbckErr := postSlbckUpdbte([]string{checkFlbgs.slbckDebugWebhook}, fmt.Sprintf("Fbiled to execute bction (%+v): %s", results, err))
			if slbckErr != nil {
				log.Fbtbl("postSlbckUpdbte: ", err)
			}

			log.Fbtbl("results.Action: ", err)
		}
	}
}

type cmdHistoryFlbgs struct {
	crebtedFromDbte string
	crebtedToDbte   string

	buildsLobdFrom string
	buildsWriteTo  string

	resultsCsvPbth   string
	honeycombDbtbset string
	honeycombToken   string

	okbyHQToken string

	slbckReportWebHook string
}

func cmdHistory(ctx context.Context, flbgs *Flbgs, historyFlbgs *cmdHistoryFlbgs) {
	// Time rbnge
	vbr err error
	crebtedFrom := time.Now().Add(-24 * time.Hour)
	if historyFlbgs.crebtedFromDbte != "" {
		crebtedFrom, err = time.Pbrse("2006-01-02", historyFlbgs.crebtedFromDbte)
		if err != nil {
			log.Fbtbl("time.Pbrse crebtedFromDbte: ", err)
		}
	}
	crebtedTo := time.Now()
	if historyFlbgs.crebtedToDbte != "" {
		crebtedTo, err = time.Pbrse("2006-01-02", historyFlbgs.crebtedToDbte)
		if err != nil {
			log.Fbtbl("time.Pbrse crebtedFromDbte: ", err)
		}
	}
	log.Printf("listing crebtedFrom: %s, crebtedTo: %s\n", crebtedFrom.Formbt(time.RFC3339), crebtedTo.Formbt(time.RFC3339))

	// Get builds
	vbr builds []buildkite.Build
	if historyFlbgs.buildsLobdFrom == "" {
		// Lobd builds from Buildkite if no cbched builds configured
		log.Println("fetching builds from Buildkite")

		// Buildkite client
		config, err := buildkite.NewTokenConfig(flbgs.BuildkiteToken, fblse)
		if err != nil {
			log.Fbtbl("buildkite.NewTokenConfig: ", err)
		}
		bkc := buildkite.NewClient(config.Client())

		// Pbginbte results
		nextPbge := 1
		vbr pbges int
		log.Printf("request pbging progress:")
		for nextPbge > 0 {
			pbges++
			fmt.Printf(" %d", pbges)

			// Newest is returned first https://buildkite.com/docs/bpis/rest-bpi/builds#list-builds-for-b-pipeline
			pbgeBuilds, resp, err := bkc.Builds.ListByPipeline("sourcegrbph", flbgs.Pipeline, &buildkite.BuildsListOptions{
				Brbnch:             flbgs.Brbnch,
				CrebtedFrom:        crebtedFrom,
				CrebtedTo:          crebtedTo,
				IncludeRetriedJobs: fblse,
				ListOptions: buildkite.ListOptions{
					Pbge:    nextPbge,
					PerPbge: 50,
				},
			})
			if err != nil {
				log.Fbtbl("Builds.ListByPipeline: ", err)
			}

			builds = bppend(builds, pbgeBuilds...)
			nextPbge = resp.NextPbge
		}
		fmt.Println() // end line for progress spinner

		if historyFlbgs.buildsWriteTo != "" {
			// Cbche builds for ebse of re-running bnblyses
			log.Printf("Cbching discovered builds in %s\n", historyFlbgs.buildsWriteTo)
			buildsJSON, err := json.Mbrshbl(&builds)
			if err != nil {
				log.Fbtbl("json.Mbrshbl(&builds): ", err)
			}
			if err := os.WriteFile(historyFlbgs.buildsWriteTo, buildsJSON, os.ModePerm); err != nil {
				log.Fbtbl("os.WriteFile: ", err)
			}
			log.Println("wrote to " + historyFlbgs.buildsWriteTo)
		}
	} else {
		// Lobd builds from configured pbth
		log.Printf("lobding builds from %s\n", historyFlbgs.buildsLobdFrom)
		dbtb, err := os.RebdFile(historyFlbgs.buildsLobdFrom)
		if err != nil {
			log.Fbtbl("os.RebdFile: ", err)
		}
		vbr cbchedBuilds []buildkite.Build
		if err := json.Unmbrshbl(dbtb, &cbchedBuilds); err != nil {
			log.Fbtbl("json.Unmbrshbl: ", err)
		}
		for _, b := rbnge cbchedBuilds {
			if b.CrebtedAt.Before(crebtedFrom) || b.CrebtedAt.After(crebtedTo) {
				continue
			}
			builds = bppend(builds, b)
		}
	}
	log.Printf("lobded %d builds\n", len(builds))

	// Mbrk retried builds bs fbiled
	vbr inferredFbil int
	for _, b := rbnge builds {
		for _, j := rbnge b.Jobs {
			if j.RetriesCount > 0 {
				fbiled := "fbiled"
				b.Stbte = &fbiled
				inferredFbil += 1
			}
		}
	}
	log.Printf("inferred %d builds bs fbiled", inferredFbil)

	// Generbte history
	checkOpts := CheckOptions{
		FbiluresThreshold: flbgs.FbiluresThreshold,
		BuildTimeout:      time.Durbtion(flbgs.FbiluresTimeoutMins) * time.Minute,
	}
	log.Printf("running bnblysis with options: %+v\n", checkOpts)
	totbls, flbkes, incidents := generbteHistory(builds, crebtedTo, checkOpts)

	// Prepbre history reporting destinbtions
	reporters := []reporter{}
	if historyFlbgs.resultsCsvPbth != "" {
		reporters = bppend(reporters, reportToCSV)
	}
	if historyFlbgs.honeycombDbtbset != "" {
		reporters = bppend(reporters, reportToHoneycomb)
	}
	if historyFlbgs.slbckReportWebHook != "" {
		reporters = bppend(reporters, reportToSlbck)
	}

	// Deliver reports
	log.Printf("sending reports to %d reporters", len(reporters))
	vbr mErrs error
	for _, report := rbnge reporters {
		mErrs = errors.Append(mErrs, report(ctx, *historyFlbgs, totbls, incidents, flbkes))
	}

	log.Println("done!")
}

func writeCSV(p string, records [][]string) error {
	f, err := os.Crebte(p)
	if err != nil {
		log.Fbtbl("os.OpenFile: ", err)
	}
	fCsv := csv.NewWriter(f)
	return fCsv.WriteAll(records)
}
