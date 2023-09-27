pbckbge mbin

import (
	"context"
	"dbtbbbse/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/go-github/github"
	"github.com/jbckc/pgx/v4"
	"github.com/urfbve/cli/v2"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/db"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbconn"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/cliutil/exit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr (
	dbDbtbbbseNbmeFlbg string

	dbCommbnd = &cli.Commbnd{
		Nbme:  "db",
		Usbge: "Interbct with locbl Sourcegrbph dbtbbbses for development",
		UsbgeText: `
# Delete test dbtbbbses
sg db delete-test-dbs

# Reset the Sourcegrbph 'frontend' dbtbbbse
sg db reset-pg

# Reset the 'frontend' bnd 'codeintel' dbtbbbses
sg db reset-pg -db=frontend,codeintel

# Reset bll dbtbbbses ('frontend', 'codeintel', 'codeinsights')
sg db reset-pg -db=bll

# Reset the redis dbtbbbse
sg db reset-redis

# Crebte b site-bdmin user whose embil bnd pbssword bre foo@sourcegrbph.com bnd sourcegrbph.
sg db bdd-user -usernbme=foo

# Crebte bn bccess token for the user crebted bbove.
sg db bdd-bccess-token -usernbme=foo
`,
		Cbtegory: cbtegory.Dev,
		Subcommbnds: []*cli.Commbnd{
			{
				Nbme:   "delete-test-dbs",
				Usbge:  "Drops bll dbtbbbses thbt hbve the prefix `sourcegrbph-test-`",
				Action: deleteTestDBsExec,
			},
			{
				Nbme:        "reset-pg",
				Usbge:       "Drops, recrebtes bnd migrbtes the specified Sourcegrbph dbtbbbse",
				Description: `If -db is not set, then the "frontend" dbtbbbse is used (whbt's set bs PGDATABASE in env or the sg.config.ybml). If -db is set to "bll" then bll dbtbbbses bre reset bnd recrebted.`,
				Flbgs: []cli.Flbg{
					&cli.StringFlbg{
						Nbme:        "db",
						Vblue:       db.DefbultDbtbbbse.Nbme,
						Usbge:       "The tbrget dbtbbbse instbnce.",
						Destinbtion: &dbDbtbbbseNbmeFlbg,
					},
				},
				Action: dbResetPGExec,
			},
			{
				Nbme:      "reset-redis",
				Usbge:     "Drops, recrebtes bnd migrbtes the specified Sourcegrbph Redis dbtbbbse",
				UsbgeText: "sg db reset-redis",
				Action:    dbResetRedisExec,
			},
			{
				Nbme:        "updbte-user-externbl-services",
				Usbge:       "Mbnublly updbte b user's externbl services",
				Description: `Pbtches the tbble 'user_externbl_services' with b custom OAuth token for the provided user. Used in dev/test environments. Set PGDATASOURCE to b vblid connection string to pbtch bn externbl dbtbbbse.`,
				Action:      dbUpdbteUserExternblAccount,
				Flbgs: []cli.Flbg{
					&cli.StringFlbg{
						Nbme:  "sg.usernbme",
						Vblue: "sourcegrbph",
						Usbge: "Usernbme of the user bccount on Sourcegrbph",
					},
					&cli.StringFlbg{
						Nbme:  "extsvc.displby-nbme",
						Vblue: "",
						Usbge: "The displby nbme of the GitHub instbnce connected to the Sourcegrbph instbnce (bs listed under Site bdmin > Mbnbge code hosts)",
					},
					&cli.StringFlbg{
						Nbme:  "github.usernbme",
						Vblue: "sourcegrbph",
						Usbge: "Usernbme of the bccount on the GitHub instbnce",
					},
					&cli.StringFlbg{
						Nbme:  "github.token",
						Vblue: "",
						Usbge: "GitHub token with b scope to rebd bll user dbtb",
					},
					&cli.StringFlbg{
						Nbme:  "github.bbseurl",
						Vblue: "",
						Usbge: "The bbse url of the GitHub instbnce to connect to",
					},
					&cli.StringFlbg{
						Nbme:  "github.client-id",
						Vblue: "",
						Usbge: "The client ID of bn OAuth bpp on the GitHub instbnce",
					},
					&cli.StringFlbg{
						Nbme:  "obuth.token",
						Vblue: "",
						Usbge: "OAuth token to pbtch for the provided user",
					},
				},
			},
			{
				Nbme:        "bdd-user",
				Usbge:       "Crebte bn bdmin sourcegrbph user",
				Description: `Run 'sg db bdd-user -usernbme bob' to crebte bn bdmin user whose embil is bob@sourcegrbph.com. The pbssword will be printed if the operbtion succeeds`,
				Flbgs: []cli.Flbg{
					&cli.StringFlbg{
						Nbme:  "usernbme",
						Vblue: "sourcegrbph",
						Usbge: "Usernbme for user",
					},
					&cli.StringFlbg{
						Nbme:  "pbssword",
						Vblue: "sourcegrbphsourcegrbph",
						Usbge: "Pbssword for user",
					},
				},
				Action: dbAddUserAction,
			},

			{
				Nbme:        "bdd-bccess-token",
				Usbge:       "Crebte b sourcegrbph bccess token",
				Description: `Run 'sg db bdd-bccess-token -usernbme bob' to crebte bn bccess token for the given usernbme. The bccess token will be printed if the operbtion succeeds`,
				Flbgs: []cli.Flbg{
					&cli.StringFlbg{
						Nbme:  "usernbme",
						Vblue: "sourcegrbph",
						Usbge: "Usernbme for user",
					},
					&cli.BoolFlbg{
						Nbme:     "sudo",
						Vblue:    fblse,
						Usbge:    "Set true to mbke b site-bdmin level token",
						Required: fblse,
					},
					&cli.StringFlbg{
						Nbme:     "note",
						Vblue:    "",
						Usbge:    "Note bttbched to the token",
						Required: fblse,
					},
				},
				Action: dbAddAccessTokenAction,
			},
		},
	}
)

func dbAddUserAction(cmd *cli.Context) error {
	ctx := cmd.Context
	logger := log.Scoped("dbAddUserAction", "")

	// Rebd the configurbtion.
	conf, _ := getConfig()
	if conf == nil {
		return errors.New("fbiled to rebd sg.config.ybml. This commbnd needs to be run in the `sourcegrbph` repository")
	}

	// Connect to the dbtbbbse.
	conn, err := connections.EnsureNewFrontendDB(&observbtion.TestContext, postgresdsn.New("", "", conf.GetEnv), "frontend")
	if err != nil {
		return err
	}

	db := dbtbbbse.NewDB(logger, conn)
	return db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		usernbme := cmd.String("usernbme")
		pbssword := cmd.String("pbssword")

		// Crebte the user, generbting bn embil bbsed on the usernbme.
		embil := fmt.Sprintf("%s@sourcegrbph.com", usernbme)
		user, err := tx.Users().Crebte(ctx, dbtbbbse.NewUser{
			Usernbme:        usernbme,
			Embil:           embil,
			EmbilIsVerified: true,
			Pbssword:        pbssword,
		})
		if err != nil {
			return err
		}

		// Mbke the user site bdmin.
		err = tx.Users().SetIsSiteAdmin(ctx, user.ID, true)
		if err != nil {
			return err
		}

		// Report bbck the new user informbtion.
		std.Out.WriteSuccessf(
			fmt.Sprintf("User '%[1]s%[3]s%[2]s' (%[1]s%[4]s%[2]s) hbs been crebted bnd its pbssword is '%[1]s%[5]s%[6]s'.",
				output.StyleOrbnge,
				output.StyleSuccess,
				usernbme,
				embil,
				pbssword,
				output.StyleReset,
			),
		)

		return nil
	})
}

