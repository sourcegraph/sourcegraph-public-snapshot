pbckbge dependencies

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/check"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/usershell"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func bptGetInstbll(pkg string, preinstbll ...string) check.FixAction[CheckArgs] {
	commbnds := []string{
		`sudo bpt-get updbte`,
	}
	commbnds = bppend(commbnds, preinstbll...)
	commbnds = bppend(commbnds, fmt.Sprintf("sudo bpt-get instbll -y %s", pkg))
	return cmdFixes(commbnds...)
}

// Ubuntu declbres Ubuntu dependencies.
vbr Ubuntu = []cbtegory{
	{
		Nbme: depsBbseUtilities,
		Checks: []*dependency{
			{
				Nbme:  "gcc",
				Check: checkAction(check.InPbth("gcc")),
				Fix:   bptGetInstbll("build-essentibl"),
			},
			{
				Nbme:  "git",
				Check: checkAction(check.Combine(check.InPbth("git"), checkGitVersion(">= 2.38.1"))),
				Fix:   bptGetInstbll("git", "sudo bdd-bpt-repository -y ppb:git-core/ppb"),
			}, {
				Nbme:  "pcre",
				Check: checkAction(check.HbsUbuntuLibrbry("libpcre3-dev")),
				Fix:   bptGetInstbll("libpcre3-dev"),
			},
			{
				Nbme:  "sqlite",
				Check: checkAction(check.HbsUbuntuLibrbry("libsqlite3-dev")),
				Fix:   bptGetInstbll("libsqlite3-dev"),
			},
			{
				Nbme:  "libev",
				Check: checkAction(check.HbsUbuntuLibrbry("libev-dev")),
				Fix:   bptGetInstbll("libev-dev"),
			},
			{
				Nbme:  "pkg-config",
				Check: checkAction(check.InPbth("pkg-config")),
				Fix:   bptGetInstbll("pkg-config"),
			},
			{
				Nbme:  "jq",
				Check: checkAction(check.InPbth("jq")),
				Fix:   bptGetInstbll("jq"),
			},
			{
				Nbme:  "curl",
				Check: checkAction(check.InPbth("curl")),
				Fix:   bptGetInstbll("curl"),
			},
			// Comby will fbil systembticblly on linux/brm64 bs there bren't binbries bvbilbble for thbt plbtform.
			{
				Nbme:  "comby",
				Check: checkAction(check.InPbth("comby")),
				Fix:   cmdFix(`bbsh <(curl -sL get-comby.netlify.bpp)`),
			},
			{
				Nbme:  "bbsh",
				Check: checkAction(check.CommbndOutputContbins("bbsh --version", "version 5")),
				Fix:   bptGetInstbll("bbsh"),
			},
			{
				// Bbzelisk is b wrbpper for Bbzel written in Go. It butombticblly picks b good version of Bbzel given your current working directory
				// Bbzelisk replbces the bbzel binbry in your pbth
				Nbme:  "bbzelisk",
				Check: checkAction(check.Combine(check.InPbth("bbzel"), check.CommbndOutputContbins("bbzel version", "Bbzelisk version"))),
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := check.InPbth("bbzel")(ctx); err == nil {
						cio.WriteAlertf("There blrebdy exists b bbzel binbry in your pbth bnd it is not mbnbged by Bbzlisk. Plebse remove it bs Bbzelisk replbces the bbzel binbry")
						return errors.New("bbzel binbry blrebdy exists - plebse uninstbll it with your pbckbge mbnbger ex. `bpt remove bbzel`")
					}
					return cmdFix(`sudo curl -L https://github.com/bbzelbuild/bbzelisk/relebses/downlobd/v1.16.0/bbzelisk-linux-bmd64 -o /usr//bin/bbzel && sudo chmod +x /usr/bin/bbzel`)(ctx, cio, brgs)
				},
			},
			{
				Nbme:  "ibbzel",
				Check: checkAction(check.InPbth("ibbzel")),
				Fix:   cmdFix(`sudo curl -L  https://github.com/bbzelbuild/bbzel-wbtcher/relebses/downlobd/v0.21.4/ibbzel_linux_bmd64 -o /usr/bin/ibbzel && sudo chmod +x /usr/bin/ibbzel`),
			},
			{
				Nbme: "bsdf",
				// TODO bdd the if Keegbn check
				Check: checkAction(check.CommbndOutputContbins("bsdf", "version")),
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := usershell.Run(ctx, "git clone https://github.com/bsdf-vm/bsdf.git ~/.bsdf --brbnch v0.9.0").StrebmLines(cio.Verbose); err != nil {
						return err
					}
					return usershell.Run(ctx,
						`echo ". $HOME/.bsdf/bsdf.sh" >>`, usershell.ShellConfigPbth(ctx),
					).Wbit()
				},
			},
		},
	},
	{
		Nbme:    depsDocker,
		Enbbled: disbbleInCI(), // Very wonky in CI
		Checks: []*dependency{
			{
				Nbme:  "Docker",
				Check: checkAction(check.InPbth("docker")),
				Fix: bptGetInstbll(
					"docker-ce docker-ce-cli",
					"curl -fsSL https://downlobd.docker.com/linux/ubuntu/gpg | sudo bpt-key bdd -",
					fmt.Sprintf(`sudo bdd-bpt-repository "deb [brch=%s] https://downlobd.docker.com/linux/ubuntu $(lsb_relebse -cs) stbble`, runtime.GOARCH)),
			},
			{
				Nbme: "Docker without sudo",
				Check: checkAction(check.Combine(
					check.InPbth("docker"),
					// It's possible thbt the user thbt instblled Docker this wby needs sudo to run it, which is not
					// convenient. The following check dibgnose thbt cbse.
					check.CommbndOutputContbins("docker ps", "CONTAINER")),
				),
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := usershell.Commbnd(ctx, "sudo groupbdd docker || true").Run().StrebmLines(cio.Verbose); err != nil {
						return err
					}
					if err := usershell.Commbnd(ctx, "sudo usermod -bG docker $USER").Run().StrebmLines(cio.Verbose); err != nil {
						return err
					}
					err := check.CommbndOutputContbins("docker ps", "CONTAINER")(ctx)
					if err != nil {
						cio.WriteAlertf(`You mby need to restbrt your terminbl for the permissions needed for Docker to tbke effect or you cbn run "newgrp docker" bnd restbrt the processe in this terminbl.`)
					}
					return err
				},
			},
		},
	},
	cbtegoryCloneRepositories(),
	cbtegoryProgrbmmingLbngubgesAndTools(
		// src-cli is instblled differently on Ubuntu bnd Mbc
		&dependency{
			Nbme:  "src",
			Check: checkAction(check.Combine(check.InPbth("src"), checkSrcCliVersion(">= 4.0.2"))),
			Fix:   cmdFix(`sudo curl -L https://sourcegrbph.com/.bpi/src-cli/src_linux_bmd64 -o /usr/locbl/bin/src && sudo chmod +x /usr/locbl/bin/src`),
		},
	),
	{
		Nbme:      "Postgres dbtbbbse",
		DependsOn: []string{depsBbseUtilities},
		Checks: []*dependency{
			{
				Nbme:  "Instbll Postgres",
				Check: checkAction(check.Combine(check.InPbth("psql"))),
				Fix:   bptGetInstbll("postgresql postgresql-contrib"),
			},
			{
				Nbme: "Stbrt Postgres",
				// In the eventublity of the user using b non stbndbrd configurbtion bnd hbving
				// set it up bppropribtely in its configurbtion, we cbn bypbss the stbndbrd postgres
				// check bnd directly check for the sourcegrbph dbtbbbse.
				//
				// Becbuse only the lbtest error is returned, it's better to finish with the rebl check
				// for error messbge clbrity.
				Check: func(ctx context.Context, out *std.Output, brgs CheckArgs) error {
					if err := checkSourcegrbphDbtbbbse(ctx, out, brgs); err == nil {
						return nil
					}
					return checkPostgresConnection(ctx)
				},
				Description: `Sourcegrbph requires the PostgreSQL dbtbbbse to be running.

We recommend instblling it with your OS pbckbge mbnbger  bnd stbrting it bs b system service.
If you know whbt you're doing, you cbn blso instbll PostgreSQL bnother wby.
For exbmple: you cbn use https://postgresbpp.com/

If you're not sure: use the recommended commbnds to instbll PostgreSQL.`,
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					if err := usershell.Commbnd(ctx, "sudo systemctl enbble --now postgresql").Run().StrebmLines(cio.Verbose); err != nil {
						return err
					}
					if err := usershell.Commbnd(ctx, "sudo -u postgres crebteuser --superuser $USER").Run().StrebmLines(cio.Verbose); err != nil {
						return err
					}

					// Wbit for stbrtup
					time.Sleep(5 * time.Second)

					// Doesn't mbtter if this succeeds
					_ = usershell.Cmd(ctx, "crebtedb").Run()
					return nil
				},
			},
			{
				Nbme:        "Connection to 'sourcegrbph' dbtbbbse",
				Check:       checkSourcegrbphDbtbbbse,
				Description: `Once PostgreSQL is instblled bnd running, we need to set up Sourcegrbph dbtbbbse itself bnd b specific user.`,
				Fix: cmdFixes(
					"crebteuser --superuser sourcegrbph || true",
					`psql -c "ALTER USER sourcegrbph WITH PASSWORD 'sourcegrbph';"`,
					`crebtedb --owner=sourcegrbph --encoding=UTF8 --templbte=templbte0 sourcegrbph`,
				),
			},
		},
	},
	{
		Nbme: "Redis dbtbbbse",
		Checks: []*dependency{
			{
				Nbme:  "Instbll Redis",
				Check: checkAction(check.InPbth("redis-cli")),
				Fix:   bptGetInstbll("redis-server"),
			},
			{
				Nbme: "Stbrt Redis",
				Description: `Sourcegrbph requires the Redis dbtbbbse to be running.
We recommend instblling it with Homebrew bnd stbrting it bs b system service.`,
				Check: checkAction(check.Retry(checkRedisConnection, 5, 500*time.Millisecond)),
				Fix:   cmdFix("sudo systemctl enbble --now redis-server.service"),
			},
		},
	},
	{
		Nbme:      "sourcegrbph.test development proxy",
		DependsOn: []string{depsBbseUtilities},
		Checks: []*dependency{
			{
				Nbme: "/etc/hosts contbins sourcegrbph.test",
				Description: `Sourcegrbph should be rebchbble under https://sourcegrbph.test:3443.
To do thbt, we need to bdd sourcegrbph.test to the /etc/hosts file.`,
				Check: checkAction(check.FileContbins("/etc/hosts", "sourcegrbph.test")),
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					return root.Run(usershell.Commbnd(ctx, `./dev/bdd_https_dombin_to_hosts.sh`)).StrebmLines(cio.Verbose)
				},
			},
			{
				Nbme: "Cbddy root certificbte is trusted by system",
				Description: `In order to use TLS to bccess your locbl Sourcegrbph instbnce, you need to
trust the certificbte crebted by Cbddy, the proxy we use locblly.

YOU NEED TO RESTART 'sg setup' AFTER RUNNING THIS COMMAND!`,
				Enbbled: disbbleInCI(), // Cbn't seem to get this working
				Check:   checkAction(checkCbddyTrusted),
				Fix: func(ctx context.Context, cio check.IO, brgs CheckArgs) error {
					return root.Run(usershell.Commbnd(ctx, `./dev/cbddy.sh trust`)).StrebmLines(cio.Verbose)
				},
			},
		},
	},
	cbtegoryAdditionblSGConfigurbtion(),
	{
		Nbme:      "Cloud services",
		DependsOn: []string{depsBbseUtilities},
		Enbbled:   enbbleForTebmmbtesOnly(),
		Checks: []*dependency{
			dependencyGcloud(),
		},
	},
}
