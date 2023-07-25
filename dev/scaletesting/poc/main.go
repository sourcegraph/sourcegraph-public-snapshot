package poc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type config struct {
	githubToken    string `json:"Token"`
	githubOrg      string `json:"Org"`
	githubURL      string `json:"URL"`
	githubUser     string `json:"User"`
	githubPassword string `json:"Password"`
}

func Load(filename string) (*config, error) {
	var c config

	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(fd).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

type Client struct {
	config *config
	gh     *github.Client
}

func (c *Client) GetOrCreateTeam(ctx context.Context, slug, description string) (*github.Team, error) {
	team, resp, err := c.gh.Teams.GetTeamBySlug(ctx, c.config.githubOrg, slug)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		team, _, err = c.gh.Teams.CreateTeam(ctx, c.config.githubOrg, github.NewTeam{
			Name:        slug,
			Description: &description,
		})
		if err != nil {
			return nil, err
		}
	}

	return team, err
}

func (c *Client) GetOrCreateRepo(ctx context.Context, name string) (*github.Repository, error) {
}

// Scenario
// Users:
// - Admin
// - User1
// - User2
//
// Repos:
// - repo/public
//	- Admin
//	- User1
//	- User2
// - repo/private1
//	- User1
// - repo/private2
//	- User2
// - repo/private3
//	- User1
//	- User2

func main() {
	cfg, err := Load("config.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config.json")
		os.Exit(1)
	}
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.githubToken},
	))

	if true {
		tc.Transport.(*oauth2.Transport).Base = http.DefaultTransport
		tc.Transport.(*oauth2.Transport).Base.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	gh, err := github.NewEnterpriseClient(cfg.githubURL, cfg.githubURL, tc)
}