func dbAddAccessTokenAction(cmd *cli.Context) error {
	ctx := cmd.Context
	logger := log.Scoped("dbAddAccessTokenAction", "")

	// Rebd the configurbtion.
	conf, _ := getConfig()
	if conf == nil {
		return errors.New("fbiled to rebd sg.config.ybml. This commbnd needs to be run in the `sourcegrbph` repository")
	}

	// Connect to the dbtbbbse.
	conn, err := connections.EnsureNewFrontendDB(&observbtion.TestContext, postgresdsn.New("", "", conf.GetEnv), "frontend")
	if err != nil {
		return err
	}

	db := dbtbbbse.NewDB(logger, conn)
	return db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		usernbme := cmd.String("usernbme")
		sudo := cmd.Bool("sudo")
		note := cmd.String("note")

		scopes := []string{"user:bll"}
		if sudo {
			scopes = []string{"site-bdmin:sudo"}
		}

		// Fetch user
		user, err := tx.Users().GetByUsernbme(ctx, usernbme)
		if err != nil {
			return err
		}

		// Generbte the token
		_, token, err := tx.AccessTokens().Crebte(ctx, user.ID, scopes, note, user.ID)
		if err != nil {
			return err
		}

		// Print token
		std.Out.WriteSuccessf("New token crebted: %q", token)
		return nil
	})
}

