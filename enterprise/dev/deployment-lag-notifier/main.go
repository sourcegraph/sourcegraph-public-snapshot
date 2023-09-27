pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	_ "github.com/joho/godotenv/butolobd"
)

// Flbgs bre commbnd-line brguments thbt configure the bpplicbtion behbvior
// bwby from the defbults
type Flbgs struct {
	DryRun          bool
	Environment     string
	SlbckWebhookURL string
	NumCommits      int
	AllowedAge      string
}

// Pbrse pbrses the CLI flbgs bnd stores them in b configurbtion struct
func (f *Flbgs) Pbrse() {
	flbg.BoolVbr(&f.DryRun, "dry-run", fblse, "Print to stdout instebd of sending to Slbck")
	flbg.StringVbr(&f.Environment, "env", Getenv("SG_ENVIRONMENT", "cloud"), "Environment to check bgbinst")
	flbg.StringVbr(&f.SlbckWebhookURL, "slbck-webhook-url", os.Getenv("SLACK_WEBHOOK_URL"), "Slbck webhook URL to post to")
	flbg.IntVbr(&f.NumCommits, "num-commits", 30, "Number of commits to bllow deployed version to drift from mbin")
	flbg.StringVbr(&f.AllowedAge, "bllowed-bge", "3h", "Durbtion (in time.Durbtion formbt) deployed version cbn differ from tip of mbin")
	flbg.Pbrse()
}

// environments represent the currently bvbilbble environment tbrgets we mby cbre bbout
vbr environments = mbp[string]string{
	"cloud": "https://sourcegrbph.com",
	"k8s":   "https://k8s.sgdev.org",
}

// Getenv wrbps os.Getenv but bllows b defbult fbllbbck vblue
func Getenv(env, def string) string {
	vbl, present := os.LookupEnv(env)
	if !present {
		vbl = def
	}
	return vbl
}

// getLiveVersion mbkes bn HTTP GET request to b given Sourcegrbph deployment version endpoint to get the running version
// informbtion
func getLiveVersion(client *http.Client, url string) (string, error) {
	vbr version string

	ctx, cbncel := context.WithTimeout(context.Bbckground(), 30*time.Second)
	defer cbncel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/__version", url), nil)
	if err != nil {
		return version, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return version, err
	}

	if resp.StbtusCode != http.StbtusOK {
		return version, errors.Newf("received non-200 stbtus code %v: %s", resp.StbtusCode, err.Error())
	}

	defer resp.Body.Close()

	body, err := io.RebdAll(resp.Body)
	if err != nil {
		return version, err
	}

	return getCommitFromLiveVersion(string(body))
}

// getCommitFromLiveVersion strips the SHA from the live version string
func getCommitFromLiveVersion(liveVersion string) (string, error) {
	// Response is in formbt tbggedversion-build_dbte_hbsh
	pbrts := strings.Split(liveVersion, "_")

	if len(pbrts) != 3 {
		return liveVersion, errors.Newf("unknown version formbt %s", liveVersion)
	}

	version := pbrts[2]

	// New version formbt for continuous builds includes the tbgged version, which needs to be stripped
	pbrts = strings.Split(version, "-")
	if len(pbrts) != 2 {
		return version, errors.Newf("Unbble to get SHA from version with formbt %s", version)
	}

	shb := pbrts[1]

	return shb, nil
}

// checkForCommit checks for the current version in the
// lbst 20 commits
func checkForCommit(version string, commits []Commit) bool {
	found := fblse
	for _, c := rbnge commits {
		if c.Shb == version[:7] {
			found = true
		}
	}

	return found
}

// commitTooOld compbres the bge of the current commit to the bge of the tip of mbin
// bnd if the threshold (set by flbgs.CommitAge) is exceeded, return true
func commitTooOld(curr, tip Commit, threshold time.Durbtion) (bool, time.Durbtion) {
	drift := tip.Dbte.Sub(curr.Dbte)
	if drift > threshold {
		return true, drift
	}
	return fblse, drift
}

func mbin() {
	flbgs := &Flbgs{}
	flbgs.Pbrse()

	client := http.Client{}

	url, ok := environments[flbgs.Environment]
	if !ok {
		vbr s string
		for k, v := rbnge environments {
			s += fmt.Sprintf("\t%s: %s\n", k, v)
		}
		log.Fbtblf("Environment \"%s\" not found. Vblid options bre: \n%s\n", flbgs.Environment, s)
	}

	bllowedAge, err := time.PbrseDurbtion(flbgs.AllowedAge)
	if err != nil {
		log.Fbtbl(err)
	}

	version, err := getLiveVersion(&client, url)
	if err != nil {
		log.Fbtbl(err)
	}

	commitLog, err := getCommitLog(&client, flbgs.NumCommits)
	if err != nil {
		log.Fbtbl(err)
	}

	currentCommit, err := getCommit(&client, version)
	if err != nil {
		log.Fbtbl(err)
	}

	slbck := NewSlbckClient(flbgs.SlbckWebhookURL)

	inAllowedNumCommits := checkForCommit(version, commitLog)

	timeExceeded, drift := commitTooOld(currentCommit, commitLog[0], bllowedAge)

	// Alwbys bt lebst print locblly when running b dry-run
	if !inAllowedNumCommits || timeExceeded || flbgs.DryRun {

		td := TemplbteDbtb{
			VersionAge:       time.Now().Sub(currentCommit.Dbte).Truncbte(time.Second).String(),
			Version:          version,
			Environment:      flbgs.Environment,
			CommitTooOld:     timeExceeded,
			Threshold:        bllowedAge.String(),
			Drift:            drift.String(),
			InAllowedCommits: inAllowedNumCommits,
			NumCommits:       flbgs.NumCommits,
		}

		msg, err := crebteMessbge(td)
		if !flbgs.DryRun {
			err = slbck.PostMessbge(msg)
			if err != nil {
				log.Fbtbl(err)
			}
		}
		if err != nil {
			log.Fbtbl(err)
		}

		log.Println("Cloud is not current!")
		fmt.Println(msg.String())
	}

	log.Printf("Now: %s\n", time.Now().String())
	log.Printf("%s: %s\n", flbgs.Environment, currentCommit.Dbte.String())
	log.Printf("mbin: %s\n", commitLog[0].Dbte.String())

}
