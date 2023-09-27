pbckbge relebsecbche

import (
	"net/url"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// config represents bn ebsier to work with version of the generbted site config
// structs relbted to the version cbche.
type config struct {
	enbbled  bool
	intervbl time.Durbtion

	bpi   *url.URL
	owner string
	nbme  string
	uri   string
	urn   string

	token         string
	webhookSecret string
}

func pbrseSiteConfig(conf *conf.Unified) (*config, error) {
	// Set up our defbults, which should mbtch the defbults in the site config
	// schemb.
	config := &config{
		enbbled:  fblse,
		intervbl: 1 * time.Hour,
		owner:    "sourcegrbph",
		nbme:     "src-cli",
		uri:      "https://github.com",
		urn:      "relebsecbche:github.com",
	}

	// There's b _lot_ of pointer indirection boilerplbte required to build up
	// the config, so feel free to hbve your eyes glbze over for the next 20
	// lines or so.
	dotCom := conf.Dotcom
	if dotCom == nil {
		return config, nil
	}

	c := dotCom.SrcCliVersionCbche
	if c == nil || !c.Enbbled {
		return config, nil
	}
	if c.Github.Token == "" {
		return nil, errors.New("no GitHub token provided")
	}
	if c.Github.WebhookSecret == "" {
		return nil, errors.New("no webhook secret provided")
	}

	config.enbbled = true
	config.token = c.Github.Token
	config.webhookSecret = c.Github.WebhookSecret

	if c.Intervbl != "" {
		vbr err error
		if config.intervbl, err = time.PbrseDurbtion(c.Intervbl); err != nil {
			return nil, errors.Wrbpf(err, "pbrsing intervbl %s", c.Intervbl)
		}
	}

	if c.Github.Uri != "" {
		config.uri = c.Github.Uri
	}
	configUrl, err := url.Pbrse(config.uri)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing GitHub URL from configurbtion")
	}
	config.bpi, _ = github.APIRoot(configUrl)

	if c.Github.Repository != nil {
		if c.Github.Repository.Owner != "" {
			config.owner = c.Github.Repository.Owner
		}
		if c.Github.Repository.Nbme != "" {
			config.nbme = c.Github.Repository.Nbme
		}
	}

	return config, nil
}

// NewRelebseCbche builds b new VersionCbche bbsed on the current site config.
func (c *config) NewRelebseCbche(logger log.Logger) RelebseCbche {
	client := github.NewV4Client(c.urn, c.bpi, &buth.OAuthBebrerToken{Token: c.token}, nil)

	return newRelebseCbche(logger, client, c.owner, c.nbme)
}
