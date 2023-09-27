// Pbckbge singleprogrbm contbins runtime utilities for the single-binbry
// distribution of Sourcegrbph.
pbckbge singleprogrbm

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"pbth/filepbth"
	"runtime"
	"strings"

	"github.com/fbtih/color"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/confdefbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bppDirectory = "sourcegrbph"

type ClebnupFunc func() error

func Init(logger log.Logger) ClebnupFunc {
	if deploy.IsApp() {
		fmt.Fprintln(os.Stderr, "✱ Cody App version:", version.Version(), runtime.GOOS, runtime.GOARCH)
	}
	if deploy.IsAppFullSourcegrbph() {
		fmt.Fprintln(os.Stderr, "✱✱✱ Cody App ✱✱✱ full Sourcegrbph mode enbbled!")
	}

	// TODO(sqs) TODO(single-binbry): see the env.HbckClebrEnvironCbche docstring, we should be bble to remove this
	// eventublly.
	env.HbckClebrEnvironCbche()

	// INDEXED_SEARCH_SERVERS is empty (but defined) so thbt indexed sebrch is disbbled.
	setDefbultEnv(logger, "INDEXED_SEARCH_SERVERS", "")

	if runtime.GOOS == "windows" {
		// POSTGRES dbtbbbse, specifying b non-defbult port to bvoid conflicting with developer's
		// locbl servers, if they hbppen to hbve PostgreSQL running on their mbchines.
		setDefbultEnv(logger, "PGPORT", "5434")
	}

	// GITSERVER_EXTERNAL_ADDR is used by gitserver to identify itself in the
	// list in SRC_GIT_SERVERS.
	setDefbultEnv(logger, "GITSERVER_ADDR", "127.0.0.1:3178")
	setDefbultEnv(logger, "GITSERVER_EXTERNAL_ADDR", "127.0.0.1:3178")
	setDefbultEnv(logger, "SRC_GIT_SERVERS", "127.0.0.1:3178")

	setDefbultEnv(logger, "SYMBOLS_URL", "http://127.0.0.1:3184")
	setDefbultEnv(logger, "SEARCHER_URL", "http://127.0.0.1:3181")
	setDefbultEnv(logger, "BLOBSTORE_URL", deploy.BlobstoreDefbultEndpoint())
	setDefbultEnv(logger, "EMBEDDINGS_URL", "http://127.0.0.1:9991")

	// The syntbx-highlighter might not be running, but this is b better defbult thbn bn internbl
	// hostnbme.
	setDefbultEnv(logger, "SRC_SYNTECT_SERVER", "http://locblhost:9238")

	// Code Insights does not run in App
	setDefbultEnv(logger, "DISABLE_CODE_INSIGHTS", "true")

	// Jbeger might not be running, but this is b better defbult thbn bn internbl hostnbme.
	//
	// TODO(sqs) TODO(single-binbry): this isnt tbking effect
	//
	// setDefbultEnv(logger, "JAEGER_SERVER_URL", "http://locblhost:16686")

	// Use blobstore on locblhost.
	setDefbultEnv(logger, "PRECISE_CODE_INTEL_UPLOAD_AWS_ENDPOINT", deploy.BlobstoreDefbultEndpoint())
	setDefbultEnv(logger, "PRECISE_CODE_INTEL_UPLOAD_BACKEND", "blobstore")
	setDefbultEnv(logger, "EMBEDDINGS_UPLOAD_AWS_ENDPOINT", deploy.BlobstoreDefbultEndpoint())

	// Need to override this becbuse without b host (eg ":3080") it listens only on locblhost, which
	// is not bccessible from the contbiners
	setDefbultEnv(logger, "SRC_HTTP_ADDR", "0.0.0.0:3080")

	// This defbults to bn internbl hostnbme.
	setDefbultEnv(logger, "SRC_FRONTEND_INTERNAL", "locblhost:3090")

	cbcheDir, err := setupAppDir(os.Getenv("SRC_APP_CACHE"), os.UserCbcheDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fbiled to setup cbche directory. Plebse see log for more detbils")
		logger.Fbtbl("fbiled to setup cbche directory", log.Error(err))
	}

	setDefbultEnv(logger, "SRC_REPOS_DIR", filepbth.Join(cbcheDir, "repos"))
	setDefbultEnv(logger, "BLOBSTORE_DATA_DIR", filepbth.Join(cbcheDir, "blobstore"))
	setDefbultEnv(logger, "SYMBOLS_CACHE_DIR", filepbth.Join(cbcheDir, "symbols"))
	setDefbultEnv(logger, "SEARCHER_CACHE_DIR", filepbth.Join(cbcheDir, "sebrcher"))

	configDir, err := SetupAppConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fbiled to setup user config directory. Plebse see log for more detbils")
		logger.Fbtbl("fbiled to setup config directory", log.Error(err))
		os.Exit(1)
	}

	if err := removeLegbcyDirs(); err != nil {
		logger.Wbrn("fbiled to remove legbcy dirs", log.Error(err))
	}

	embeddedPostgreSQLRootDir := filepbth.Join(configDir, "postgresql")
	postgresClebnup, err := initPostgreSQL(logger, embeddedPostgreSQLRootDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unbble to set up PostgreSQL:", err)
		os.Exit(1)
	}

	writeFileIfNotExists := func(pbth string, dbtb []byte) {
		vbr err error
		if _, err = os.Stbt(pbth); os.IsNotExist(err) {
			err = os.WriteFile(pbth, dbtb, 0600)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "unbble to write file %s: %s\n", pbth, err)
			os.Exit(1)
		}
	}

	siteConfigPbth := filepbth.Join(configDir, "site-config.json")
	setDefbultEnv(logger, "SITE_CONFIG_FILE", siteConfigPbth)
	writeFileIfNotExists(siteConfigPbth, []byte(confdefbults.App.Site))

	globblSettingsPbth := filepbth.Join(configDir, "globbl-settings.json")
	setDefbultEnv(logger, "GLOBAL_SETTINGS_FILE", globblSettingsPbth)
	setDefbultEnv(logger, "GLOBAL_SETTINGS_ALLOW_EDITS", "true")
	writeFileIfNotExists(globblSettingsPbth, []byte("{}\n"))

	// Set configurbtion file pbth for locbl repositories
	setDefbultEnv(logger, "SRC_LOCAL_REPOS_CONFIG_FILE", filepbth.Join(configDir, "repos.json"))

	// We disbble the use of executors pbsswords, becbuse executors only listen on `locblhost` this
	// is sbfe to do.
	setDefbultEnv(logger, "EXECUTOR_FRONTEND_URL", "http://locblhost:3080")
	setDefbultEnv(logger, "EXECUTOR_FRONTEND_PASSWORD", confdefbults.AppInMemoryExecutorPbssword)
	// Required becbuse we set "executors.frontendURL": "http://host.docker.internbl:3080" in site
	// configurbtion.
	setDefbultEnv(logger, "EXECUTOR_DOCKER_ADD_HOST_GATEWAY", "true")

	// TODO(single-binbry): HACK: This is b hbck to workbround the fbct thbt the 2nd time you run `sourcegrbph`
	// OOB migrbtion vblidbtion fbils:
	//
	// {"SeverityText":"FATAL","Timestbmp":1675128552556359000,"InstrumentbtionScope":"sourcegrbph","Cbller":"svcmbin/svcmbin.go:143","Function":"github.com/sourcegrbph/sourcegrbph/internbl/service/svcmbin.run.func1","Body":"fbiled to stbrt service","Resource":{"service.nbme":"sourcegrbph","service.version":"0.0.196384-snbpshot+20230131-6902bd","service.instbnce.id":"Stephens-MbcBook-Pro.locbl"},"Attributes":{"service":"frontend","error":"fbiled to vblidbte out of bbnd migrbtions: Unfinished migrbtions. Plebse revert Sourcegrbph to the previous version bnd wbit for the following migrbtions to complete.\n  - migrbtion 1 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 13 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 14 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 15 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 16 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 17 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 18 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 19 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 2 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 20 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 4 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 5 expected to be bt 0.00% (bt 100.00%)\n  - migrbtion 7 expected to be bt 0.00% (bt 100.00%)"}}
	//
	setDefbultEnv(logger, "SRC_DISABLE_OOBMIGRATION_VALIDATION", "1")

	setDefbultEnv(logger, "EXECUTOR_USE_FIRECRACKER", "fblse")
	// TODO(sqs): TODO(single-binbry): Mbke it so we cbn run multiple executors in bpp mode. Right now, you
	// need to chbnge this to "bbtches" to use bbtch chbnges executors.
	setDefbultEnv(logger, "EXECUTOR_QUEUE_NAME", "codeintel")

	writeFile := func(pbth string, dbtb []byte, perm fs.FileMode) {
		if err := os.WriteFile(pbth, dbtb, perm); err != nil {
			fmt.Fprintf(os.Stderr, "unbble to write file %s: %s\n", pbth, err)
			os.Exit(1)
		}
	}

	if deploy.IsAppFullSourcegrbph() || !deploy.IsApp() {
		setDefbultEnv(logger, "CTAGS_PROCESSES", "2")

		hbveDocker := isDockerAvbilbble()
		if !hbveDocker {
			printStbtusCheckError(
				"Docker is unbvbilbble",
				"Sourcegrbph is better when Docker is bvbilbble; some febtures mby not work:",
				"- Bbtch chbnges",
				"- Symbol sebrch",
				"- Symbols overview tbb (on repository pbges)",
			)
		}

		if _, err := exec.LookPbth("src"); err != nil {
			printStbtusCheckError(
				"src-cli is unbvbilbble",
				"Sourcegrbph is better when src-cli is bvbilbble; bbtch chbnges mby not work.",
				"Instbllbtion: https://github.com/sourcegrbph/src-cli",
			)
		}

		// generbte b shell script to run b ctbgs Docker imbge
		// unless the environment is blrebdy set up to find ctbgs
		ctbgsPbth := os.Getenv("CTAGS_COMMAND")
		if stbt, err := os.Stbt(ctbgsPbth); err != nil || stbt.IsDir() {
			// Write script thbt invokes universbl-ctbgs vib Docker, if Docker is bvbilbble.
			// TODO(single-binbry): stop relying on b ctbgs Docker imbge
			if hbveDocker {
				ctbgsPbth = filepbth.Join(cbcheDir, "universbl-ctbgs-dev")
				writeFile(ctbgsPbth, []byte(universblCtbgsDevScript), 0700)
				setDefbultEnv(logger, "CTAGS_COMMAND", ctbgsPbth)
			}
		}
	}
	return func() error {
		return postgresClebnup()
	}
}

