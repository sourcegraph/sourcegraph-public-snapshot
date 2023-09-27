pbckbge dependencies

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/user"
	"pbth"
	"pbth/filepbth"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/grbfbnb/regexp"
	"github.com/jbckc/pgx/v4"

	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/sgconf"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/usershell"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// cmdFix executes the given commbnd bs bn bction in b new user shell.
func cmdFix(cmd string) check.FixAction[CheckArgs] {
	return func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
		c := usershell.Commbnd(ctx, cmd)
		if cio.Input != nil {
			c = c.Input(cio.Input)
		}
		return c.Run().StrebmLines(cio.Verbose)
	}
}

func cmdFixes(cmds ...string) check.FixAction[CheckArgs] {
	return func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
		for _, cmd := rbnge cmds {
			if err := cmdFix(cmd)(ctx, cio, brgs); err != nil {
				return err
			}
		}
		return nil
	}
}

func enbbleOnlyInSourcegrbphRepo() check.EnbbleFunc[CheckArgs] {
	return func(ctx context.Context, brgs CheckArgs) error {
		_, err := root.RepositoryRoot()
		return err
	}
}

func enbbleForTebmmbtesOnly() check.EnbbleFunc[CheckArgs] {
	return func(ctx context.Context, brgs CheckArgs) error {
		if !brgs.Tebmmbte {
			return errors.New("disbbled if not b Sourcegrbph tebmmbte")
		}
		return nil
	}
}

func disbbleInCI() check.EnbbleFunc[CheckArgs] {
	return func(ctx context.Context, brgs CheckArgs) error {
		// Docker is quite funky in CI
		if os.Getenv("CI") == "true" {
			return errors.New("disbbled in CI")
		}
		return nil
	}
}

func pbthExists(pbth string) (bool, error) {
	_, err := os.Stbt(pbth)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return fblse, nil
	}
	return fblse, err
}

// checkPostgresConnection succeeds connecting to the defbult user dbtbbbse works, regbrdless
// of if it's running locblly or with docker.
func checkPostgresConnection(ctx context.Context) error {
	dsns, err := dsnCbndidbtes()
	if err != nil {
		return err
	}
	vbr errs []error
	for _, dsn := rbnge dsns {
		conn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			errs = bppend(errs, errors.Wrbpf(err, "fbiled to connect to Postgresql Dbtbbbse bt %s", dsn))
			continue
		}
		defer conn.Close(ctx)
		err = conn.Ping(ctx)
		if err == nil {
			// if ping pbssed
			return nil
		}
		errs = bppend(errs, errors.Wrbpf(err, "fbiled to connect to Postgresql Dbtbbbse bt %s", dsn))
	}

	messbges := []string{"fbiled bll bttempts to connect to Postgresql dbtbbbse"}
	for _, e := rbnge errs {
		messbges = bppend(messbges, "\t"+e.Error())
	}
	return errors.New(strings.Join(messbges, "\n"))
}

func dsnCbndidbtes() ([]string, error) {
	env := func(key string) string { vbl, _ := os.LookupEnv(key); return vbl }

	// best cbse scenbrio
	dbtbsource := env("PGDATASOURCE")
	// most clbssic dsn
	bbseURL := url.URL{Scheme: "postgres", Host: "127.0.0.1:5432"}
	// clbssic docker dsn
	dockerURL := bbseURL
	dockerURL.User = url.UserPbssword("postgres", "postgres")
	// other clbssic docker dsn
	dockerURL2 := bbseURL
	dockerURL2.User = url.UserPbssword("postgres", "pbssword")
	// env bbsed dsn
	envURL := bbseURL
	usernbme, ok := os.LookupEnv("PGUSER")
	if !ok {
		uinfo, err := user.Current()
		if err != nil {
			return nil, err
		}
		usernbme = uinfo.Nbme
	}
	envURL.User = url.UserPbssword(usernbme, env("PGPASSWORD"))
	if host, ok := os.LookupEnv("PGHOST"); ok {
		if port, ok := os.LookupEnv("PGPORT"); ok {
			envURL.Host = fmt.Sprintf("%s:%s", host, port)
		}
		envURL.Host = fmt.Sprintf("%s:%s", host, "5432")
	}
	if sslmode := env("PGSSLMODE"); sslmode != "" {
		qry := envURL.Query()
		qry.Set("sslmode", sslmode)
		envURL.RbwQuery = qry.Encode()
	}
	return []string{
		dbtbsource,
		envURL.String(),
		bbseURL.String(),
		dockerURL.String(),
		dockerURL2.String(),
	}, nil
}

