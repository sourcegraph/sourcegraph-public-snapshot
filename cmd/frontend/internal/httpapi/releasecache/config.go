package releasecache

import (
	"net/url"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// config represents an easier to work with version of the generated site config
// structs related to the version cache.
type config struct {
	enabled  bool
	interval time.Duration

	api   *url.URL
	owner string
	name  string
	urn   string

	token         string
	webhookSecret string
}

func parseSiteConfig(conf *conf.Unified) (*config, error) {
	// Set up our defaults, which should match the defaults in the site config
	// schema.
	config := &config{
		enabled:  false,
		interval: 1 * time.Hour,
		owner:    "sourcegraph",
		name:     "src-cli",
		urn:      "https://github.com",
	}

	// There's a _lot_ of pointer indirection boilerplate required to build up
	// the config, so feel free to have your eyes glaze over for the next 20
	// lines or so.
	dotCom := conf.Dotcom
	if dotCom == nil {
		return config, nil
	}

	c := dotCom.SrcCliVersionCache
	if c == nil || !c.Enabled {
		return config, nil
	}
	if c.Github.Token == "" {
		return nil, errors.New("no GitHub token provided")
	}
	if c.Github.WebhookSecret == "" {
		return nil, errors.New("no webhook secret provided")
	}

	config.enabled = true
	config.token = c.Github.Token
	config.webhookSecret = c.Github.WebhookSecret

	if c.Interval != "" {
		var err error
		if config.interval, err = time.ParseDuration(c.Interval); err != nil {
			return nil, errors.Wrapf(err, "parsing interval %s", c.Interval)
		}
	}

	// We need to turn the GitHub URI into the "urn" and "apiUrl" required to
	// build a GitHub V4Client.
	if c.Github.Uri != "" {
		config.urn = c.Github.Uri
	}
	url, err := url.Parse(config.urn)
	if err != nil {
		return nil, errors.Wrap(err, "parsing GitHub URL from configuration")
	}
	config.api, _ = github.APIRoot(url)

	if c.Github.Repository != nil {
		if c.Github.Repository.Owner != "" {
			config.owner = c.Github.Repository.Owner
		}
		if c.Github.Repository.Name != "" {
			config.name = c.Github.Repository.Name
		}
	}

	return config, nil
}

// NewReleaseCache builds a new VersionCache based on the current site config.
func (c *config) NewReleaseCache(logger log.Logger) ReleaseCache {
	client := github.NewV4Client(c.urn, c.api, &auth.OAuthBearerToken{Token: c.token}, nil)

	return newReleaseCache(logger, client, c.owner, c.name)
}