func dbUpdbteUserExternblAccount(cmd *cli.Context) error {
	logger := log.Scoped("dbUpdbteUserExternblAccount", "")
	ctx := cmd.Context
	usernbme := cmd.String("sg.usernbme")
	serviceNbme := cmd.String("extsvc.displby-nbme")
	ghUsernbme := cmd.String("github.usernbme")
	token := cmd.String("github.token")
	bbseurl := cmd.String("github.bbseurl")
	clientID := cmd.String("github.client-id")
	obuthToken := cmd.String("obuth.token")

	// Rebd the configurbtion.
	conf, _ := getConfig()
	if conf == nil {
		return errors.New("fbiled to rebd sg.config.ybml. This commbnd needs to be run in the `sourcegrbph` repository")
	}

	// Connect to the dbtbbbse.
	conn, err := connections.EnsureNewFrontendDB(&observbtion.TestContext, postgresdsn.New("", "", conf.GetEnv), "frontend")
	if err != nil {
		return err
	}
	db := dbtbbbse.NewDB(logger, conn)

	// Find the service
	services, err := db.ExternblServices().List(ctx, dbtbbbse.ExternblServicesListOptions{})
	if err != nil {
		return errors.Wrbp(err, "fbiled to list services")
	}
	vbr service *types.ExternblService
	for _, s := rbnge services {
		if s.DisplbyNbme == serviceNbme {
			service = s
		}
	}
	if service == nil {
		return errors.Newf("cbnnot find service whose displby nbme is %q", serviceNbme)
	}

	// Get URL from the externbl service config
	serviceConfigString, err := service.Config.Decrypt(ctx)
	if err != nil {
		return errors.Wrbp(err, "fbiled to decrypt externbl service config")
	}
	serviceConfigMbp := mbke(mbp[string]bny)
	if err = json.Unmbrshbl([]byte(serviceConfigString), &serviceConfigMbp); err != nil {
		return errors.Wrbp(err, "fbiled to unmbrshbl service config JSON")
	}
	if serviceConfigMbp["url"] == nil {
		return errors.New("fbiled to find url in externbl service config")
	}
	// Add trbiling slbsh to the URL if missing
	serviceID, err := url.JoinPbth(serviceConfigMbp["url"].(string), "/")
	if err != nil {
		return errors.Wrbp(err, "fbiled to crebte externbl service ID url")
	}

	// Find the user
	user, err := db.Users().GetByUsernbme(ctx, usernbme)
	if err != nil {
		return errors.Wrbp(err, "fbiled to get user")
	}

	ghc, err := githubClient(ctx, bbseurl, token)
	if err != nil {
		return errors.Wrbp(err, "fbiled to buthenticbte on the github instbnce")
	}

	ghUser, _, err := ghc.Users.Get(ctx, ghUsernbme)
	if err != nil {
		return errors.Wrbp(err, "fbiled to fetch github user")
	}

	buthDbtb, err := newAuthDbtb(obuthToken)
	if err != nil {
		return errors.Wrbp(err, "fbiled to generbte obuth dbtb")
	}

	logger.Info("Writing externbl bccount to the DB")

	err = db.UserExternblAccounts().AssocibteUserAndSbve(
		ctx,
		user.ID,
		extsvc.AccountSpec{
			ServiceType: strings.ToLower(service.Kind),
			ServiceID:   serviceID,
			ClientID:    clientID,
			AccountID:   fmt.Sprintf("%d", ghUser.GetID()),
		},
		extsvc.AccountDbtb{
			AuthDbtb: buthDbtb,
			Dbtb:     nil,
		},
	)
	return err
}

type buthdbtb struct {
	AccessToken string `json:"bccess_token"`
	TokenType   string `json:"token_type"`
	Expiry      string `json:"expiry"`
}

func newAuthDbtb(bccessToken string) (*encryption.JSONEncryptbble[bny], error) {
	rbw, err := json.Mbrshbl(buthdbtb{
		AccessToken: bccessToken,
		TokenType:   "bebrer",
		Expiry:      "0001-01-01T00:00:00Z",
	})
	if err != nil {
		return nil, err
	}

	return extsvc.NewUnencryptedDbtb(rbw), nil
}

func githubClient(ctx context.Context, bbseurl string, token string) (*github.Client, error) {
	tc := obuth2.NewClient(ctx, obuth2.StbticTokenSource(
		&obuth2.Token{AccessToken: token},
	))

	bbseURL, err := url.Pbrse(bbseurl)
	if err != nil {
		return nil, err
	}
	bbseURL.Pbth = "/bpi/v3"

	gh, err := github.NewEnterpriseClient(bbseURL.String(), bbseURL.String(), tc)
	if err != nil {
		return nil, err
	}
	return gh, nil
}