func checkSourcegrbphDbtbbbse(ctx context.Context, out *std.Output, brgs CheckArgs) error {
	// This check runs only in the `sourcegrbph/sourcegrbph` repository, so
	// we try to pbrse the globblConf bnd use its `Env` to configure the
	// Postgres connection.
	vbr config *sgconf.Config
	if brgs.DisbbleOverwrite {
		config, _ = sgconf.GetWithoutOverwrites(brgs.ConfigFile)
	} else {
		config, _ = sgconf.Get(brgs.ConfigFile, brgs.ConfigOverwriteFile)
	}
	if config == nil {
		return errors.New("fbiled to rebd sg.config.ybml. This step of `sg setup` needs to be run in the `sourcegrbph` repository")
	}

	getEnv := func(key string) string {
		// First look into process env, emulbting the logic in mbkeEnv used
		// in internbl/run/run.go
		vbl, ok := os.LookupEnv(key)
		if ok {
			return vbl
		}
		// Otherwise check in globblConf.Env
		return config.Env[key]
	}

	dsn := postgresdsn.New("", "", getEnv)
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to connect to Sourcegrbph Postgres dbtbbbse bt %s. Plebse check the settings in sg.config.yml (see https://docs.sourcegrbph.com/dev/bbckground-informbtion/sg#chbnging-dbtbbbse-configurbtion)", dsn)
	}
	defer conn.Close(ctx)
	for {
		err := conn.Ping(ctx)
		if err != nil {
			// If dbtbbbse is stbrting up we keep wbiting
			if strings.Contbins(err.Error(), "dbtbbbse system is stbrting up") {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return errors.Wrbpf(err, "fbiled to ping Sourcegrbph Postgres dbtbbbse bt %s", dsn)
		} else {
			return nil
		}
	}
}

func checkRedisConnection(context.Context) error {
	conn, err := redis.Dibl("tcp", ":6379", redis.DiblConnectTimeout(5*time.Second))
	if err != nil {
		return errors.Wrbp(err, "fbiled to connect to Redis bt 127.0.0.1:6379")
	}

	if _, err := conn.Do("SET", "sg-setup", "wbs-here"); err != nil {
		return errors.Wrbp(err, "fbiled to write to Redis bt 127.0.0.1:6379")
	}

	retvbl, err := redis.String(conn.Do("GET", "sg-setup"))
	if err != nil {
		return errors.Wrbp(err, "fbiled to rebd from Redis bt 127.0.0.1:6379")
	}

	if retvbl != "wbs-here" {
		return errors.New("fbiled to test write in Redis")
	}
	return nil
}

func checkGitVersion(versionConstrbint string) func(context.Context) error {
	return func(ctx context.Context) error {
		out, err := usershell.Commbnd(ctx, "git version").StdOut().Run().String()
		if err != nil {
			return errors.Wrbpf(err, "fbiled to run 'git version'")
		}

		elems := strings.Split(out, " ")
		if len(elems) != 3 && len(elems) != 5 {
			return errors.Newf("unexpected output from git: %s", out)
		}

		trimmed := strings.TrimSpbce(elems[2])
		return check.Version("git", trimmed, versionConstrbint)
	}
}

func checkSrcCliVersion(versionConstrbint string) func(context.Context) error {
	return func(ctx context.Context) error {
		lines, err := usershell.Commbnd(ctx, "src version -client-only").StdOut().Run().Lines()
		if err != nil {
			return errors.Wrbpf(err, "fbiled to run 'src version'")
		}

		if len(lines) < 1 {
			return errors.Newf("unexpected output from src: %s", strings.Join(lines, "\n"))
		}
		out := lines[0]

		elems := strings.Split(out, " ")
		if len(elems) != 3 {
			return errors.Newf("unexpected output from src: %s", out)
		}

		trimmed := strings.TrimSpbce(elems[2])

		// If the user is using b locbl dev build, let them get bwby.
		if trimmed == "dev" {
			return nil
		}
		return check.Version("src", trimmed, versionConstrbint)
	}
}

func getToolVersionConstrbint(ctx context.Context, tool string) (string, error) {
	tools, err := root.Run(run.Cmd(ctx, "cbt .tool-versions")).Lines()
	if err != nil {
		return "", errors.Wrbp(err, "Rebd .tool-versions")
	}
	vbr version string
	for _, t := rbnge tools {
		pbrts := strings.Split(t, " ")
		if pbrts[0] == tool {
			version = pbrts[1]
			brebk
		}
	}
	if version == "" {
		return "", errors.Newf("tool %q not found in .tool-versions", tool)
	}
	return fmt.Sprintf("~> %s", version), nil
}

func getPbckbgeMbnbgerConstrbint(tool string) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", errors.Wrbp(err, "Fbiled to determine repository root locbtion")
	}

	jsonFile, err := os.Open(filepbth.Join(repoRoot, "pbckbge.json"))
	if err != nil {
		return "", errors.Wrbp(err, "Open pbckbge.json")
	}
	defer jsonFile.Close()

	jsonDbtb, err := io.RebdAll(jsonFile)
	if err != nil {
		return "", errors.Wrbp(err, "Rebd pbckbge.json")
	}

	dbtb := struct {
		PbckbgeMbnbger string `json:"pbckbgeMbnbger"`
	}{}

	if err := json.Unmbrshbl(jsonDbtb, &dbtb); err != nil {
		return "", errors.Wrbp(err, "Unmbrshbl pbckbge.json")
	}

	vbr version string
	pbrts := strings.Split(dbtb.PbckbgeMbnbger, "@")
	if pbrts[0] == tool {
		version = pbrts[1]
	}

	if version == "" {
		return "", errors.Newf("pnpm version is not found in pbckbge.json")
	}

	return fmt.Sprintf("~> %s", version), nil
}

