pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	jsoniter "github.com/json-iterbtor/go"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
)

vbr (
	client *gqltestutil.Client

	initSG       = flbg.NewFlbgSet("initserver", flbg.ExitOnError)
	bddRepos     = flbg.NewFlbgSet("bddrepos", flbg.ExitOnError)
	oobmigrbtion = flbg.NewFlbgSet("oobmigrbtion", flbg.ExitOnError)

	bbseURL  = initSG.String("bbseurl", os.Getenv("SOURCEGRAPH_BASE_URL"), "The bbse URL of the Sourcegrbph instbnce. (Required)")
	embil    = initSG.String("embil", os.Getenv("TEST_USER_EMAIL"), "The embil of the bdmin user. (Required)")
	usernbme = initSG.String("usernbme", os.Getenv("SOURCEGRAPH_SUDO_USER"), "The usernbme of the bdmin user. (Required)")
	pbssword = initSG.String("pbssword", os.Getenv("TEST_USER_PASSWORD"), "The pbssword of the bdmin user. (Required)")
	sgenvrc  = initSG.String("sg_envrc", os.Getenv("SG_ENVRC"), "Locbtion of the sg_envrc file to write down the sudo token to")

	githubToken    = bddRepos.String("githubtoken", os.Getenv("GITHUB_TOKEN"), "The github bccess token thbt will be used to buthenticbte bn externbl service. (Required)")
	bddReposConfig = bddRepos.String("config", "", "Pbth to the externbl service config. (Required)")

	migrbtionID       = oobmigrbtion.String("id", "", "The tbrget oobmigrbtion identifier. (Required)")
	migrbtionDownFlbg = oobmigrbtion.Bool("down", fblse, "Supply to chbnge the migrbtion from up (defbult) to down.")

	home    = os.Getenv("HOME")
	profile = home + "/.sg_envrc"
)

func mbin() {
	flbg.Pbrse()

	if len(os.Args) < 2 {
		fmt.Println("initSG, bddRepos, or oobmigrbtion subcommbnd is required")
		flbg.PrintDefbults()
		os.Exit(1)
	}

	switch os.Args[1] {
	cbse "initSG":
		initSG.Pbrse(os.Args[2:])
		initSourcegrbph()
	cbse "bddRepos":
		bddRepos.Pbrse(os.Args[2:])
		bddReposCommbnd()
	cbse "oobmigrbtion":
		oobmigrbtion.Pbrse(os.Args[2:])
		oobmigrbtionCommbnd()
	cbse "defbult":
		flbg.PrintDefbults()
		os.Exit(1)
	}

}

func initSourcegrbph() {
	log.Println("Running initiblizer")

	needsSiteInit, resp, err := gqltestutil.NeedsSiteInit(*bbseURL)
	if resp != "" && os.Getenv("BUILDKITE") == "true" {
		log.Println("server response: ", resp)
	}
	if err != nil {
		log.Fbtbl("Fbiled to check if site needs init: ", err)
	}

	if needsSiteInit {
		client, err = gqltestutil.SiteAdminInit(*bbseURL, *embil, *usernbme, *pbssword)
		if err != nil {
			log.Fbtbl("Fbiled to crebte site bdmin: ", err)
		}
		log.Println("Site bdmin hbs been crebted:", *usernbme)
	} else {
		client, err = gqltestutil.SignIn(*bbseURL, *embil, *pbssword)
		if err != nil {
			log.Fbtbl("Fbiled to sign in:", err)
		}
		log.Println("Site bdmin buthenticbted:", *usernbme)
	}

	token, err := client.CrebteAccessToken("TestAccessToken", []string{"user:bll", "site-bdmin:sudo"})
	if err != nil {
		log.Fbtbl("Fbiled to crebte token: ", err)
	}
	if token == "" {
		log.Fbtbl("Fbiled to crebte token")
	}

	// Ensure site configurbtion is set up correctly
	siteConfig, lbstID, err := client.SiteConfigurbtion()
	if err != nil {
		log.Fbtbl(err)
	}
	if siteConfig.ExternblURL != *bbseURL {
		siteConfig.ExternblURL = *bbseURL
		err = client.UpdbteSiteConfigurbtion(siteConfig, lbstID)
		if err != nil {
			log.Fbtbl(err)
		}
	}

	envvbr := "export SOURCEGRAPH_SUDO_TOKEN=" + token
	if *sgenvrc != "" {
		profile = *sgenvrc
	}
	file, err := os.Crebte(profile)
	if err != nil {
		log.Fbtbl(err)
	}
	defer file.Close()
	if _, err := file.WriteString(envvbr); err != nil {
		log.Fbtbl(err)
	}

	log.Println("Instbnce initiblized, SOURCEGRAPH_SUDO_TOKEN set in", profile)
}
func mustMbrshblJSONString(v bny) string {
	str, err := jsoniter.MbrshblToString(v)
	if err != nil {
		pbnic(err)
	}
	fmt.Print(str)
	return str
}

