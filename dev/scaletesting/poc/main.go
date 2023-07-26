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
	org    *github.Organization
	gh     *github.Client
}

func NewClient(ctx context.Context, cfg config) (*Client, error) {
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

	org, _, err := gh.Organizations.Get(ctx, cfg.GithubOrg)
	if err != nil {
		return nil, err
	}

	c := Client{
		gh:     gh,
		org:    org,
		config: &cfg,
	}

	return &c, err
}

func (c *Client) AddTeamMembership(ctx context.Context, user *github.User, team *github.Team) error {
	// is this user already part of the team?
	_, resp, err := c.gh.Teams.GetTeamMembershipByID(ctx, c.org.GetID(), team.GetID(), user.GetLogin())
	if resp.StatusCode == 200 {
		// user is already part of this team
		return nil
	} else if resp.StatusCode >= 500 {
		return fmt.Errorf("server error[%d]: %v", resp.StatusCode, err)
	}

	// user isn't part of the team so lets add them
	log.Printf("[INFO] Add user %q to team %s", user.GetLogin(), team.GetName())
	_, _, err = c.gh.Teams.AddTeamMembershipByID(ctx, c.org.GetID(), team.GetID(), user.GetLogin(), &github.TeamAddTeamMembershipOptions{
		Role: "member",
	})

	return err

}

func (c *Client) GetOrCreateTeam(ctx context.Context, newTeam *github.NewTeam) (*github.Team, error) {
	team, resp, err := c.gh.Teams.GetTeamBySlug(ctx, c.config.GithubOrg, newTeam.Name)
	switch resp.StatusCode {
	case 200:
		return team, nil
	case 404:
		team, _, err = c.gh.Teams.CreateTeam(ctx, c.config.GithubOrg, *newTeam)
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
	ctx := context.Background()
	cfg, err := Load("config.json")
	if err != nil {
		log.Fatalf("[ERR] failed to load config.json: %v\n", err)
	}

	c, err := NewClient(ctx, *cfg)
	if err != nil {
		log.Fatalf("[ERR] failed to create client: %v", err)
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
			log.Printf("[INFO] user match %q\n", name)
			templateUsers[key].User = u
		} else {
			log.Printf("[INFO] skip %q\n", name)
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
		team, err := c.GetOrCreateTeam(ctx, &github.NewTeam{
			Name:        t.Name,
			Description: &t.Description,
			Privacy:     strp("closed"),
		})
		if err != nil {
			log.Fatalf("[ERR] failed to get/create team %s: %v", t.Name, err)
		}

		for _, key := range t.MemberKeys {
			user := templateUsers[key]
			if err := c.AddTeamMembership(ctx, user.User, team); err != nil {
				log.Printf("[ERR] failed to add %q to team %v: %v", user.User.GetLogin(), team.GetName(), err)
				continue
			}
			templateUsers[key].Teams = append(templateUsers[key].Teams, team)
		}
	}

	for _, v := range templateUsers {
		fmt.Println(v.String())
	}
}