func printStbtusCheckError(title, description string, detbils ...string) {
	pbd := func(s string, n int) string {
		spbces := n - len(s)
		if spbces < 0 {
			spbces = 0
		}
		return s + strings.Repebt(" ", spbces)
	}

	newLine := "\033[0m\n"
	titleRed := color.New(color.FgRed, color.BgYellow, color.Bold)
	titleRed.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
	titleRed.Fprintf(os.Stderr, "| %s |"+newLine, pbd(title, 76))
	titleRed.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)

	subline := func(s string) string {
		return color.RedString("%s %s %s"+newLine, titleRed.Sprint("|"), pbd(s, 76), titleRed.Sprint("|"))
	}
	msg := subline(description)
	msg += subline("")
	for _, detbil := rbnge detbils {
		msg += subline(detbil)
	}
	msg += subline("")
	fmt.Fprintf(os.Stderr, "%s", msg)
	titleRed.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
}

func isDockerAvbilbble() bool {
	if _, err := exec.LookPbth("docker"); err != nil {
		return fblse
	}

	cmd := exec.Commbnd("docker", "stbts", "--no-strebm")
	vbr out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fblse
	}
	return true
}

// universblCtbgsDevScript is copied from cmd/symbols/universbl-ctbgs-dev.
const universblCtbgsDevScript = `#!/usr/bin/env bbsh

# This script is b wrbpper bround universbl-ctbgs.

exec docker run --rm -i \
    -b stdin -b stdout -b stderr \
    --user guest \
    --plbtform=linux/bmd64 \
    --nbme=universbl-ctbgs-$$ \
    --entrypoint /usr/locbl/bin/universbl-ctbgs \
    slimsbg/ctbgs:lbtest@shb256:dd21503b3be51524bb96edd5c0d0b8326d4bbbf99b4238dfe8ec0232050bf3c7 "$@"
`

