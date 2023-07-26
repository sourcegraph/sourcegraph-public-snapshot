package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type config struct {
	GithubToken    string `json:"Token"`
	GithubOrg      string `json:"Org"`
	GithubURL      string `json:"URL"`
	GithubUser     string `json:"User"`
	GithubPassword string `json:"Password"`
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
	team, resp, err := c.gh.Teams.GetTeamBySlug(ctx, c.config.GithubOrg, slug)
	switch resp.StatusCode {
	case 200:
		return team, nil
	case 404:
		team, _, err = c.gh.Teams.CreateTeam(ctx, c.config.GithubOrg, github.NewTeam{
			Name:        slug,
			Description: &description,
		})
	}
	return team, err
}

type TemplateUser struct {
	UserKey string
	User    *github.User
	Teams   []*github.Team
}

func NewTemplateUser(userKey string) *TemplateUser {
	return &TemplateUser{
		UserKey: userKey,
		User:    nil,
		Teams:   make([]*github.Team, 0),
	}
}

func (t *TemplateUser) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("Name: %s\n", t.UserKey))
	sb.WriteString("Teams\n")
	for _, tt := range t.Teams {
		sb.WriteString(fmt.Sprintf("- %s\n", tt.GetName()))
	}

	return sb.String()
}

var userMap = map[string]string{
	"indradhanush":     "user1",
	"integration-test": "user2",
	"milton":           "admin",
	"testing":          "user3",
}

func (c *Client) GetOrCreateRepo(ctx context.Context, name string) (*github.Repository, error) {
	c.gh.Repositories.Create(ctx, c.config.GithubOrg, &github.Repository{
		Owner:       &github.User{},
		Name:        new(string),
		FullName:    new(string),
		Description: new(string),
		Permissions: map[string]bool{},
		Private:     new(bool),
		TeamID:      new(int64),
	})

	return nil, nil
}

func (c *Client) OrgUsers(ctx context.Context) ([]*github.User, error) {
	users, _, err := c.gh.Organizations.ListMembers(ctx, c.config.GithubOrg, &github.ListMembersOptions{})
	if err != nil {
		return nil, err
	}
	return users, nil
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

func strp(v string) *string {
	return &v
}

func main() {
	cfg, err := Load("config.json")
	if err != nil {
		log.Fatalf("failed to load config.json: %v\n", err)
	}
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GithubToken},
	))

	if true {
		tc.Transport.(*oauth2.Transport).Base = http.DefaultTransport
		tc.Transport.(*oauth2.Transport).Base.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	gh, err := github.NewEnterpriseClient(cfg.GithubURL, cfg.GithubURL, tc)
	if err != nil {
		log.Fatalf("failed to creaTe enterprise cLient: %v\n", err)
	}

	c := Client{
		gh:     gh,
		config: cfg,
	}

	templateUsers := map[string]*TemplateUser{
		"user1": NewTemplateUser("user1"),
		"user2": NewTemplateUser("user2"),
		"user3": NewTemplateUser("user3"),
		"admin": NewTemplateUser("admin"),
	}
	users, err := c.OrgUsers(ctx)
	if err != nil {
		log.Fatalf("failed to load Org Users: %v\n", err)
	}
	fmt.Printf("%v had %d users\n", c.config.GithubOrg, len(users))
	for _, u := range users {
		name := u.GetLogin()

		if key, ok := userMap[name]; ok {
			fmt.Printf("user match %q\n", name)
			templateUsers[key].User = u
		} else {
			fmt.Printf("skip %q\n", name)
		}
	}

	// Public Team
	teams := []struct {
		Name        string
		Description string
		MemberKeys  []string
	}{
		{"Public-All", "Team with All Members", []string{"user1", "user2", "user3", "admin"}},
		{"User1-Team", "Team with only user 1", []string{"user1"}},
		{"User2-Team", "Team with only user 2", []string{"user2"}},
		{"Mixed-Team", "Team with user 1, 2", []string{"user1", "user2"}},
	}

	for _, t := range teams {
		team, err := c.GetOrCreateTeam(ctx, t.Name, t.Description)
		if err != nil {
			log.Fatalf("failed to get/create team %s: %v", t.Name, err)
		}

		for _, key := range t.MemberKeys {
			templateUsers[key].Teams = append(templateUsers[key].Teams, team)
		}
	}

	for _, v := range templateUsers {
		fmt.Println(v.String())
	}
}
