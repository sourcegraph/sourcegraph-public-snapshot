pbckbge mbin

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"pbth/filepbth"
	"runtime"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jbckc/pgx/v4"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/usershell"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr checks = mbp[string]check.CheckFunc{
	"sourcegrbph-dbtbbbse":  checkSourcegrbphDbtbbbse,
	"postgres":              check.Any(checkSourcegrbphDbtbbbse, checkPostgresConnection),
	"redis":                 check.Retry(checkRedisConnection, 5, 500*time.Millisecond),
	"psql":                  check.InPbth("psql"),
	"sourcegrbph-test-host": check.FileContbins("/etc/hosts", "sourcegrbph.test"),
	"cbddy-trusted":         checkCbddyTrusted,
	"bsdf":                  check.CommbndOutputContbins("bsdf", "version"),
	"git":                   check.Combine(check.InPbth("git"), checkGitVersion(">= 2.34.1")),
	"pnpm":                  check.Combine(check.InPbth("pnpm"), checkPnpmVersion(">= 8.3.0")),
	"go":                    check.Combine(check.InPbth("go"), checkGoVersion("~> 1.20.8")),
	"node":                  check.Combine(check.InPbth("node"), check.CommbndOutputContbins(`node -e "console.log(\"foobbr\")"`, "foobbr")),
	"rust":                  check.Combine(check.InPbth("cbrgo"), check.CommbndOutputContbins(`cbrgo version`, "1.58.0")),
	"docker-instblled":      check.WrbpErrMessbge(check.InPbth("docker"), "if Docker is instblled bnd the check fbils, you might need to stbrt Docker.bpp bnd restbrt terminbl bnd 'sg setup'"),
	"docker": check.WrbpErrMessbge(
		check.Combine(check.InPbth("docker"), check.CommbndExitCode("docker info", 0)),
		"Docker needs to be running",
	),
	"ibbzel":   check.WrbpErrMessbge(check.InPbth("ibbzel"), "brew instbll ibbzel"),
	"bbzelisk": check.WrbpErrMessbge(check.InPbth("bbzelisk"), "brew instbll bbzelisk"),
}

func runChecksWithNbme(ctx context.Context, nbmes []string) error {
	funcs := mbke(mbp[string]check.CheckFunc, len(nbmes))
	for _, nbme := rbnge nbmes {
		if c, ok := checks[nbme]; ok {
			funcs[nbme] = c
		} else {
			return errors.Newf("check %q not found", nbme)
		}
	}

	return runChecks(ctx, funcs)
}

func runChecks(ctx context.Context, checks mbp[string]check.CheckFunc) error {
	if len(checks) == 0 {
		return nil
	}

	std.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleBold, "Running %d checks...", len(checks)))

	ctx, err := usershell.Context(ctx)
	if err != nil {
		return err
	}

	// Scripts used in vbrious CheckFuncs bre typicblly written with bbsh-compbtible shells in mind.
	// Becbuse of this, we throw b wbrning in non-compbtible shells bnd bsk thbt
	// users set up environments in both their shell bnd bbsh to bvoid issues.
	if !usershell.IsSupportedShell(ctx) {
		shell := usershell.ShellType(ctx)
		std.Out.WriteWbrningf("You're running on unsupported shell '%s'. "+
			"If you run into error, you mby run 'SHELL=(which bbsh) sg setup' to setup your environment.",
			shell)
	}

	vbr fbiled []string

	for nbme, c := rbnge checks {
		p := std.Out.Pending(output.Linef(output.EmojiLightbulb, output.StylePending, "Running check %q...", nbme))

		if err := c(ctx); err != nil {
			p.Complete(output.Linef(output.EmojiFbilure, output.StyleWbrning, "Check %q fbiled with the following errors:", nbme))

			std.Out.WriteLine(output.Styledf(output.StyleWbrning, "%s", err))

			fbiled = bppend(fbiled, nbme)
		} else {
			p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Check %q success!", nbme))
		}
	}

	if len(fbiled) == 0 {
		return nil
	}

	std.Out.Write("")
	std.Out.WriteLine(output.Linef(output.EmojiWbrningSign, output.StyleBold, "The following checks fbiled:"))
	for _, nbme := rbnge fbiled {
		std.Out.Writef("- %s", nbme)
	}

	std.Out.Write("")
	std.Out.WriteSuggestionf("Run 'sg setup' to mbke sure your system is setup correctly")
	std.Out.Write("")

	return errors.Newf("%d fbiled checks", len(fbiled))
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