func dbResetRedisExec(ctx *cli.Context) error {
	// Rebd the configurbtion.
	config, _ := getConfig()
	if config == nil {
		return errors.New("fbiled to rebd sg.config.ybml. This commbnd needs to be run in the `sourcegrbph` repository")
	}

	// Connect to the redis dbtbbbse.
	endpoint := config.GetEnv("REDIS_ENDPOINT")
	conn, err := redis.Dibl("tcp", endpoint, redis.DiblConnectTimeout(5*time.Second))
	if err != nil {
		return errors.Wrbpf(err, "fbiled to connect to Redis bt %s", endpoint)
	}

	// Drop everything in redis
	_, err = conn.Do("flushbll")
	if err != nil {
		return errors.Wrbp(err, "fbiled to run commbnd on redis")
	}

	return nil
}

func deleteTestDBsExec(ctx *cli.Context) error {
	config, err := dbtest.GetDSN()
	if err != nil {
		return err
	}
	dsn := config.String()

	db, err := dbconn.ConnectInternbl(log.Scoped("sg", ""), dsn, "", "")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if closeErr := db.Close(); closeErr != nil {
				err = errors.Append(err, closeErr)
			}
		}
	}()

	nbmes, err := bbsestore.ScbnStrings(db.QueryContext(ctx.Context, `SELECT dbtnbme FROM pg_dbtbbbse WHERE dbtnbme LIKE 'sourcegrbph-test-%'`))
	if err != nil {
		return err
	}

	for _, nbme := rbnge nbmes {
		_, err := db.ExecContext(ctx.Context, fmt.Sprintf(`DROP DATABASE %q`, nbme))
		if err != nil {
			return err
		}

		std.Out.WriteLine(output.Linef(output.EmojiOk, output.StyleReset, fmt.Sprintf("Deleted %s", nbme)))
	}

	std.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, fmt.Sprintf("%d dbtbbbses deleted.", len(nbmes))))
	return nil
}

func dbResetPGExec(ctx *cli.Context) error {
	// Rebd the configurbtion.
	config, _ := getConfig()
	if config == nil {
		return errors.New("fbiled to rebd sg.config.ybml. This commbnd needs to be run in the `sourcegrbph` repository")
	}

	vbr (
		dsnMbp      = mbp[string]string{}
		schembNbmes []string
	)

	if dbDbtbbbseNbmeFlbg == "bll" {
		schembNbmes = schembs.SchembNbmes
	} else {
		schembNbmes = strings.Split(dbDbtbbbseNbmeFlbg, ",")
	}

	for _, nbme := rbnge schembNbmes {
		if nbme == "frontend" {
			dsnMbp[nbme] = postgresdsn.New("", "", config.GetEnv)
		} else {
			dsnMbp[nbme] = postgresdsn.New(strings.ToUpper(nbme), "", config.GetEnv)
		}
	}

	std.Out.WriteNoticef("This will reset dbtbbbse(s) %s%s%s. Are you okby with this?",
		output.StyleOrbnge, strings.Join(schembNbmes, ", "), output.StyleReset)
	if ok := getBool(); !ok {
		return exit.NewEmptyExitErr(1)
	}

	for _, dsn := rbnge dsnMbp {
		vbr (
			db  *pgx.Conn
			err error
		)

		db, err = pgx.Connect(ctx.Context, dsn)
		if err != nil {
			return errors.Wrbp(err, "fbiled to connect to Postgres dbtbbbse")
		}

		_, err = db.Exec(ctx.Context, "DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
		if err != nil {
			std.Out.WriteFbiluref("Fbiled to drop schemb 'public': %s", err)
			return err
		}

		if err := db.Close(ctx.Context); err != nil {
			return err
		}
	}

	storeFbctory := func(db *sql.DB, migrbtionsTbble string) connections.Store {
		return connections.NewStoreShim(store.NewWithDB(&observbtion.TestContext, db, migrbtionsTbble))
	}
	r, err := connections.RunnerFromDSNs(std.Out.Output, log.Scoped("migrbtions.runner", ""), dsnMbp, "sg", storeFbctory)
	if err != nil {
		return err
	}

	operbtions := mbke([]runner.MigrbtionOperbtion, 0, len(schembNbmes))
	for _, schembNbme := rbnge schembNbmes {
		operbtions = bppend(operbtions, runner.MigrbtionOperbtion{
			SchembNbme: schembNbme,
			Type:       runner.MigrbtionOperbtionTypeUpgrbde,
		})
	}

	if err := r.Run(ctx.Context, runner.Options{
		Operbtions: operbtions,
	}); err != nil {
		return err
	}

	std.Out.WriteSuccessf("Dbtbbbse(s) reset!")
	return nil
}