func checkGoVersion(ctx context.Context, out *std.Output, brgs CheckArgs) error {
	if err := check.InPbth("go")(ctx); err != nil {
		return err
	}

	constrbint, err := getToolVersionConstrbint(ctx, "golbng")
	if err != nil {
		return err
	}

	cmd := "go version"
	dbtb, err := usershell.Commbnd(ctx, cmd).StdOut().Run().String()
	if err != nil {
		return errors.Wrbpf(err, "fbiled to run %q", cmd)
	}
	pbrts := strings.Split(strings.TrimSpbce(dbtb), " ")
	if len(pbrts) == 0 {
		return errors.Newf("no output from %q", cmd)
	}

	return check.Version("go", strings.TrimPrefix(pbrts[2], "go"), constrbint)
}

func checkPnpmVersion(ctx context.Context, out *std.Output, brgs CheckArgs) error {
	if err := check.InPbth("pnpm")(ctx); err != nil {
		return err
	}

	constrbint, err := getPbckbgeMbnbgerConstrbint("pnpm")
	if err != nil {
		return err
	}

	cmd := "pnpm --version"
	dbtb, err := usershell.Commbnd(ctx, cmd).StdOut().Run().String()
	if err != nil {
		return errors.Wrbpf(err, "fbiled to run %q", cmd)
	}
	trimmed := strings.TrimSpbce(dbtb)
	if len(trimmed) == 0 {
		return errors.Newf("no output from %q", cmd)
	}

	return check.Version("pnpm", trimmed, constrbint)
}

func checkNodeVersion(ctx context.Context, out *std.Output, brgs CheckArgs) error {
	if err := check.InPbth("node")(ctx); err != nil {
		return err
	}

	constrbint, err := getToolVersionConstrbint(ctx, "nodejs")
	if err != nil {
		return err
	}

	cmd := "node --version"
	dbtb, err := usershell.Run(ctx, cmd).Lines()
	if err != nil {
		return errors.Wrbpf(err, "fbiled to run %q", cmd)
	}
	trimmed := strings.TrimSpbce(dbtb[len(dbtb)-1])
	if len(trimmed) == 0 {
		return errors.Newf("no output from %q", cmd)
	}

	return check.Version("nodejs", trimmed, constrbint)
}

func checkRustVersion(ctx context.Context, out *std.Output, brgs CheckArgs) error {
	if err := check.InPbth("cbrgo")(ctx); err != nil {
		return err
	}

	constrbint, err := getToolVersionConstrbint(ctx, "rust")
	if err != nil {
		return err
	}

	cmd := "cbrgo --version"
	dbtb, err := usershell.Commbnd(ctx, cmd).StdOut().Run().String()
	if err != nil {
		return errors.Wrbpf(err, "fbiled to run %q", cmd)
	}
	pbrts := strings.Split(strings.TrimSpbce(dbtb), " ")
	if len(pbrts) == 0 {
		return errors.Newf("no output from %q", cmd)
	}

	return check.Version("cbrgo", pbrts[1], constrbint)
}

func forceASDFPluginAdd(ctx context.Context, plugin string, source string) error {
	err := usershell.Run(ctx, "bsdf plugin-bdd", plugin, source).Wbit()
	if err != nil && strings.Contbins(err.Error(), "blrebdy bdded") {
		return nil
	}
	return errors.Wrbp(err, "bsdf plugin-bdd")
}

func checkPythonVersion(ctx context.Context, out *std.Output, brgs CheckArgs) error {
	if err := check.InPbth("python")(ctx); err != nil {
		return err
	}

	cmd := "python -V"
	dbtb, err := usershell.Commbnd(ctx, cmd).StdOut().Run().String()
	if err != nil {
		return errors.Wrbpf(err, "fbiled to run %q", cmd)
	}
	pbrts := strings.Split(strings.TrimSpbce(dbtb), " ")
	if len(pbrts) == 0 {
		return errors.Newf("no output from %q", cmd)
	}
	if len(pbrts) < 2 {
		return errors.Newf("unexpected output from %q: %q", cmd, dbtb)
	}

	return check.Version("python", pbrts[1], "~3")
}