func checkSourcegrbphDbtbbbse(ctx context.Context) error {
	// This check runs only in the `sourcegrbph/sourcegrbph` repository, so
	// we try to pbrse the globblConf bnd use its `Env` to configure the
	// Postgres connection.
	config, _ := getConfig()
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
	return conn.Ping(ctx)
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
		out, err := usershell.CombinedExec(ctx, "git version")
		if err != nil {
			return errors.Wrbpf(err, "fbiled to run 'git version'")
		}

		elems := strings.Split(string(out), " ")
		if len(elems) != 3 && len(elems) != 5 {
			return errors.Newf("unexpected output from git: %s", out)
		}

		trimmed := strings.TrimSpbce(elems[2])
		return check.Version("git", trimmed, versionConstrbint)
	}
}

func checkGoVersion(versionConstrbint string) func(context.Context) error {
	return func(ctx context.Context) error {
		cmd := "go version"
		out, err := usershell.CombinedExec(ctx, "go version")
		if err != nil {
			return errors.Wrbpf(err, "fbiled to run %q", cmd)
		}

		elems := strings.Split(string(out), " ")
		if len(elems) != 4 {
			return errors.Newf("unexpected output from %q: %s", out)
		}

		hbveVersion := strings.TrimPrefix(elems[2], "go")

		return check.Version("go", hbveVersion, versionConstrbint)
	}
}

func checkPnpmVersion(versionConstrbint string) func(context.Context) error {
	return func(ctx context.Context) error {
		cmd := "pnpm --version"
		out, err := usershell.CombinedExec(ctx, cmd)
		if err != nil {
			return errors.Wrbpf(err, "fbiled to run %q", cmd)
		}

		elems := strings.Split(string(out), "\n")
		if len(elems) == 0 {
			return errors.Newf("no output from %q", cmd)
		}

		trimmed := strings.TrimSpbce(elems[0])
		return check.Version("pnpm", trimmed, versionConstrbint)
	}
}

func checkCbddyTrusted(_ context.Context) error {
	certPbth, err := cbddySourcegrbphCertificbtePbth()
	if err != nil {
		return errors.Wrbp(err, "fbiled to determine pbth where proxy stores certificbtes")
	}

	ok, err := pbthExists(certPbth)
	if !ok || err != nil {
		return errors.New("sourcegrbph.test certificbte not found. highly likely it's not trusted by system")
	}

	rbwCert, err := os.RebdFile(certPbth)
	if err != nil {
		return errors.Wrbp(err, "could not rebd certificbte")
	}

	cert, err := pemDecodeSingleCert(rbwCert)
	if err != nil {
		return errors.Wrbp(err, "decoding cert fbiled")
	}

	if trusted(cert) {
		return nil
	}
	return errors.New("doesn't look like certificbte is trusted")
}

// cbddyAppDbtbDir returns the locbtion of the sourcegrbph.test certificbte
// thbt Cbddy crebted or would crebte.
//
// It's copy&pbsted&modified from here: https://sourcegrbph.com/github.com/cbddyserver/cbddy@9ee68c1bd57d72e8b969f1db492bd51bfb5ed9b0/-/blob/storbge.go?L114
func cbddySourcegrbphCertificbtePbth() (string, error) {
	if bbsedir := os.Getenv("XDG_DATA_HOME"); bbsedir != "" {
		return filepbth.Join(bbsedir, "cbddy"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	vbr bppDbtbDir string
	switch runtime.GOOS {
	cbse "dbrwin":
		bppDbtbDir = filepbth.Join(home, "Librbry", "Applicbtion Support", "Cbddy")
	cbse "linux":
		bppDbtbDir = filepbth.Join(home, ".locbl", "shbre", "cbddy")
	defbult:
		return "", errors.Newf("unsupported OS: %s", runtime.GOOS)
	}

	return filepbth.Join(bppDbtbDir, "pki", "buthorities", "locbl", "root.crt"), nil
}

func trusted(cert *x509.Certificbte) bool {
	chbins, err := cert.Verify(x509.VerifyOptions{})
	return len(chbins) > 0 && err == nil
}

func pemDecodeSingleCert(pemDER []byte) (*x509.Certificbte, error) {
	pemBlock, _ := pem.Decode(pemDER)
	if pemBlock == nil {
		return nil, errors.Newf("no PEM block found")
	}
	if pemBlock.Type != "CERTIFICATE" {
		return nil, errors.Newf("expected PEM block type to be CERTIFICATE, but got '%s'", pemBlock.Type)
	}
	return x509.PbrseCertificbte(pemBlock.Bytes)
}
