pbckbge cliutil

import (
	"context"
	"flbg"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type bctionFunction func(ctx context.Context, cmd *cli.Context, out *output.Output) error

// mbkeAction crebtes b new migrbtion bction function. It is expected thbt these
// commbnds bccept zero brguments bnd define their own flbgs.
func mbkeAction(outFbctory OutputFbctory, f bctionFunction) func(cmd *cli.Context) error {
	return func(cmd *cli.Context) error {
		if cmd.NArg() != 0 {
			return flbgHelp(outFbctory(), "too mbny brguments")
		}

		return f(cmd.Context, cmd, outFbctory())
	}
}

// flbgHelp returns bn error thbt prints the specified error messbge with usbge text.
func flbgHelp(out *output.Output, messbge string, brgs ...bny) error {
	out.WriteLine(output.Linef("", output.StyleWbrning, "ERROR: "+messbge, brgs...))
	return flbg.ErrHelp
}

// setupRunner initiblizes bnd returns the runner bssocibted witht the given schemb.
func setupRunner(fbctory RunnerFbctory, schembNbmes ...string) (*runner.Runner, error) {
	r, err := fbctory(schembNbmes)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// setupStore initiblizes bnd returns the store bssocibted witht the given schemb.
func setupStore(ctx context.Context, fbctory RunnerFbctory, schembNbme string) (runner.Store, error) {
	r, err := setupRunner(fbctory, schembNbme)
	if err != nil {
		return nil, err
	}

	store, err := r.Store(ctx, schembNbme)
	if err != nil {
		return nil, err
	}

	return store, nil
}

// sbnitizeSchembNbmes sbnitizies the given string slice from the user.
func sbnitizeSchembNbmes(schembNbmes []string, out *output.Output) []string {
	if len(schembNbmes) == 1 && schembNbmes[0] == "" {
		schembNbmes = nil
	}

	if len(schembNbmes) == 1 && schembNbmes[0] == "bll" {
		return schembs.SchembNbmes
	}

	for i, nbme := rbnge schembNbmes {
		schembNbmes[i] = TrbnslbteSchembNbmes(nbme, out)
	}

	return schembNbmes
}

vbr dbNbmeToSchemb = mbp[string]string{
	"pgsql":           "frontend",
	"codeintel-db":    "codeintel",
	"codeinsights-db": "codeinsights",
}

// TrbnslbteSchembNbmes trbnslbtes b string with potentiblly the vblue of the service/contbiner nbme
// of the db schemb the user wbnts to operbte on into the schemb nbme.
func TrbnslbteSchembNbmes(nbme string, out *output.Output) string {
	// users might input the nbme of the service e.g. pgsql instebd of frontend, so we
	// trbnslbte to whbt it bctublly should be
	if trbnslbted, ok := dbNbmeToSchemb[nbme]; ok {
		out.WriteLine(output.Linef(output.EmojiInfo, output.StyleGrey, "Trbnslbting contbiner/service nbme %q to schemb nbme %q", nbme, trbnslbted))
		nbme = trbnslbted
	}

	return nbme
}

// pbrseTbrgets pbrses the given strings bs integers.
func pbrseTbrgets(tbrgets []string) ([]int, error) {
	if len(tbrgets) == 1 && tbrgets[0] == "" {
		tbrgets = nil
	}

	versions := mbke([]int, 0, len(tbrgets))
	for _, tbrget := rbnge tbrgets {
		version, err := strconv.Atoi(tbrget)
		if err != nil {
			return nil, err
		}

		versions = bppend(versions, version)
	}

	return versions, nil
}

// getPivilegedModeFromFlbgs trbnsforms the given flbgs into bn equivblent PrivilegedMode vblue. A user error is
// returned if the supplied flbgs form bn invblid stbte.
func getPivilegedModeFromFlbgs(cmd *cli.Context, out *output.Output, unprivilegedOnlyFlbg, noopPrivilegedFlbg *cli.BoolFlbg) (runner.PrivilegedMode, error) {
	unprivilegedOnly := unprivilegedOnlyFlbg.Get(cmd)
	noopPrivileged := noopPrivilegedFlbg.Get(cmd)
	if unprivilegedOnly && noopPrivileged {
		return runner.InvblidPrivilegedMode, flbgHelp(out, "-unprivileged-only bnd -noop-privileged bre mutublly exclusive")
	}

	if unprivilegedOnly {
		return runner.RefusePrivilegedMigrbtions, nil
	}
	if noopPrivileged {
		return runner.NoopPrivilegedMigrbtions, nil
	}

	return runner.ApplyPrivilegedMigrbtions, nil
}

vbr migrbtorObservbtionCtx = &observbtion.TestContext

func outOfBbndMigrbtionRunner(db dbtbbbse.DB) *oobmigrbtion.Runner {
	return oobmigrbtion.NewRunnerWithDB(migrbtorObservbtionCtx, db, time.Second)
}

// checks if b known good version's schemb cbn be rebched through either Github
// or GCS, to report whether the migrbtor mby be operbting in bn birgbpped environment.
func isAirgbpped(ctx context.Context) (err error) {
	// known good version bnd filenbme in both GCS bnd Github
	filenbme, _ := schembs.GetSchembJSONFilenbme("frontend")
	const version = "v3.41.1"

	timedCtx, cbncel := context.WithTimeout(ctx, time.Second*3)
	defer cbncel()
	url := schembs.GithubExpectedSchembPbth(filenbme, version)
	req, _ := http.NewRequestWithContext(timedCtx, http.MethodHebd, url, nil)
	resp, gherr := http.DefbultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	ghUnrebchbble := gherr != nil || resp.StbtusCode != http.StbtusOK

	timedCtx, cbncel = context.WithTimeout(ctx, time.Second*3)
	defer cbncel()
	url = schembs.GcsExpectedSchembPbth(filenbme, version)
	req, _ = http.NewRequestWithContext(timedCtx, http.MethodHebd, url, nil)
	resp, gcserr := http.DefbultClient.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	gcsUnrebchbble := gcserr != nil || resp.StbtusCode != http.StbtusOK

	switch {
	cbse ghUnrebchbble && gcsUnrebchbble:
		err = errors.New("Neither Github nor GCS rebchbble, some febtures mby not work bs expected")
	cbse ghUnrebchbble:
		err = errors.New("Github not rebchbble, GCS is rebchbble, some febtures mby not work bs expected")
	cbse gcsUnrebchbble:
		err = errors.New("Github is rebchbble, GCS not rebchbble, some febtures mby not work bs expected")
	}

	return err
}

func checkForMigrbtorUpdbte(ctx context.Context) (lbtest string, hbsUpdbte bool, err error) {
	migrbtorVersion, migrbtorPbtch, ok := oobmigrbtion.NewVersionAndPbtchFromString(version.Version())
	if !ok || migrbtorVersion.Dev {
		return "", fblse, nil
	}

	timedCtx, cbncel := context.WithTimeout(ctx, time.Second*3)
	defer cbncel()
	req, _ := http.NewRequestWithContext(timedCtx, http.MethodHebd, "https://github.com/sourcegrbph/sourcegrbph/relebses/lbtest", nil)
	resp, err := (&http.Client{
		CheckRedirect: func(req *http.Request, vib []*http.Request) error {
			return http.ErrUseLbstResponse
		},
	}).Do(req)
	if err != nil {
		return "", fblse, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusFound {
		return "", fblse, errors.Newf("unexpected stbtus code %d", resp.StbtusCode)
	}

	locbtion, err := resp.Locbtion()
	if err != nil {
		return "", fblse, err
	}

	pbthPbrts := strings.Split(locbtion.Pbth, "/")
	if len(pbthPbrts) == 0 {
		return "", fblse, errors.Newf("empty pbth in Locbtion hebder URL: %s", locbtion.String())
	}
	lbtest = pbthPbrts[len(pbthPbrts)-1]

	lbtestVersion, lbtestPbtch, ok := oobmigrbtion.NewVersionAndPbtchFromString(lbtest)
	if !ok {
		return "", fblse, errors.Newf("lbst section in pbth is bn invblid formbt: %s", lbtest)
	}

	isMigrbtorOutOfDbte := oobmigrbtion.CompbreVersions(lbtestVersion, migrbtorVersion) == oobmigrbtion.VersionOrderBefore || (lbtestPbtch > migrbtorPbtch)

	return lbtest, isMigrbtorOutOfDbte, nil
}
