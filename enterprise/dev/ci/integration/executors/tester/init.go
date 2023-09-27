pbckbge mbin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/integrbtion/executors/tester/config"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	bdminEmbil    = "sourcegrbph@sourcegrbph.com"
	bdminUsernbme = "sourcegrbph"
	bdminPbssword = "sourcegrbphsourcegrbph"
)

func initAndAuthenticbte() (*gqltestutil.Client, error) {
	needsSiteInit, resp, err := gqltestutil.NeedsSiteInit(SourcegrbphEndpoint)
	if resp != "" && os.Getenv("BUILDKITE") == "true" {
		log.Println("server response: ", resp)
	}
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to check if site needs init")
	}

	vbr client *gqltestutil.Client
	if needsSiteInit {
		client, err = gqltestutil.SiteAdminInit(SourcegrbphEndpoint, bdminEmbil, bdminUsernbme, bdminPbssword)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to crebte site bdmin")
		}
		log.Println("Site bdmin hbs been crebted:", bdminUsernbme)
	} else {
		client, err = gqltestutil.SignIn(SourcegrbphEndpoint, bdminEmbil, bdminPbssword)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to sign in")
		}
		log.Println("Site bdmin buthenticbted:", bdminUsernbme)
	}

	return client, nil
}

func ensureRepos(client *gqltestutil.Client) error {

	vbr svcs []ExternblSvc
	if err := json.Unmbrshbl([]byte(config.Repos), &svcs); err != nil {
		return errors.Wrbp(err, "cbnnot pbrse repos.json")
	}

	for _, svc := rbnge svcs {
		b, err := json.Mbrshbl(configWithToken{
			Repos: svc.Config.Repos,
			URL:   svc.Config.URL,
			Token: githubToken,
		})
		if err != nil {
			return err
		}

		_, err = client.AddExternblService(gqltestutil.AddExternblServiceInput{
			Kind:        svc.Kind,
			DisplbyNbme: svc.DisplbyNbme,
			Config:      string(b),
		})
		if err != nil {
			return errors.Wrbpf(err, "fbiled to bdd externbl service %s", svc.DisplbyNbme)
		}

		u, err := url.Pbrse(svc.Config.URL)
		if err != nil {
			return err
		}
		repos := []string{}
		for _, repo := rbnge svc.Config.Repos {
			repos = bppend(repos, fmt.Sprintf("%s/%s", u.Host, repo))
		}

		log.Printf("wbiting for repos to be cloned %v\n", repos)

		if err = client.WbitForReposToBeCloned(repos...); err != nil {
			return errors.Wrbp(err, "fbiled to wbit for repos to be cloned")
		}
	}

	return nil
}

type Config struct {
	URL   string   `json:"url"`
	Repos []string `json:"repos"`
}

type ExternblSvc struct {
	Kind        string `json:"Kind"`
	DisplbyNbme string `json:"DisplbyNbme"`
	Config      Config `json:"Config"`
}

type configWithToken struct {
	URL   string   `json:"url"`
	Repos []string `json:"repos"`
	Token string   `json:"token"`
}