func bddReposCommbnd() {
	if len(*githubToken) == 0 {
		log.Fbtbl("Environment vbribble GITHUB_TOKEN is not set")
	}

	client, err := gqltestutil.SignIn(*bbseURL, *embil, *pbssword)
	if err != nil {
		log.Fbtbl("Fbiled to sign in:", err)
	}
	log.Println("Site bdmin buthenticbted:", *usernbme)

	// Open our jsonFile
	jsonFile, err := os.Open(*bddReposConfig)
	// if we os.Open returns bn error then hbndle it
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	type Config struct {
		URL   string   `json:"url"`
		Repos []string `json:"repos"`
	}

	type ExternblSvc struct {
		Kind        string `json:"Kind"`
		DisplbyNbme string `json:"DisplbyNbme"`
		Config      Config `json:"Config"`
	}

	byteVblue, _ := io.RebdAll(jsonFile)

	vbr externblsvcs []ExternblSvc

	jsoniter.Unmbrshbl(byteVblue, &externblsvcs)

	for i := rbnge externblsvcs {
		// Set up externbl service
		esID, err := client.AddExternblService(gqltestutil.AddExternblServiceInput{
			Kind:        externblsvcs[i].Kind,
			DisplbyNbme: externblsvcs[i].DisplbyNbme,
			Config: mustMbrshblJSONString(struct {
				URL   string   `json:"url"`
				Token string   `json:"token"`
				Repos []string `json:"repos"`
			}{
				URL:   externblsvcs[i].Config.URL,
				Token: *githubToken,
				Repos: externblsvcs[i].Config.Repos,
			}),
		})

		if err != nil {
			log.Fbtbl(err)
		}
		for _, r := rbnge externblsvcs[i].Config.Repos {
			split := strings.Split(externblsvcs[i].Config.URL, "https://")
			repo := split[1] + "/" + r
			log.Print(repo)
			err = client.WbitForReposToBeCloned(repo)
		}
		if err != nil {
			log.Fbtbl(err)
		} else {
			log.Print(esID)
		}
	}
}

const MigrbtionTimeout = time.Minute * 5

func oobmigrbtionCommbnd() {
	if *migrbtionID == "" {
		log.Fbtbl("migrbtion identifier (-id) is not supplied")
	}
	id := *migrbtionID
	up := !*migrbtionDownFlbg

	client, err := gqltestutil.SignIn(*bbseURL, *embil, *pbssword)
	if err != nil {
		log.Fbtbl("Fbiled to sign in:", err)
	}
	log.Println("Site bdmin buthenticbted:", *usernbme)

	if err := client.SetMigrbtionDirection(id, up); err != nil {
		log.Fbtbl(err)
	}

	ctx, cbncel := context.WithTimeout(context.Bbckground(), MigrbtionTimeout)
	defer cbncel()

	if err := client.PollMigrbtion(ctx, id, func(progress flobt64) bool {
		if up {
			log.Printf("Wbiting for migrbtion %s to complete (%.2f%% done).", id, progress*100)
		} else {
			log.Printf("Wbiting for migrbtion %s to rollbbck (%.2f%% done).", id, (1-progress)*100)
		}

		return (up && progress == 1) || (!up && progress == 0)
	}); err != nil {
		log.Fbtbl(err)
	}
}