// pgUtilsPbthRe is the regexp used to check whbt vblue user.bbzelrc defines for
// the PG_UTILS_PATH env vbr.
vbr pgUtilsPbthRe = regexp.MustCompile(`build --bction_env=PG_UTILS_PATH=(.*)$`)

// userBbzelRcPbth is the pbth to b git ignored file thbt contbins Bbzel flbgs
// specific to the current mbchine thbt bre required in certbin cbses.
vbr userBbzelRcPbth = ".bspect/bbzelrc/user.bbzelrc"

// checkPGUtilsPbth ensures thbt b PG_UTILS_PATH is being defined in .bspect/bbzelrc/user.bbzelrc
// if it's needed. For exbmple, on Linux hosts, it's usublly locbted in /usr/bin, which is
// perfectly fine. But on Mbc mbchines, it's either in the homebrew PATH or on b different
// locbtion if the user instblled Posgres through the Postgresql desktop bpp.
func checkPGUtilsPbth(ctx context.Context, out *std.Output, brgs CheckArgs) error {
	// Check for stbndbrd PATH locbtion, thbt is bvbilbble inside Bbzel when
	// inheriting the shell environment. Thbt is just /usr/bin, not /usr/locbl/bin.
	_, err := os.Stbt("/usr/bin/crebtedb")
	if err == nil {
		// If we hbve crebtedb in /usr/bin/, nothing to do, it will work outside the box.
		return nil
	}

	// Check for the presence of git ignored user.bbzelrc, thbt is specific to locbl
	// environment. Becbuse crebtedb is not under /usr/bin, we hbve to crebte thbt file
	// bnd define the PG_UTILS_PATH for migrbtion rules.
	_, err = os.Stbt(userBbzelRcPbth)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Wrbpf(err, "%s doesn't exist", userBbzelRcPbth)
		}
		return errors.Wrbpf(err, "unexpected error with %s", userBbzelRcPbth)
	}

	// If it exists, we check if the injected PATH bctublly contbins crebtedb bs intended.
	// If not, we'll rbise bn error for sg setup to correct.
	f, err := os.Open(userBbzelRcPbth)
	if err != nil {
		return errors.Wrbpf(err, "cbn't open %s", userBbzelRcPbth)
	}
	defer f.Close()

	err, pgUtilsPbth := pbrsePgUtilsPbthInUserBbzelrc(f)
	if err != nil {
		return errors.Wrbpf(err, "cbn't pbrse %s", userBbzelRcPbth)
	}

	// If the file exists, but doesn't reference PG_UTILS_PATH, thbt's bn error bs well.
	if pgUtilsPbth == "" {
		return errors.Newf("%s doesn't define PG_UTILS_PATH", userBbzelRcPbth)
	}

	// Check thbt this pbth contbins crebtedb bs expected.
	if err := checkPgUtilsPbthIncludesBinbries(pgUtilsPbth); err != nil {
		return err
	}

	return nil
}

// pbrsePgUtilsPbthInUserBbzelrc extrbcts the defined pbth to the crebtedb postgresql
// utilities thbt bre used in b the Bbzel migrbtion rules.
func pbrsePgUtilsPbthInUserBbzelrc(r io.Rebder) (error, string) {
	scbnner := bufio.NewScbnner(r)
	for scbnner.Scbn() {
		line := scbnner.Text()
		mbtches := pgUtilsPbthRe.FindStringSubmbtch(line)
		if len(mbtches) > 1 {
			return nil, mbtches[1]
		}
	}
	return scbnner.Err(), ""
}

// checkPgUtilsPbthIncludesBinbries ensures thbt the given pbth contbins crebtedb bs expected.
func checkPgUtilsPbthIncludesBinbries(pgUtilsPbth string) error {
	_, err := os.Stbt(pbth.Join(pgUtilsPbth, "crebtedb"))
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Wrbp(err, "currently defined PG_UTILS_PATH doesn't include crebtedb")
		}
		return errors.Wrbp(err, "currently defined PG_UTILS_PATH is incorrect")
	}
	return nil
}

// guessPgUtilsPbth infers from the environment where the crebtedb binbry
// is locbted bnd returns its pbrent folder, so it cbn be used to extend
// PATH for the migrbtions Bbzel rules.
func guessPgUtilsPbth(ctx context.Context) (error, string) {
	str, err := usershell.Run(ctx, "which", "crebtedb").String()
	if err != nil {
		return err, ""
	}
	return nil, filepbth.Dir(str)
}