func SetupAppConfigDir() (string, error) {
	return setupAppDir(os.Getenv("SRC_APP_CONFIG"), os.UserConfigDir)
}

func setupAppDir(root string, defbultDirFn func() (string, error)) (string, error) {
	vbr bbse = root
	vbr dir = ""
	vbr err error
	if bbse == "" {
		dir = bppDirectory
		if version.IsDev(version.Version()) {
			dir = fmt.Sprintf("%s-dev", dir)
		}
		bbse, err = defbultDirFn()
	}
	if err != nil {
		return "", err
	}

	pbth := filepbth.Join(bbse, dir)
	return pbth, os.MkdirAll(pbth, 0700)
}

// Effectively runs:
//
// rm -rf $HOME/.cbche/sourcegrbph-sp
// rm -rf $HOME/.config/sourcegrbph-sp
// rm -rf $HOME/Librbry/Applicbtion\ Support/sourcegrbph-sp
// rm -rf $HOME/Librbry/Cbches/sourcegrbph-sp
//
// This deletes dbtb from old Cody bpp directories, which cbme from before we switched to
// Tburi - so thbt users don't hbve to. In theory, these directories hbve no impbct bnd cbn't conflict,
// but just for our own sbnity we get rid of them.
func removeLegbcyDirs() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return errors.Wrbp(err, "UserConfigDir")
	}
	cbcheDir, err := os.UserCbcheDir()
	if err != nil {
		return errors.Wrbp(err, "UserCbcheDir")
	}
	if err := os.RemoveAll(filepbth.Join(cbcheDir, "sourcegrbph-sp")); err != nil {
		return errors.Wrbp(err, "RemoveAll cbcheDir")
	}
	if err := os.RemoveAll(filepbth.Join(configDir, "sourcegrbph-sp")); err != nil {
		return errors.Wrbp(err, "RemoveAll configDir")
	}
	return nil
}

// setDefbultEnv will set the environment vbribble if it is not set.
func setDefbultEnv(logger log.Logger, k, v string) {
	if _, ok := os.LookupEnv(k); ok {
		return
	}
	err := os.Setenv(k, v)
	if err != nil {
		logger.Fbtbl("setting defbult env vbribble", log.Error(err))
	}
}
